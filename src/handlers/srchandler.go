/*
Copyright 2013 Google Inc. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/glog"
	"github.com/google/tsviewdb/src/cachinghandler"
	"github.com/google/tsviewdb/src/common"
	"github.com/google/tsviewdb/src/db"
	"github.com/google/tsviewdb/src/handlers/handlerutils"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	maxInputSize = 2 << 24 // 16MB max input payload size for PUT or POST.
)

type DBStruct struct {
	D db.DB
}

type SrcHandler DBStruct

func (this *SrcHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmaster := time.Now()
	switch r.Method { // Likely faster not to use a map[string]func.
	case "GET":
		if strings.Index(r.URL.Path, common.SrcsPath) != 0 {
			handlerutils.HttpError(w, "Bad path: "+r.URL.Path, http.StatusBadRequest)
			return
		}
		this.getHandler(w, r)
	case "POST":
		if strings.Index(r.URL.Path, common.SrcPath) != 0 {
			handlerutils.HttpError(w, "Bad path: "+r.URL.Path, http.StatusBadRequest)
			return
		}
		src := r.URL.Path[len(common.SrcPath):]
		this.postHandler(w, r, src)
	case "PUT":
		if strings.Index(r.URL.Path, common.SrcPath) != 0 {
			handlerutils.HttpError(w, "Bad path: "+r.URL.Path, http.StatusBadRequest)
			return
		}
		src := r.URL.Path[len(common.SrcPath):]
		this.putHandler(w, r, src)
	default:
		handlerutils.HttpError(w, "Bad method: "+r.Method, http.StatusBadRequest)
		return
	}

	glog.V(2).Infof("PERF: total service time: %v\n", time.Now().Sub(tmaster))
}

func getPayload(r *http.Request) ([]byte, error) {
	if size := r.ContentLength; size < 0 {
		return nil, errors.New("Unknown content size.")
	} else if size > maxInputSize {
		return nil, errors.New("Content too large.  Limit: " + string(maxInputSize))
	}
	return ioutil.ReadAll(r.Body)
}

func (this *SrcHandler) putHandler(w http.ResponseWriter, r *http.Request, src string) {
	glog.V(2).Infoln("src PUT handler")
	inputPayload, err := getPayload(r)
	if err != nil {
		handlerutils.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var sInfo db.SourceInfoUncomp
	if len(inputPayload) > 0 {
		err = json.Unmarshal(inputPayload, &sInfo)
		if err != nil {
			handlerutils.HttpError(w, "Malformed PUT data.", http.StatusBadRequest)
			return
		}
	}

	if err = this.D.WriteDir(sInfo, src); err != nil {
		handlerutils.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func (this *SrcHandler) postHandler(w http.ResponseWriter, r *http.Request, src string) {
	glog.V(2).Infoln("src POST handler")
	inputPayload, err := getPayload(r)
	if err != nil {
		handlerutils.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	var rec db.WriteRecord

	if len(inputPayload) > 0 {
		err = json.Unmarshal(inputPayload, &rec)
		if err != nil {
			handlerutils.HttpError(w, "Malformed POST data.", http.StatusBadRequest)
			return
		}
	}

	if rec.RecordTimestamp == nil {
		timestamp := time.Now().UnixNano() / 1e6 // Millis.
		rec.RecordTimestamp = &timestamp
	}

	rowId, err := this.D.WriteRow(rec, src)
	if err != nil {
		handlerutils.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}
	fmt.Fprintf(w, `{"id":"%s"}`, rowId)
}

func (this *SrcHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	glog.V(2).Infoln("src GET handler")
	rawQuery := r.URL.RawQuery
	q, _ := url.ParseQuery(rawQuery)
	t := q.Get("type")
	if t == "" {
		t = "json"
	}
	switch t { // Likely faster not to use a map[string]func.
	case "png":
		cachinghandler.HandleWithCache(w, r, "srcs-png", rawQuery)
	case "inline-graph":
		cachinghandler.HandleWithCache(w, r, "srcs-inline-graph", rawQuery)
	case "json":
		cachinghandler.HandleWithCache(w, r, "srcs-json", rawQuery)
	default:
		handlerutils.HttpError(w, "Bad srcs 'type' parameter: "+t, http.StatusBadRequest)
	}
}

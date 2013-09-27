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
	"bytes"
	"fmt"
	"github.com/golang/glog"
	"github.com/google/tsviewdb/src/handlers/handlerutils"
	"github.com/google/tsviewdb/src/handlers/templateloader"
	"net/http"
	"time"
)

type SearchHandler DBStruct

func (this *SearchHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmaster := time.Now()

	switch r.Method {
	case "GET":
		this.getHandler(w, r)
	default:
		handlerutils.HttpError(w, "Bad method: "+r.Method, http.StatusBadRequest)
		return
	}

	glog.V(2).Infof("PERF: total service time: %v\n", time.Now().Sub(tmaster))
}

func (this *SearchHandler) getHandler(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	searchStr := q.Get("q") + "*"

	sInfo, err := readDir(this.D, q, searchStr)
	if err != nil {
		handlerutils.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", 20))
	w.Header().Set("Content-Type", "text/html")

	var b bytes.Buffer
	tTemplate := time.Now()
	err = templateloader.Templates.ExecuteTemplate(&b, "search.template-html", struct {
		Title string
		Names []string
	}{
		Title: searchStr,
		Names: sInfo.Names,
	})
	glog.V(2).Infof("PERF: template generation time: %v\n", time.Now().Sub(tTemplate))

	contents := b.Bytes()
	if handlerutils.EtagMatch(w, r, contents) {
		return
	}
	w.Write(contents)
}

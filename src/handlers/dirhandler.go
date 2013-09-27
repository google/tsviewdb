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
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/google/tsviewdb/src/common"
	"github.com/google/tsviewdb/src/db"
	"github.com/google/tsviewdb/src/handlers/handlerutils"
	"net/http"
	"net/url"
	"time"
)

/////////////////////////////////////////////////////////////////////////////
// SEARCH AND DIRECTORY HANDLER

type DirHandler DBStruct

func (this *DirHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmaster := time.Now()

	searchPath := r.URL.Path[len(common.DirPath):]
	glog.V(2).Infoln("searchPath", searchPath)

	switch r.Method {
	case "GET":
		this.getHandler(w, r, searchPath)
	case "DELETE":
		this.deleteHandler(w, searchPath)
	default:
		handlerutils.HttpError(w, "Bad method: "+r.Method, http.StatusBadRequest)
		return
	}

	glog.V(2).Infof("PERF: total service time: %v\n", time.Now().Sub(tmaster))
}

func (this *DirHandler) getHandler(w http.ResponseWriter, r *http.Request, s string) {
	q := r.URL.Query()
	sInfo, err := readDir(this.D, q, s)
	if err != nil {
		handlerutils.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", 20))

	var b bytes.Buffer
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(&b).Encode(sInfo); err != nil {
		handlerutils.HttpError(w, "An error occured during JSON marshalling.", http.StatusInternalServerError)
		return
	}

	contents := b.Bytes()
	if handlerutils.EtagMatch(w, r, contents) {
		return
	}
	w.Write(contents)
}

func readDir(d db.DB, q url.Values, s string) (db.SourceInfoUncomp, error) {
	returnMetrics := q.Get("returnMetrics") == "1"
	returnUnits := q.Get("returnUnits") == "1"
	returnSelectForDefaults := q.Get("returnSelectForDefaults") == "1"
	defaultsOnly := q.Get("defaultsOnly") == "1"

	var prefixMatch bool
	if (len(s) > 0) && (s[len(s)-1:len(s)] == "*") {
		prefixMatch = true
		s = s[:len(s)-1]
	}

	dirSearchReq := db.DirectorySearchRequest{
		Prefix:                  s,
		FileRestrict:            "",
		ReturnMetrics:           returnMetrics,
		ReturnUnits:             returnUnits,
		ReturnSelectForDefaults: returnSelectForDefaults,
		DefaultsOnly:            defaultsOnly,
		DirPrefixMatch:          prefixMatch,
		FilePrefixMatch:         false}

	// Directory only search.
	sInfo, err := d.ReadDir(dirSearchReq)
	if err != nil {
		return db.SourceInfoUncomp{}, err
	}

	// File only search.
	if len(sInfo.Names) == 0 {
		path, file := common.GetSrcComponents(s)
		dirSearchReq.Prefix = path
		dirSearchReq.FileRestrict = file
		dirSearchReq.DirPrefixMatch = false
		dirSearchReq.FilePrefixMatch = prefixMatch
		sInfo, err = d.ReadDir(dirSearchReq)
		if err != nil {
			return db.SourceInfoUncomp{}, err
		}
	}

	return sInfo, nil
}

func (this *DirHandler) deleteHandler(w http.ResponseWriter, src string) {
	glog.V(2).Infoln("search DELETE handler")
	path, file := common.GetSrcComponents(src)
	glog.V(2).Infof("Deleting path: %s, file:%s", path, file)
	if err := this.D.DeleteDir(path, file); err != nil {
		handlerutils.HttpError(w, fmt.Sprintf("An error occured deleting path: %s, file: %s", path, file),
			http.StatusInternalServerError)
		return
	}
}

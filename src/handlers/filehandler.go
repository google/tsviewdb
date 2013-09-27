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
	"fmt"
	"github.com/golang/glog"
	"github.com/google/tsviewdb/src/handlers/handlerutils"
	"net/http"
	"strconv"
	"time"
)

type FileHandler struct {
	files            servedFiles
	startTimeRFC1123 string
	startTime        time.Time
}

func NewFileHandler(resourceDirectory string) *FileHandler {
	files, err := LoadFiles(resourceDirectory)
	if err != nil {
		glog.Fatalln(err.Error())
	}

	startTime := time.Now()
	startTimeRFC1123 := startTime.Format(time.RFC1123)
	return &FileHandler{
		files:            files,
		startTime:        startTime,
		startTimeRFC1123: startTimeRFC1123}
}

func (this *FileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmaster := time.Now()

	filePath := r.URL.Path[1:]
	glog.V(3).Infoln("filePath", filePath)

	switch r.Method {
	case "GET":
		this.getHandler(w, r, filePath)
	default:
		handlerutils.HttpError(w, "Bad method: "+r.Method, http.StatusBadRequest)
		return
	}

	glog.V(2).Infof("PERF: total service time: %v\n", time.Now().Sub(tmaster))
}

func (this *FileHandler) getHandler(w http.ResponseWriter, r *http.Request, filePath string) {
	switch {
	case filePath == "":
		filePath = "tsviewdb.html"
	case filePath == "v":
		filePath = "viz.html"
	}

	if this.notModifiedSince(w, r) {
		return
	}

	sFile, ok := this.files[filePath]
	if !ok {
		handlerutils.HttpError(w, "Resource not found: "+filePath, http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Length", strconv.Itoa(len(sFile.Contents)))
	w.Header().Set("Content-Type", sFile.ContentType)
	handlerutils.SetGzipContentHeader(w, r)
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", 3600*24*7))
	w.Header().Set("Last-Modified", this.startTimeRFC1123)
	w.Write(sFile.Contents)
}

func (this *FileHandler) notModifiedSince(w http.ResponseWriter, r *http.Request) bool {
	ifModifiedSince := r.Header.Get("If-Modified-Since")
	if ifModifiedSince != "" {
		// TODO: HTTP spec says need a difference of at least a minute.
		timeModified, err := time.Parse(time.RFC1123, ifModifiedSince)
		if err == nil && timeModified.Before(this.startTime) {
			w.WriteHeader(304)
			return true
		}
	}
	return false
}

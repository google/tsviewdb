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

package handlerutils

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/spaolacci/murmur3"
	"net/http"
	"strings"
	"time"
)

// HttpError is a wrapper around http.Error.  It resets the content-encoding
// field as well.
func HttpError(w http.ResponseWriter, error string, code int) {
	w.Header().Del("Content-Encoding")
	http.Error(w, error, code)
}

func EtagMatch(w http.ResponseWriter, r *http.Request, contents []byte) bool {
	tHash := time.Now()
	h1, h2 := murmur3.Sum128(contents)
	etag := fmt.Sprintf(`"%x%x"`, h1, h2)
	glog.V(3).Infof("PERF: murmurm3 generation time: %v\n", time.Now().Sub(tHash))

	ifNoneMatch := r.Header.Get("If-None-Match")
	if ifNoneMatch == etag {
		w.WriteHeader(304)
		return true
	}

	w.Header().Set("Etag", etag)
	return false
}

func SetGzipContentHeader(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
	}
}

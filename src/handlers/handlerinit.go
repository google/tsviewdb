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
	"flag"
	"github.com/google/tsviewdb/src/cachinghandler"
	"github.com/google/tsviewdb/src/common"
	"github.com/google/tsviewdb/src/db"
	"github.com/google/tsviewdb/src/gziphandler"
	"github.com/google/tsviewdb/src/handlers/templateloader"
	"github.com/google/tsviewdb/src/rangecontent"
	"net/http"
	"os"
	"path/filepath"
)

const (
	// Likely path from GOPATH to TSVIEWDBROOT.
	rootPath = "src/github.com/google/tsviewdb"
)

func getTSVIEWDBROOT() (tsviewdbRoot string) {
	tsviewdbRoot = os.Getenv("TSVIEWDBROOT")
	if tsviewdbRoot == "" {
		goPath := os.Getenv("GOPATH")
		if goPath != "" {
			return filepath.Join(goPath, rootPath)
		}
	}
	return
}

var resourceDir = flag.String("resourceDir",
	filepath.Join(getTSVIEWDBROOT(), "resources"),
	"Path to resource directory.")

func initCaches(d db.DB) {
	cachinghandler.Initialize()
	cachinghandler.RegisterCacheContentCreator(d, "srcs-json", rangecontent.MakeSrcsJsonContent,
		"application/json", true)
	cachinghandler.RegisterCacheContentCreator(d, "record-json", makeRecordJsonContent,
		"application/json", true)
	cachinghandler.RegisterCacheContentCreator(d, "srcs-inline-graph", rangecontent.MakeSrcsInlineGraphContent,
		"text/html; charset=UTF-8", true)
	cachinghandler.RegisterCacheContentCreator(d, "srcs-png", rangecontent.MakeSrcsPngContent,
		"image/png", false)
}

func InitializeAndRegister(d db.DB) {
	templateloader.LoadTemplates(*resourceDir)
	initCaches(d)

	srcHandler := &SrcHandler{D: d}
	http.Handle(common.SrcPath, srcHandler)
	http.Handle(common.SrcsPath, srcHandler)
	http.Handle(common.RecordPath, &RecordHandler{D: d})
	http.Handle(common.DirPath, gziphandler.NewGZipHandler(&DirHandler{D: d}))
	http.Handle(common.SearchPath, gziphandler.NewGZipHandler(&SearchHandler{D: d}))
	http.Handle("/", NewFileHandler(*resourceDir))
}

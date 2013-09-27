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
	"github.com/google/tsviewdb/src/gziphandler"
	"io/ioutil"
	"mime"
	"os"
	"path/filepath"
)

var (
	whiteListedPatterns = []string{"*.html", "*.css", "*.js", "*.ico"}
)

func getFileContents(filename string) (contents []byte, err error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return nil, err
	}

	contents, err = ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}
	return
}

type servedFile struct {
	Contents    []byte
	ContentType string
}

type servedFiles map[string]*servedFile

// loadFile is not safe for concurrent use!
func (s *servedFiles) loadFile(filename string) (err error) {
	uncompressedContents, err := getFileContents(filename)
	if err != nil {
		return err
	}
	baseName := filepath.Base(filename)

	contentType := mime.TypeByExtension(filepath.Ext(baseName))
	var contents bytes.Buffer
	if err = gziphandler.GzipContent(&contents, uncompressedContents); err != nil {
		return err
	}
	(*s)[baseName] = &servedFile{Contents: contents.Bytes(), ContentType: contentType}
	return
}

func LoadFiles(resourceDirectory string) (servedFiles, error) {
	files := servedFiles(make(map[string]*servedFile))
	var fileList []string
	for _, pat := range whiteListedPatterns {
		fList, err := filepath.Glob(filepath.Join(resourceDirectory, pat))
		if err != nil {
			return files, err
		}
		fileList = append(fileList, fList...)
	}

	mime.AddExtensionType(".ico", "image/x-icon")

	for _, fileName := range fileList {
		if err := files.loadFile(fileName); err != nil {
			return files, err
		}
	}

	return files, nil
}

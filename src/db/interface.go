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

package db

type DB interface {
	Init() (err error)
	WriteRow(wRecord WriteRecord, src string) (rowKey string, err error)
	ReadRow(req RowRequest) (returnVal *ReadRecord, err error)
	ReadRows(req RowRangeRequests) (returnVal *DataTable, err error)
	DeleteRow(rowKey string) (err error)
	WriteDir(si SourceInfoUncomp, src string) (err error)
	ReadDir(req DirectorySearchRequest) (result SourceInfoUncomp, err error)
	DeleteDir(path, file string) (err error)
}

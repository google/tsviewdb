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
	"github.com/google/tsviewdb/src/cachinghandler"
	"github.com/google/tsviewdb/src/common"
	"github.com/google/tsviewdb/src/db"
	"github.com/google/tsviewdb/src/db/requests"
	"github.com/google/tsviewdb/src/handlers/handlerutils"
	"net/http"
	"time"
)

type RecordHandler DBStruct

func (this *RecordHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tmaster := time.Now()

	id := r.URL.Path[len(common.RecordPath):]
	glog.V(2).Infoln("id", id)

	switch r.Method {
	case "GET":
		q := r.URL.Query()
		q.Add("id", id)
		newQS := q.Encode()
		this.getHandler(w, r, newQS)
	case "DELETE":
		this.deleteHandler(w, id)
	default:
		handlerutils.HttpError(w, "Bad method: "+r.Method, http.StatusBadRequest)
		return
	}

	glog.V(2).Infof("PERF: total service time: %v\n", time.Now().Sub(tmaster))
}

func (this *RecordHandler) getHandler(w http.ResponseWriter, r *http.Request, rawQuery string) {
	cachinghandler.HandleWithCache(w, r, "record-json", rawQuery)
}

func makeRecordJsonContent(d db.DB, b *bytes.Buffer, rawQuery string) (err error) {
	req, err := requests.MakeRowReq(rawQuery)
	if err != nil {
		return err
	}
	glog.V(3).Infoln("rowKey", req.Id)

	t2 := time.Now()
	returnValue, err := d.ReadRow(req)
	if err != nil {
		return err
	}
	glog.V(2).Infof("PERF: DB read time: %v\n", time.Now().Sub(t2))

	t3 := time.Now()
	if err := json.NewEncoder(b).Encode(returnValue); err != nil {
		return err
	}
	glog.V(2).Infof("PERF: JSON marshal time: %v\n", time.Now().Sub(t3))

	return nil
}

func (this *RecordHandler) deleteHandler(w http.ResponseWriter, id string) {
	glog.V(3).Infof("Deleting id: %s", id)
	if err := this.D.DeleteRow(id); err != nil {
		handlerutils.HttpError(w, fmt.Sprintf("An error occured deleting id: %s", id),
			http.StatusInternalServerError)
		return
	}
}

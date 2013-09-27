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

package rangecontent

import (
	"errors"
	"github.com/golang/glog"
	"github.com/google/tsviewdb/src/common"
	"github.com/google/tsviewdb/src/db"
	"github.com/google/tsviewdb/src/db/requests"
	"net/url"
	"time"
)

// getDataTableRaw returns a db.DataTable with the X-axis as time, all other
// columns sorted by name, and no specified row order.
func getDataTableRaw(D db.DB, req db.RowRangeRequests) (dTable *db.DataTable, err error) {
	t2 := time.Now()
	if len(req.FilteredSources) == 0 {
		return dTable, errors.New("No sources selected.")
	}
	dTable, err = D.ReadRows(req)
	if err != nil {
		return nil, err
	}

	t1 := time.Now()

	if len(dTable.ColumnNames) > 0 {
		originalName := dTable.ColumnNames[0]
		dTable.ColumnNames[0] = "!" + originalName // Force first column to sort first.
		dTable.SortDataColumns()
		dTable.ColumnNames[0] = originalName // Now remove "!"
		glog.V(2).Infof("PERF: Data column sort time: %v\n", time.Now().Sub(t1))
	}

	t2Delay := time.Now().Sub(t2)
	glog.V(2).Infof("PERF: DB read time: %v\n", t2Delay)
	if glog.V(2) && t2Delay.Seconds() > 0 && len(dTable.ColumnNames) > 0 {
		glog.Infof("PERF: DB row reads/sec: %d\n", int64(float64(len(dTable.Data))/t2Delay.Seconds()))
	}
	return dTable, nil
}

// getDataTable returns a db.Datatable with the X-axis as specified and rows
// sorted by the X-axis.
func getDataTable(D db.DB, rawQuery string) (dTable *db.DataTable, err error) {
	req, err := requests.MakeRowRangeReqs(rawQuery)
	if err != nil {
		return nil, err
	}
	dTable, err = getDataTableRaw(D, req)
	if err != nil {
		return nil, err
	}

	regressParams, err := requests.MakeRegressionParams(rawQuery)
	if err != nil {
		return nil, err
	}
	if regressParams.Selected {
		// Regression detection needs an ascending time sort, so perform this first
		// when the X-axis is already time.
		dTable.SortRows(0)
		dTable.GetVerifiedRegression(regressParams)
	}

	timeSort := req.SortByColumn == common.TimeName
	if !timeSort {
		if err = dTable.ChangeXAxisToColumnFromTime(req.SortByColumn); err != nil {
			return nil, err
		}
	}

	// Sort only if we haven't already (because we didn't enable regression
	// detection) or need to because we've changed the X-axis.
	if !regressParams.Selected || !timeSort {
		dTable.SortRows(0)
	}

	if req.SortByConfig != "" {
		err = dTable.ChangeXAxisToConfigColumn(req.SortByConfig,
			dTable.Timestamps == nil) // Timestamps not saved.
		if err != nil {
			return nil, err
		}
		// Use slower stable sort to allow 2-level sort, most typically of the form:
		// <configKey>:<Time>
		dTable.SortRowsStable(0)
	}

	q, _ := url.ParseQuery(rawQuery)
	reverse := q.Get("reverse") == "1"
	if reverse {
		dTable.ReverseRows()
	}

	if req.EqualX { // Only perform after all sorting is done.
		tAddColumn := time.Now()
		if dTable.Timestamps != nil { // Timestamps already saved.
			dTable.OverwriteXAxisWithRecordNum()
		} else {
			dTable.ChangeXAxisToRecordNumFromTime()
		}
		glog.V(3).Infof("PERF: add recordNum column time: %v\n", time.Now().Sub(tAddColumn))
	}
	return dTable, nil
}

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
	"bytes"
	"errors"
	"github.com/golang/glog"
	"github.com/google/tsviewdb/src/common"
	"github.com/google/tsviewdb/src/db"
	"github.com/google/tsviewdb/src/handlers/templateloader"
	"net/url"
	"strings"
	"time"
)

func MakeSrcsInlineGraphContent(d db.DB, b *bytes.Buffer, rawQuery string) error {
	dTable, err := getDataTable(d, rawQuery)
	if err != nil {
		return err
	}

	q, _ := url.ParseQuery(rawQuery)
	showShadow := q.Get("showShadow") == "1"

	aggregatesListStr := q.Get("aggregates")
	aggregatesMap := make(map[string]int) // Map from aggregate to index.
	if len(aggregatesListStr) > 0 {
		for idx, aggregate := range strings.Split(aggregatesListStr, ",") {
			aggregatesMap[aggregate] = idx
		}
	}

	if showShadow {
		if len(dTable.ColumnNames) < 4 || ((len(dTable.ColumnNames)-1)%3) != 0 {
			return errors.New("Need exactly 3 aggregates to show shadow plots.")
		}

		var data [][]interface{}
		var columnNames []string
		for i := 0; i < len(dTable.ColumnNames); i += 3 {
			metric, _ := common.GetMetricComponents(dTable.ColumnNames[i])
			columnNames = append(columnNames, metric)
		}

		// aggRemap is used to reorder the aggregates to reflect the query param order.
		aggRemap := make(map[int]int) // Remap actual index to target index.
		for i := 1; i < 4; i++ {
			_, aggregate := common.GetMetricComponents(dTable.ColumnNames[i])
			aggRemap[i-1] = aggregatesMap[aggregate]
		}

		for _, dRow := range dTable.Data {
			newRow := []interface{}{(*dRow)[0]} // Start row with X value.
			for i := 1; i < len(*dRow); i += 3 {
				v0 := (*dRow)[i+aggRemap[0]]
				v1 := (*dRow)[i+aggRemap[1]]
				v2 := (*dRow)[i+aggRemap[2]]
				newRow = append(newRow, []*float64{v0, v1, v2})
			}
			data = append(data, newRow)
		}
		tTemplate := time.Now()
		err = templateloader.Templates.ExecuteTemplate(b, "in-graph.template-html", struct {
			Data        [][]interface{}
			ColumnNames []string
			ShowShadow  bool
			XLabel      string
		}{
			Data:        data,
			ColumnNames: columnNames,
			ShowShadow:  true,
			XLabel:      dTable.ColumnNames[0],
		})
		glog.V(2).Infof("PERF: template generation time: %v\n", time.Now().Sub(tTemplate))
	} else {
		tTemplate := time.Now()
		err = templateloader.Templates.ExecuteTemplate(b, "in-graph.template-html", struct {
			Data        []*[]*float64
			ColumnNames []string
			ShowShadow  bool
			XLabel      string
		}{
			Data:        dTable.Data,
			ColumnNames: dTable.ColumnNames,
			ShowShadow:  false,
			XLabel:      dTable.ColumnNames[0],
		})
		glog.V(2).Infof("PERF: template generation time: %v\n", time.Now().Sub(tTemplate))
	}
	if err != nil {
		return err
	}

	return nil
}

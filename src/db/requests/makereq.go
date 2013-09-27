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

package requests

import (
	"errors"
	"flag"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"github.com/google/tsviewdb/src/common"
	"github.com/google/tsviewdb/src/db"
	"github.com/google/tsviewdb/src/regress"
	"github.com/google/tsviewdb/src/srcparse"
	"net/url"
	"strconv"
	"strings"
	"time"
)

var defaultResults = flag.Int("defaultResults", 25,
	"Default number of result records to return when no query string startDate set.")
var defaultMaxResults = flag.Int("defaultMaxResults", 100000,
	"Maximum number of result records to return when not specified with query string maxResults.")

const (
	millisPerDay = 3600 * 24 * 1000
)

func getEpochMillis(date string) int64 {
	t, err := time.Parse("20060102", date)
	if err != nil {
		glog.Errorln("An error occured during format conversion for", date)
		return 0
	}
	return t.Unix() * 1000
}

func MakeRowReq(rawQuery string) (db.RowRequest, error) {
	q, _ := url.ParseQuery(rawQuery)
	id := q.Get("id")
	noReturnAggregates := q.Get("noReturnAggregates") == "1"
	rowReq := db.RowRequest{
		Id:                 id,
		NoReturnAggregates: noReturnAggregates}
	return rowReq, nil
}

func MakeRowRangeReqs(rawQuery string) (db.RowRangeRequests, error) {
	q, _ := url.ParseQuery(rawQuery)

	srcs, _ := q["src"]

	var daysOfData int64
	daysOfDataStr := q.Get("daysOfData")
	if daysOfDataStr != "" {
		var err error
		daysOfData, err = strconv.ParseInt(daysOfDataStr, 10, 64)
		if err != nil {
			return db.RowRangeRequests{}, err
		}
	}

	var endTimestamp int64
	endDate := q.Get("endDate")
	if endDate == "" {
		endTimestamp = time.Now().Unix() * 1000
	} else {
		endTimestamp = getEpochMillis(endDate)
	}

	var startTimestamp int64
	startDate := q.Get("startDate")
	if startDate == "" {
		if daysOfData > 0 {
			startTimestamp = endTimestamp - millisPerDay*daysOfData
		}
	} else {
		startTimestamp = getEpochMillis(startDate)
	}

	var maxResults int
	maxResultStr := q.Get("maxResults")
	if maxResultStr == "" {
		if startDate != "" { // If specific startDate set, then return as much as possible.
			maxResults = *defaultMaxResults
		} else {
			maxResults = *defaultResults
		}
	} else {
		if _, err := fmt.Sscanf(maxResultStr, "%d", &maxResults); err != nil {
			return db.RowRangeRequests{}, errors.New("Bad input for maxResults parameter.")
		}
	}

	setAggregateIfMissing := q.Get("setAggregateIfMissing") == "1"

	aggregatesListStr := q.Get("aggregates")
	var aggregatesFilter map[string]bool
	if len(aggregatesListStr) > 0 {
		aggregatesFilter = make(map[string]bool)
		for _, aggregate := range strings.Split(aggregatesListStr, ",") {
			aggregatesFilter[aggregate] = true
		}
	}

	metricsFilterStr := q.Get("metrics")
	var metricsFilter map[string]bool
	if len(metricsFilterStr) > 0 {
		metricsFilter = make(map[string]bool)
		for _, metric := range strings.Split(metricsFilterStr, ",") {
			metricsFilter[metric] = true
		}
	}

	var configsFilter map[string]string
	if configsFilterList, ok := q["config"]; ok {
		configsFilter = make(map[string]string)
		for _, kvPair := range configsFilterList {
			kvArray := strings.SplitN(kvPair, "=", 2)
			k := kvArray[0]
			var v string
			if len(kvArray) > 1 {
				v = kvArray[1]
			}
			configsFilter[k] = v
		}
	}

	equalX := q.Get("equalX") == "1"
	sortByConfig := q.Get("sortByConfig")
	sortByColumn := q.Get("sortByColumn")
	if sortByColumn == "" {
		sortByColumn = common.TimeName
	}

	returnIds := q.Get("returnIds") == "1"
	returnConfigs := q.Get("returnConfigs") == "1"
	noReturnAggregates := q.Get("noReturnAggregates") == "1"

	// Now put together request struct.

	filteredSources := make([]db.FilteredSource, len(srcs))
	for i, s := range srcs {
		sr := srcparse.Parse(s)

		loopMetricsFilter := metricsFilter
		if sr.Metric != "" {
			loopMetricsFilter = map[string]bool{sr.Metric: true}
		}
		loopAggregatesFilter := aggregatesFilter
		if sr.Aggregate != "" {
			loopAggregatesFilter = map[string]bool{sr.Aggregate: true}
		}
		loopConfigsFilter := configsFilter
		if len(sr.Configs) != 0 {
			loopConfigsFilter = sr.Configs
		}

		filteredSources[i] = db.FilteredSource{
			Source:           sr.Source,
			MetricsFilter:    loopMetricsFilter,
			AggregatesFilter: loopAggregatesFilter,
			ConfigsFilter:    loopConfigsFilter}
	}

	qualifier := db.Qualifier{
		StartTimestamp:        startTimestamp,
		EndTimestamp:          endTimestamp,
		MaxResults:            maxResults,
		SetAggregateIfMissing: setAggregateIfMissing,
		EqualX:                equalX,
		SortByColumn:          sortByColumn,
		SortByConfig:          sortByConfig,
		ReturnIds:             returnIds,
		ReturnConfigs:         returnConfigs,
		NoReturnAggregates:    noReturnAggregates}

	req := db.RowRangeRequests{
		FilteredSources: filteredSources,
		Qualifier:       qualifier}

	if glog.V(4) {
		glog.Infoln("req", spew.Sdump(req))
	}

	return req, nil
}

func MakeRegressionParams(rawQuery string) (r regress.RegressionParams, err error) {
	q, _ := url.ParseQuery(rawQuery)

	regressPosStr := q.Get("regressPos")
	if regressPosStr != "" {
		var pos float64
		pos, err = strconv.ParseFloat(regressPosStr, 64)
		r.Pos = &pos
		if err != nil {
			return
		}
		r.Selected = true
	}

	regressNegStr := q.Get("regressNeg")
	if regressNegStr != "" {
		var neg float64
		neg, err = strconv.ParseFloat(regressNegStr, 64)
		r.Neg = &neg
		if err != nil {
			return
		}
		r.Selected = true
	}

	if !r.Selected {
		return
	}

	radiusStr := q.Get("regressRadius")
	if radiusStr != "" {
		r.Radius, err = strconv.Atoi(radiusStr)
		if err != nil {
			return
		}
	}

	windowStr := q.Get("regressWindow")
	r.Window = 1 // Default (0 is an error)
	if windowStr != "" {
		r.Window, err = strconv.Atoi(windowStr)
		if err != nil {
			return
		}
		if r.Window < 1 {
			return r, errors.New("regressWindow must be > 0")
		}
	}

	r.UsePercent = q.Get("regressUsePercent") == "1"

	ignoreLTStr := q.Get("regressIgnoreLT")
	if ignoreLTStr != "" {
		r.IgnoreLT, err = strconv.ParseFloat(ignoreLTStr, 64)
		if err != nil {
			return
		}
	}

	r.ReturnSegments = q.Get("regressReturnSegments") == "1"

	return
}

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

package cassandradb

import (
	"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/goprotobuf/proto"
	"errors"
	"github.com/adilhn/gossie/src/gossie"
	"github.com/golang/glog"
	"github.com/google/tsviewdb/src/common"
	"github.com/google/tsviewdb/src/db"
	"github.com/google/tsviewdb/src/db/dbcommon"
	pb "github.com/google/tsviewdb/src/proto"
	"strings"
	"time"
)

func (c *CassandraDB) WriteRow(wRecord db.WriteRecord, src string) (rowKey string, err error) {
	if len(wRecord.AggregatesColumnNames) != len(wRecord.Aggregates) {
		return "", errors.New("Aggregates names and data don't match.")
	}

	var timestamp int64
	if wRecord.RecordTimestamp != nil {
		timestamp = *wRecord.RecordTimestamp
	}
	uuidString := uuid.New()
	rowKey = dbcommon.MakeRowKey(src, timestamp, uuidString)

	////////////////////////////////////////////////////////////////////////////
	// Process points.

	ptsMap := make(map[string][]float64) // Used for creating missing aggregates in aggregate processing loop.
	// aggs is a map from metric name to map from aggregate to value.
	aggs := make(map[string]map[string]*float64)

	if len(wRecord.Points) > 0 {
		var ptsDataType pb.DataType
		dataTypeInt32, setPtsDataType := pb.DataType_value[strings.ToUpper(wRecord.PointsDataType)]
		if setPtsDataType { // Only set proto field if explicitly set by user.  Otherwise rely on default.
			ptsDataType = pb.DataType(dataTypeInt32)
		}

		ptsRow := &gossie.Row{[]byte(rowKey), nil}
		for _, pointRecord := range wRecord.Points {
			if (len(pointRecord.Timestamps) > 0) && (len(pointRecord.Data) != len(pointRecord.Timestamps)) {
				return "", errors.New("Points data and timestamps don't match for: " + pointRecord.Name)
			}

			p := &pb.Points{}
			if setPtsDataType { // Only set proto field if explicitly set by user.  Otherwise rely on default.
				dType := ptsDataType // Make new allocation for each one.
				p.Type = &dType
			}
			ptsMap[pointRecord.Name] = pointRecord.Data // Save for creating missing aggregates.
			// This blank map ensures we create any missing aggregates in the aggregate handling loops below.
			aggs[pointRecord.Name] = make(map[string]*float64)

			p.ValuesDouble = pointRecord.Data
			p.MakeDeltaValuesScaled(p.GetType())

			var previousTS int64
			for _, timestamp := range pointRecord.Timestamps {
				p.DeltaTimestamps = append(p.DeltaTimestamps, timestamp-previousTS)
				previousTS = timestamp
			}

			serializedData, e := proto.Marshal(p)
			if e != nil {
				return "", e
			}

			ptsRow.Columns = append(ptsRow.Columns, &gossie.Column{
				Name:  []byte(pointRecord.Name),
				Value: serializedData})
		}

		tPointsWrite := time.Now()
		if err := c.pool.Writer().Insert(dbcommon.CFPoints, ptsRow).Run(); err != nil {
			return "", err
		}
		glog.V(2).Infof("PERF: DB points write time: %v\n", time.Now().Sub(tPointsWrite))
	}

	////////////////////////////////////////////////////////////////////////////
	// Process aggregates (or make aggregates from points).

	if (len(wRecord.AggregatesColumnNames) + len(wRecord.Points)) > 0 {
		var aggDataType pb.DataType
		dataTypeInt32, setAggDataType := pb.DataType_value[strings.ToUpper(wRecord.AggregatesDataType)]
		if setAggDataType { // Only set proto field if explicitly set by user.  Otherwise rely on default.
			aggDataType = pb.DataType(dataTypeInt32)
		}

		// Parse aggregatesColumnNames and build aggs map of map.
		for idx, fullName := range wRecord.AggregatesColumnNames {
			metricName, aggregateName := common.GetMetricComponents(fullName)
			if metricName == "" {
				return "", errors.New("Missing metric name in:" + fullName)
			}
			if aggregateName == "" {
				return "", errors.New("Missing aggregate name in:" + fullName)
			}

			if aggs[metricName] == nil {
				aggs[metricName] = make(map[string]*float64)
			}
			value := wRecord.Aggregates[idx]
			aggs[metricName][aggregateName] = value
		}

		// Iterate over data in aggs map of map.
		aggRow := &gossie.Row{[]byte(rowKey), nil}
		for metricName, aggMap := range aggs {
			a := new(pb.Aggregation)
			if setAggDataType { // Only set proto field if explicitly set by user.  Otherwise rely on default.
				dType := aggDataType // Make new allocation for each one.
				a.Type = &dType
			}
			a.Double = &pb.Aggregation_AggregationDouble{}
			for aggregateName, valuePtr := range aggMap {
				a.SetDoubleField(aggregateName, valuePtr)
			}
			a.CreateMissingDoubleAggregates(ptsMap[metricName])

			a.MakeScaled(a.GetType())

			serializedData, e := proto.Marshal(a)
			if e != nil {
				return "", e
			}

			aggRow.Columns = append(aggRow.Columns, &gossie.Column{
				Name:  []byte(metricName),
				Value: serializedData,
			})
		}

		tAggregatesWrite := time.Now()
		if err := c.pool.Writer().Insert(dbcommon.CFAggregates, aggRow).Run(); err != nil {
			return "", err
		}
		glog.V(2).Infof("PERF: DB aggregates write time: %v\n", time.Now().Sub(tAggregatesWrite))
	}

	////////////////////////////////////////////////////////////////////////////
	// Process configs.

	if len(wRecord.ConfigPairs) > 0 {
		cfgRow := &gossie.Row{[]byte(rowKey), nil}
		for k, v := range wRecord.ConfigPairs {
			cfgRow.Columns = append(cfgRow.Columns, &gossie.Column{
				Name:  []byte(k),
				Value: []byte(v),
			})
		}
		if err := c.pool.Writer().Insert(dbcommon.CFConfigs, cfgRow).Run(); err != nil {
			return "", err
		}
	}

	////////////////////////////////////////////////////////////////////////////
	// Process src.

	if (len(wRecord.Points) + len(wRecord.AggregatesColumnNames) + len(wRecord.ConfigPairs)) > 0 {
		srcRow := &gossie.Row{[]byte(rowKey), nil}
		srcRow.Columns = append(srcRow.Columns, &gossie.Column{
			Name:  []byte(src),
			Value: []byte{},
		})
		return rowKey, c.pool.Writer().Insert(dbcommon.CFSource, srcRow).Run()
	} else {
		return "", errors.New("No data to write.")
	}
}

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
	"code.google.com/p/goprotobuf/proto"
	"errors"
	"github.com/golang/glog"
	"github.com/google/tsviewdb/src/common"
	"github.com/google/tsviewdb/src/db"
	"github.com/google/tsviewdb/src/db/dbcommon"
	pb "github.com/google/tsviewdb/src/proto"
	"sort"
	"strings"
	"time"
)

type dtablePtrErr struct {
	*db.DataTable
	err error
}

func (c *CassandraDB) ReadRows(req db.RowRangeRequests) (returnVal *db.DataTable, err error) {
	numTables := len(req.FilteredSources)
	dTables := make([]*db.DataTable, numTables)
	resultsChan := make(chan dtablePtrErr, numTables)
	// Start all requests in parallel.
	for i := range req.FilteredSources {
		go func(j int) {
			glog.V(3).Infoln("Starting req", j)
			var err error
			tmpDTable, err := c.readRowRange(req, j) // Read one of the sources.
			resultsChan <- dtablePtrErr{tmpDTable, err}
		}(i)
	}
	// Gather results.
	for i := 0; i < numTables; i++ {
		var err error
		dp := <-resultsChan
		dTables[i], err = dp.DataTable, dp.err
		if err != nil {
			return nil, err
		}
	}

	if len(dTables) == 1 { // Optimization for common case.
		return dTables[0], nil
	}

	// Merge tables.
	glog.V(3).Infoln("len(dTables)", len(dTables))
	var srcs []string
	for i := 0; i < numTables; i++ {
		srcs = append(srcs, req.FilteredSources[i].Source)
	}
	resultTable := db.MergeDataTables(dTables, srcs, req.ReturnIds, req.ReturnConfigs)
	return resultTable, nil
}

func (c *CassandraDB) readRowRange(req db.RowRangeRequests, reqNum int) (returnVal *db.DataTable, err error) {
	fs := req.FilteredSources[reqNum]
	src := fs.Source
	metricsFilter := fs.MetricsFilter
	aggregatesFilter := fs.AggregatesFilter
	configsFilter := fs.ConfigsFilter

	startPrefix, endPrefix := dbcommon.MakeRowPrefixes(src, req.StartTimestamp,
		req.EndTimestamp, true)

	glog.V(3).Infoln("startPrefix", startPrefix)
	glog.V(3).Infoln("endPrefix", endPrefix)

	// Start multiple column family requests in the background.
	var aggregationResultChan <-chan rowResults
	if !req.NoReturnAggregates {
		aggregationResultChan = getColumnFamilyRange(dbcommon.CFAggregates, c.pool, startPrefix, endPrefix, req.MaxResults)
	}
	var cfgResultChan <-chan rowResults
	if req.ReturnConfigs {
		cfgResultChan = getColumnFamilyRange(dbcommon.CFConfigs, c.pool, startPrefix, endPrefix, req.MaxResults)
	}

	dataTable := new(db.DataTable)

	/////////////////////////////////////////////////////////////////////////////
	// Read configs.

	var excludeIdSet map[string]bool
	if req.ReturnConfigs {
		if configsFilter != nil {
			excludeIdSet = make(map[string]bool)
		}
		cfgResult := <-cfgResultChan
		if cfgResult.err != nil {
			return nil, cfgResult.err
		}
		cfgRows := cfgResult.Rows
		glog.V(3).Infoln("len(cfgRows)", len(cfgRows))

		// Map from name to data slot to write data in data row.
		columnNameReverseMap := make(map[string]int)

		for _, cfgRow := range cfgRows {
			if cfgRow == nil {
				continue
			}
			// Created at least as much space as we know we'll use.  For data that
			// contains the same config names for every record (typical) this space
			// allocation will not change after the first row read.
			ctrow := make([]*string, len(dataTable.ConfigsColumnNames))

			var rowMatch bool // Used only when configsFilter is set.
			for _, column := range cfgRow.Columns {
				columnName := string(column.Name)
				valueStr := string(column.Value)
				if (configsFilter != nil) && (configsFilter[columnName] == valueStr) {
					rowMatch = true
				}
				if columnNameIndex, ok := columnNameReverseMap[columnName]; !ok { // Which slot to write data.
					columnNameReverseMap[columnName] = len(dataTable.ConfigsColumnNames)
					dataTable.ConfigsColumnNames = append(dataTable.ConfigsColumnNames, columnName)
					ctrow = append(ctrow, &valueStr)
				} else {
					ctrow[columnNameIndex] = &valueStr
				}
			}

			if (configsFilter != nil) && !rowMatch { // If not match, dump row and continue.
				excludeIdSet[string(cfgRow.Key)] = true // Mark row as excluded for aggregates.
				ctrow = nil                             // Mark as garbage
				continue
			}

			dataTable.Configs = append(dataTable.Configs, &ctrow)

			if req.NoReturnAggregates && req.ReturnIds {
				dataTable.IdColumn = append(dataTable.IdColumn, string(cfgRow.Key))
			}

		}
		t2 := time.Now()
		dataTable.SortConfigsColumns()
		glog.V(2).Infof("PERF: Config sort time: %v\n", time.Now().Sub(t2))
	}

	/////////////////////////////////////////////////////////////////////////////
	// Read aggregates.

	if !req.NoReturnAggregates {
		aggregationResult := <-aggregationResultChan
		if aggregationResult.err != nil {
			return nil, aggregationResult.err
		}
		aggregateRows := aggregationResult.Rows
		glog.V(3).Infoln("len(aggregateRows) = ", len(aggregateRows))

		var totalAggregationTime time.Duration

		// Map from name to data slot to write data in data row.
		columnNameReverseMap := make(map[string]int)

		dataTable.ColumnNames = append(dataTable.ColumnNames, common.TimeName) // First column.

		for _, aggregatesRow := range aggregateRows {
			if aggregatesRow == nil {
				continue
			}
			if (configsFilter != nil) && excludeIdSet[string(aggregatesRow.Key)] {
				continue
			}
			// Created at least as much space as we know we'll use.  For data that
			// contains the same metrics for every record (typical) this space
			// allocation will not change after the first row read.
			dtrow := make([]*float64, len(dataTable.ColumnNames))

			dataTable.Data = append(dataTable.Data, &dtrow)
			dtrow[0] = proto.Float64(float64(dbcommon.GetTimestamp(aggregatesRow.Key)))
			if req.ReturnIds {
				dataTable.IdColumn = append(dataTable.IdColumn, string(aggregatesRow.Key))
			}

			for _, column := range aggregatesRow.Columns {
				if (metricsFilter != nil) && !metricsFilter[string(column.Name)] {
					continue
				}

				aggregation := new(pb.Aggregation)

				if err := proto.Unmarshal(column.Value, aggregation); err != nil {
					return nil, errors.New("An error occured during aggregation unmarshalling.")
				}

				t0 := time.Now()
				aggregation.MakeDouble()

				fields, values := pb.GetDoubleFieldsAndValuesFiltered(aggregation,
					aggregatesFilter, req.SetAggregateIfMissing)
				totalAggregationTime += time.Now().Sub(t0)

				for fieldIndex, field := range fields {
					columnName := strings.Join([]string{string(column.Name), field}, ".")
					val := values[fieldIndex]
					if columnNameIndex, ok := columnNameReverseMap[columnName]; !ok { // Which slot to write data.
						columnNameReverseMap[columnName] = len(dataTable.ColumnNames)
						dataTable.ColumnNames = append(dataTable.ColumnNames, columnName)
						dtrow = append(dtrow, val)
					} else {
						dtrow[columnNameIndex] = val
					}
				}
			}

		}
		if len(dataTable.ColumnNames) == 1 {
			return nil, errors.New("No results for: " + req.FilteredSources[reqNum].Source)
		}

		glog.V(3).Infof("PERF: accumulated aggregate unpacking time: %v\n", totalAggregationTime)
	}

	return dataTable, nil
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

func (c *CassandraDB) ReadRow(req db.RowRequest) (returnVal *db.ReadRecord, err error) {
	// Start multiple requests in the background.
	var aggResultChan <-chan rowResult
	if !req.NoReturnAggregates {
		aggResultChan = getColumnFamily(dbcommon.CFAggregates, c.pool, req.Id)
	}
	pointsResultChan := getColumnFamily(dbcommon.CFPoints, c.pool, req.Id)
	srcResultChan := getColumnFamily(dbcommon.CFSource, c.pool, req.Id)
	cfgResultChan := getColumnFamily(dbcommon.CFConfigs, c.pool, req.Id)

	////////////////////////////////////////////////////////////////////////////
	// Read Points.

	pointsResult := <-pointsResultChan
	if pointsResult.err != nil {
		return nil, errors.New("An error occured reading points data.")
	}
	pointsRow := pointsResult.Row
	readRecord := &db.ReadRecord{Points: make([]*[]*float64, 0)}

	// Mapping from time to points data row.  Note that we don't need a
	// columnNameReverseMap as we do when reading rows to determine which slot to
	// write the actual data because we are reading only one row which has a fixed
	// set of column names.
	pointsMap := make(map[int64]*[]*float64)

	if pointsRow != nil {
		readRecord.RecordTimestamp = proto.Int64(dbcommon.GetTimestamp(pointsRow.Key))

		// Add time column with prepended "!" to force it to sort first.  We remove
		// the "!" after we're all done.
		readRecord.PointsColumnNames = append(readRecord.PointsColumnNames, "!"+common.TimeName)

		var checkedPointsTypeAlready bool
		for colIdx, col := range pointsRow.Columns {
			p := &pb.Points{}
			if err := proto.Unmarshal(col.Value, p); err != nil {
				return nil, errors.New("An error occured during points unmarshalling.")
			}

			if !checkedPointsTypeAlready {
				checkedPointsTypeAlready = true
				pointsDataType := p.Type
				if pointsDataType != nil {
					readRecord.PointsDataType = pointsDataType.String()
				}
			}

			p.MakeValuesDouble()

			if (len(p.DeltaTimestamps) > 0) && (len(p.DeltaTimestamps) == len(p.ValuesDouble)) {
				var previousTS int64
				for dataIdx, deltaTS := range p.DeltaTimestamps {
					timestamp := deltaTS + previousTS
					previousTS = timestamp
					dataRow, ok := pointsMap[timestamp]
					if !ok {
						newDataRow := make([]*float64, len(pointsRow.Columns)+1) // Space for timestamp
						tsVal := float64(timestamp)                              // Make copy.
						newDataRow[0] = &tsVal
						pointsMap[timestamp] = &newDataRow
						dataRow = &newDataRow
					}
					(*dataRow)[colIdx+1] = &p.ValuesDouble[dataIdx]
				}
			} else { // Either no timestamps or timestamps and data were different lengths.
				for timestamp, val := range p.ValuesDouble { // Use monotonic increasing timestamp.
					dataRow, ok := pointsMap[int64(timestamp)]
					if !ok {
						newDataRow := make([]*float64, len(pointsRow.Columns)+1) // Space for timestamp
						tsVal := float64(timestamp)                              // Make copy.
						newDataRow[0] = &tsVal
						pointsMap[int64(timestamp)] = &newDataRow
						dataRow = &newDataRow
					}
					floatVal := val // Make copy.
					(*dataRow)[colIdx+1] = &floatVal
				}
			}

			readRecord.PointsColumnNames = append(readRecord.PointsColumnNames, string(col.Name))
		}

		// Attach rows to readRecord.Points in order:
		// First obtain sorted list of keys.
		var allKeys []int64
		for k := range pointsMap {
			allKeys = append(allKeys, k)
		}
		sort.Sort(common.Int64Slice(allKeys))

		// Now attach rows.
		for _, k := range allKeys {
			readRecord.Points = append(readRecord.Points, pointsMap[k])
		}

		readRecord.SortPoints()
		readRecord.PointsColumnNames[0] = common.TimeName
	}

	////////////////////////////////////////////////////////////////////////////
	// Read Src.

	srcResult := <-srcResultChan
	if srcResult.err != nil {
		//		return nil, errors.New("An error occured reading src data.")
		return nil, srcResult.err
	}
	srcRow := srcResult.Row

	if srcRow != nil {
		if len(srcRow.Columns) > 0 {
			sourceStr := string(srcRow.Columns[0].Name)
			readRecord.Source = &sourceStr
		}
	}

	////////////////////////////////////////////////////////////////////////////
	// Read Aggregates.
	if !req.NoReturnAggregates {
		aggResult := <-aggResultChan
		if aggResult.err != nil {
			return nil, errors.New("An error occured reading aggregate data.")
		}
		aggRow := aggResult.Row

		if aggRow != nil {
			if readRecord.RecordTimestamp == nil {
				readRecord.RecordTimestamp = proto.Int64(dbcommon.GetTimestamp(aggRow.Key))
			}

			var checkedAggregatesTypeAlready bool
			for _, column := range aggRow.Columns {
				aggregation := new(pb.Aggregation)
				if err := proto.Unmarshal(column.Value, aggregation); err != nil {
					return nil, errors.New("An error occured during aggregation unmarshalling.")
				}
				aggregation.MakeDouble()

				if !checkedAggregatesTypeAlready { // We set all for a record to the same.
					checkedAggregatesTypeAlready = true
					aggregatesDataType := aggregation.Type
					if aggregatesDataType != nil {
						readRecord.AggregatesDataType = aggregatesDataType.String()
					}
				}

				fields, values := pb.GetDoubleFieldsAndValues(aggregation)

				for fieldIndex, field := range fields {
					columnName := strings.Join([]string{string(column.Name), field}, ".")
					readRecord.AggregatesColumnNames = append(readRecord.AggregatesColumnNames, columnName)
					readRecord.Aggregates = append(readRecord.Aggregates, values[fieldIndex])
				}
			}
			readRecord.SortAggregates()
		}
	}

	////////////////////////////////////////////////////////////////////////////
	// Read Configs.

	cfgResult := <-cfgResultChan
	if cfgResult.err != nil {
		return nil, errors.New("An error occured reading config data.")
	}
	cfgRow := cfgResult.Row

	if cfgRow != nil {
		if readRecord.RecordTimestamp == nil {
			readRecord.RecordTimestamp = proto.Int64(dbcommon.GetTimestamp(cfgRow.Key))
		}

		readRecord.ConfigPairs = make(map[string]string)
		for _, column := range cfgRow.Columns {
			readRecord.ConfigPairs[string(column.Name)] = string(column.Value)
		}
	}

	return readRecord, nil
}

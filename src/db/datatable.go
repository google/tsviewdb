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

import (
	"github.com/google/tsviewdb/src/common"
	"github.com/google/tsviewdb/src/regress"
	"sort"
	"strconv"
	"strings"
)

// src:metric.aggregate$key1=value1$key2=value2 decomposes into below.
type FilteredSource struct {
	Source           string
	MetricsFilter    map[string]bool
	AggregatesFilter map[string]bool
	ConfigsFilter    map[string]string // Setting key only will separate into different configs.
}

type Qualifier struct {
	StartTimestamp int64
	EndTimestamp   int64
	MaxResults     int

	SetAggregateIfMissing bool

	EqualX       bool
	SortByColumn string
	SortByConfig string // Overrides SortByColumn if both set.
	NoTimeColumn bool

	ReturnIds          bool
	ReturnConfigs      bool
	NoReturnAggregates bool
}

type RowRangeRequests struct {
	FilteredSources []FilteredSource
	Qualifier
}

type DataTable struct {
	ColumnNames        []string      `json:"aggregatesColumnNames,omitempty"`
	Data               []*[]*float64 `json:"aggregates,omitempty"`
	IdColumn           []string      `json:"ids,omitempty"`
	ConfigsColumnNames []string      `json:"configsColumnNames,omitempty"`
	Configs            []*[]*string  `json:"configs,omitempty"`
	Timestamps         []*float64    `json:"timestamps,omitempty"`
}

func (d *DataTable) SortDataColumns() {
	p := parallelStringsFloatTable{names: &d.ColumnNames, data: d.Data}
	p.FixRowLengths()
	p.SortDataColumns()
}

func (d *DataTable) OverwriteXAxisWithRecordNum() {
	d.ColumnNames[0] = common.RecordNumName
	for i, _ := range d.Data {
		fi := float64(i)
		(*d.Data[i])[0] = &fi
	}
}

func (d *DataTable) ChangeXAxisToRecordNumFromTime() {
	d.Timestamps = make([]*float64, len(d.Data))
	d.ColumnNames[0] = common.RecordNumName
	for i, _ := range d.Data {
		fi := float64(i)
		d.Timestamps[i] = (*d.Data[i])[0]
		(*d.Data[i])[0] = &fi
	}
}

func (d *DataTable) ChangeXAxisToColumnFromTime(colName string) error {
	colIdx, err := d.IndexForName(colName)
	if err != nil {
		return err
	}
	p := parallelStringsFloatTable{names: &d.ColumnNames, data: d.Data}
	d.Timestamps = make([]*float64, len(d.Data))
	p.copyColumnToX(colIdx, colName, &d.Timestamps)
	return nil
}

func (d *DataTable) ChangeXAxisToConfigColumn(colName string, fromTime bool) error {
	colIdx, err := d.IndexForConfigName(colName)
	if err != nil {
		return err
	}
	d.ColumnNames[0] = colName
	if fromTime {
		d.Timestamps = make([]*float64, len(d.Data))
	}

	for i, _ := range d.Data {
		if fromTime {
			d.Timestamps[i] = (*d.Data[i])[0] // Copy timestamps out.
		}
		configsRow := d.Configs[i]
		if configsRow == nil {
			continue
		}
		sPtr := (*configsRow)[colIdx]
		if sPtr == nil {
			continue
		}
		f, _ := strconv.ParseFloat(*sPtr, 64) // Swallow errors.
		(*d.Data[i])[0] = &f
	}
	return nil
}

func (d *DataTable) DeleteColumn(colName string) error {
	p := parallelStringsFloatTable{names: &d.ColumnNames, data: d.Data}
	return p.deleteColumn(colName)
}

func isNonData(name string) bool {
	return (name == common.TimeName) || (name == common.RecordNumName) ||
		(strings.Index(name, common.RegressNamePrefix) == 0)
}

func (d *DataTable) GetVerifiedRegression(rParams regress.RegressionParams) {
	p := parallelStringsFloatTable{names: &d.ColumnNames, data: d.Data}

	t := regress.NewTable(d.Data)
	for i := range d.ColumnNames {
		if isNonData(d.ColumnNames[i]) { // Don't compute regressions over known non-data columns.
			continue
		}
		result := t.GetVerifiedRegression(i, rParams)
		if result != nil {
			p.addColumn(common.RegressNamePrefix+d.ColumnNames[i], result)
		}
	}
}

var haveStableSort bool

func (d *DataTable) baseSortRows(idx int, reverse bool, stable bool) {
	ds := tableSort{
		v: d.Data,
		extraSwap: func(i, j int) {
			if len(d.IdColumn) == len(d.Data) {
				d.IdColumn[i], d.IdColumn[j] = d.IdColumn[j], d.IdColumn[i]
			}
			if len(d.Configs) == len(d.Data) {
				d.Configs[i], d.Configs[j] = d.Configs[j], d.Configs[i]
			}
			if len(d.Timestamps) == len(d.Data) {
				d.Timestamps[i], d.Timestamps[j] = d.Timestamps[j], d.Timestamps[i]
			}
		},
		sortIdx: 0,
	}
	var sortFunc func(sort.Interface)
	if stable {
		// TODO: uncomment once go1.2 is widely available.
		// sortFunc = sort.Stable
		sortFunc = sort.Sort
	} else {
		sortFunc = sort.Sort
	}

	if reverse {
		sortFunc(sort.Reverse(ds))
	} else {
		sortFunc(ds)
	}
}

func (d *DataTable) ReverseRows() {
	ds := tableSort{
		v: d.Data,
		extraSwap: func(i, j int) {
			if len(d.IdColumn) == len(d.Data) {
				d.IdColumn[i], d.IdColumn[j] = d.IdColumn[j], d.IdColumn[i]
			}
			if len(d.Configs) == len(d.Data) {
				d.Configs[i], d.Configs[j] = d.Configs[j], d.Configs[i]
			}
		},
	}
	ds.ReverseRows()
}

func (d *DataTable) IndexForName(colName string) (int, error) {
	return StringSlice(d.ColumnNames).IndexForName(colName)
}

func (d *DataTable) IndexForConfigName(colName string) (int, error) {
	return StringSlice(d.ConfigsColumnNames).IndexForName(colName)
}

func (d *DataTable) SortRows(colIdx int) {
	d.baseSortRows(colIdx, false, false)
}

func (d *DataTable) ReverseSortRows(colIdx int) {
	d.baseSortRows(colIdx, true, false)
}

func (d *DataTable) SortRowsStable(colIdx int) {
	d.baseSortRows(colIdx, false, true)
}

func (d *DataTable) ReverseSortRowsStable(colIdx int) {
	d.baseSortRows(colIdx, true, true)
}

func (d *DataTable) SortConfigsColumns() {
	p := parallelStringsStringTable{names: d.ConfigsColumnNames, data: d.Configs}
	p.FixRowLengths()
	p.SortDataColumns()
}

type fullRow struct {
	data    *[]*float64
	configs *[]*string
	id      string
}

// MergeDataTables returns a single DataTable from multiple ones.  Rows and
// columns are not in any defined order.  To sort columns use: SortDataColumns()
// To sort rows use: SortRows(colNum) or ReverseSortRows(colNum)
func MergeDataTables(dTables []*DataTable, srcs []string, returnIds,
	returnConfigs bool) (resultTable *DataTable) {
	resultTable = &DataTable{}

	rowMap := make(map[float64]*fullRow) // Map from X value to row.

	// Map from name to data slot to write data in data row.
	columnNameReverseMap := make(map[string]int)
	configColumnNameReverseMap := make(map[string]int)

	for i, dTable := range dTables {
		src := srcs[i]
		for rowIdx, rowp := range dTable.Data {
			if rowp == nil {
				continue
			}
			row := *rowp
			xp := row[0]
			if xp == nil {
				continue
			}
			xVal := *xp

			dtrow, ok := rowMap[xVal]
			if !ok {
				// Created at least as much space as we know we'll use.  Will be
				// increased through the append function for following tables.
				newDTRow := &fullRow{}
				data := make([]*float64, len(resultTable.ColumnNames))
				newDTRow.data = &data
				configs := make([]*string, len(resultTable.ConfigsColumnNames))
				newDTRow.configs = &configs
				dtrow = newDTRow
				rowMap[xVal] = dtrow
			}

			// Handle data.
			for j, colName := range dTable.ColumnNames {
				var columnName string
				if colName == common.TimeName {
					if (len(*(*dtrow).data) > 0) && ((*(*dtrow).data)[0] != nil) { // Don't write X column more than once.
						continue
					}
					columnName = colName
				} else {
					columnName = strings.Join([]string{src, colName}, ":")
				}

				val := row[j]
				if columnNameIndex, ok := columnNameReverseMap[columnName]; !ok { // Which slot to write data.
					columnNameReverseMap[columnName] = len(resultTable.ColumnNames)
					resultTable.ColumnNames = append(resultTable.ColumnNames, columnName)
					*(*dtrow).data = append(*(*dtrow).data, val)
				} else {
					(*dtrow.data)[columnNameIndex] = val
				}
			}

			// Handle Ids.
			if returnIds {
				(*dtrow).id = dTable.IdColumn[rowIdx]
			}

			// Handle configs.
			if returnConfigs {
				configRow := dTable.Configs[rowIdx]
				if configRow == nil {
					continue
				}
				for j, columnName := range dTable.ConfigsColumnNames {
					val := (*configRow)[j]
					if columnNameIndex, ok := configColumnNameReverseMap[columnName]; !ok { // Which slot to write data.
						configColumnNameReverseMap[columnName] = len(resultTable.ConfigsColumnNames)
						resultTable.ConfigsColumnNames = append(resultTable.ConfigsColumnNames, columnName)
						*(*dtrow).configs = append(*(*dtrow).configs, val)
					} else {
						(*dtrow.configs)[columnNameIndex] = val
					}
				}
			}

		} // end row processing
	} // end table processing

	for _, row := range rowMap {
		resultTable.Data = append(resultTable.Data, row.data)
		if returnIds {
			resultTable.IdColumn = append(resultTable.IdColumn, row.id)
		}
		if returnConfigs {
			resultTable.Configs = append(resultTable.Configs, row.configs)
		}
	}

	return
}

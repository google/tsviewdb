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
	"sort"
)

type RowRequest struct {
	Id                 string
	NoReturnAggregates bool
}

type PointsRecord struct {
	Name       string    `json:"name,omitempty"`
	Timestamps []int64   `json:"timestamps,omitempty"`
	Data       []float64 `json:"data,omitempty"`
}

type WriteRecord struct {
	RecordTimestamp       *int64            `json:"recordTimestamp,omitempty"`
	Points                []PointsRecord    `json:"points,omitempty"`
	PointsDataType        string            `json:"pointsDataType,omitempty"`
	AggregatesColumnNames []string          `json:"aggregatesColumnNames,omitempty"`
	Aggregates            []*float64        `json:"aggregates,omitempty"`
	AggregatesDataType    string            `json:"aggregatesDataType,omitempty"`
	ConfigPairs           map[string]string `json:"configPairs,omitempty"`
}

type ReadRecord struct {
	Source                *string           `json:"source,omitempty"`
	RecordTimestamp       *int64            `json:"recordTimestamp,omitempty"`
	PointsColumnNames     []string          `json:"pointsColumnNames,omitempty"`
	Points                []*[]*float64     `json:"points,omitempty"`
	PointsDataType        string            `json:"pointsDataType,omitempty"`
	AggregatesColumnNames []string          `json:"aggregatesColumnNames,omitempty"`
	Aggregates            []*float64        `json:"aggregates,omitempty"`
	AggregatesDataType    string            `json:"aggregatesDataType,omitempty"`
	ConfigPairs           map[string]string `json:"configPairs,omitempty"`
	Id                    string            `json:"id,omitempty"`
}

func (r *ReadRecord) SortPoints() {
	p := parallelStringsFloatTable{names: &r.PointsColumnNames, data: r.Points}
	p.SortDataColumns()
}

func (r *ReadRecord) SortAggregates() {
	psf := parallelStringsFloats{names: r.AggregatesColumnNames, floats: r.Aggregates}
	sort.Sort(&psf)
}

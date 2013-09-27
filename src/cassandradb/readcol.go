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
	"github.com/adilhn/gossie/src/gossie"
	"github.com/golang/glog"
)

// Read helper functions.

type rowResult struct {
	Row *gossie.Row
	err error
}

// getColumnFamily starts a read in a goroutine and returns a channel.
func getColumnFamily(cf string, pool gossie.ConnectionPool, rowKey string) <-chan rowResult {
	outChan := make(chan rowResult)

	go func() {
		glog.V(3).Infoln("Reading data for " + cf + " ...")
		row, err := pool.Reader().Cf(cf).Get([]byte(rowKey))
		if err != nil {
			outChan <- rowResult{nil, err}
		}
		outChan <- rowResult{row, nil}
	}()

	return outChan
}

type rowResults struct {
	Rows []*gossie.Row
	err  error
}

// getColumnFamilyRange starts a range read in a goroutine and returns a channel.
func getColumnFamilyRange(cf string, pool gossie.ConnectionPool, startPrefix string,
	endPrefix string, maxResults int) <-chan rowResults {
	outChan := make(chan rowResults)

	go func() {
		glog.V(3).Infoln("Reading data for " + cf + " ...")
		rows, err := pool.Reader().Cf(cf).ReturnNilRows(true).RangeGet(
			&gossie.Range{Start: []byte(startPrefix), End: []byte(endPrefix), Count: maxResults})
		if err != nil {
			outChan <- rowResults{nil, err}
		}
		outChan <- rowResults{rows, nil}
	}()

	return outChan
}

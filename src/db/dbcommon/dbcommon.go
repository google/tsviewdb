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

package dbcommon

import (
	"code.google.com/p/goprotobuf/proto"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/golang/glog"
	"github.com/google/tsviewdb/src/db"
	pb "github.com/google/tsviewdb/src/proto"
	"io"
)

const (
	/*
	 For Bigtable:

	 Bigtable rows are sorted in ascending order.  We need a descending sort so we
	 can efficiently retrieve the most-recent value.  We subtract our timestamp
	 from this value to reverse the sort.

	 For Cassandra:

	 Cassandra rows are sorted in ascending order.  Rows can be read in reverse
	 but there is an efficiency hit because the IndexedSliceReader (internal
	 implementation name) has to be used instead of the SimpleSliceReader.  For
	 details see: http://thelastpickle.com/2011/07/04/Cassandra-Query-Plans/
	 For this reason, we'd like a descending sort so we can efficiently retrieve
	 the most-recent value.  We subtract our timestamp from this value to reverse
	 the sort.
	 NOTE: Cassandra cannot perform prefix searches; range searches are between
	       two exact keys.
	*/
	MaxTimeMillis = 9999999999999

	CFChildren   = "children"
	CFAggregates = "aggregates"
	CFPoints     = "points"
	CFConfigs    = "configs"
	CFSource     = "source"
)

///////////////////////////////////////////////////////////////////////////////
// COMMON DB UTILITY FUNCTIONS
///////////////////////////////////////////////////////////////////////////////

// TODO: Use murmur3 instead: https://github.com/dgryski/dgohash
func getMd5Hash(input string) (output string) {
	h := md5.New()
	io.WriteString(h, input)
	return hex.EncodeToString(h.Sum(nil))
}

func MakeRowKey(src string, timestampMillis int64, uuid string) (key string) {
	baseRowKey := getMd5Hash(src)
	// See notes for MaxTimeMillis about effecting a descending sort.
	timestamp := MaxTimeMillis - timestampMillis
	return fmt.Sprintf("%s_%013d_%s", baseRowKey, timestamp, uuid)
}

func PlusOne(input string) string {
	return input[:len(input)-1] + string(input[len(input)-1]+1)
}

func MakeRowPrefixes(src string, startTimestamp, endTimestamp int64,
	endExclusive bool) (startKey, endKey string) {
	baseRowKey := getMd5Hash(src)
	// See notes for MaxTimeMillis about effecting a descending sort.
	startTime := MaxTimeMillis - endTimestamp
	endTime := MaxTimeMillis - startTimestamp

	startKey = fmt.Sprintf("%s_%013d", baseRowKey, startTime)
	endKey = fmt.Sprintf("%s_%013d", baseRowKey, endTime)

	glog.V(4).Infoln("startKey", startKey)
	glog.V(4).Infoln("endKey", endKey)

	if endExclusive {
		endKeyPlusOne := PlusOne(endKey)
		glog.V(4).Infoln("endKeyPlusOne", endKeyPlusOne)
		return startKey, endKeyPlusOne
	}
	return startKey, endKey
}

func GetTimestamp(rowKey []byte) (parsedTimestamp int64) {
	if len(rowKey) >= 46 { // (32 hash) + _ + (13 timestamp) = 46
		fmt.Sscanf(string(rowKey[33:46]), "%d", &parsedTimestamp)
		parsedTimestamp = MaxTimeMillis - parsedTimestamp
	}
	return
}

func SerializeSourceInfoUncomp(si db.SourceInfoUncomp) []byte {
	s := &pb.SourceInfo{}

	// Build dict as we set units indices.
	unitsDict := make(map[string]int)
	var mapCounter int
	for _, unitName := range si.Units {
		idx, exists := unitsDict[unitName]
		if !exists {
			unitsDict[unitName] = mapCounter
			mapCounter++
		}
		s.UnitsIndices = append(s.UnitsIndices, int32(idx))
	}
	s.UnitsMap = make([]string, mapCounter)
	// Invert unitsDict to make serialized map.
	for unitName, idx := range unitsDict {
		s.UnitsMap[idx] = unitName
	}

	s.MetricNames = si.Names
	s.SelectForDefaults = si.SelectForDefaults

	data, _ := proto.Marshal(s)
	return data
}

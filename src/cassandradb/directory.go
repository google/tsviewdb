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
	"fmt"
	"github.com/adilhn/gossie/src/gossie"
	"github.com/golang/glog"
	"github.com/google/tsviewdb/src/common"
	"github.com/google/tsviewdb/src/db"
	"github.com/google/tsviewdb/src/db/dbcommon"
	pb "github.com/google/tsviewdb/src/proto"
	"strings"
)

func match(req db.DirectorySearchRequest, columnName []byte) bool {
	return (req.FileRestrict == "") ||
		req.FilePrefixMatch && strings.HasPrefix(string(columnName), req.FileRestrict) ||
		!req.FilePrefixMatch && (string(columnName) == req.FileRestrict)
}

func (c *CassandraDB) ReadDir(req db.DirectorySearchRequest) (sInfo db.SourceInfoUncomp, err error) {
	prefix := "/" + req.Prefix
	var prefixPlusOne string
	if req.DirPrefixMatch {
		prefixPlusOne = dbcommon.PlusOne(prefix)
	} else {
		prefixPlusOne = prefix
	}

	rows, err := c.pool.Reader().Cf(dbcommon.CFChildren).RangeGet(
		&gossie.Range{Start: []byte(prefix), End: []byte(prefixPlusOne), Count: 100})
	if err != nil {
		return db.SourceInfoUncomp{}, err
	}

	for _, row := range rows {
		rowName := row.Key[1:]
		for _, column := range row.Columns {
			if !match(req, column.Name) {
				continue
			}
			s := new(pb.SourceInfo)
			if err := proto.Unmarshal(column.Value, s); err != nil {
				return db.SourceInfoUncomp{}, err
			}
			if req.ReturnMetrics || req.ReturnUnits {
				for nameIndex, metricName := range s.MetricNames {
					selectForDefaultsConsistent := len(s.SelectForDefaults) == len(s.MetricNames)
					outputOkay := !req.DefaultsOnly ||
						(req.DefaultsOnly && selectForDefaultsConsistent && s.SelectForDefaults[nameIndex])
					if !outputOkay {
						continue
					}
					name := fmt.Sprintf("%s/%s:%s", rowName, column.Name, metricName)
					sInfo.Names = append(sInfo.Names, name)
					unitIndicesConsistent := len(s.UnitsIndices) == len(s.MetricNames)
					if req.ReturnUnits && unitIndicesConsistent {
						units := s.UnitsMap[s.UnitsIndices[nameIndex]]
						sInfo.Units = append(sInfo.Units, units)
					}
					if req.ReturnSelectForDefaults && (selectForDefaultsConsistent) {
						sInfo.SelectForDefaults = append(sInfo.SelectForDefaults, s.SelectForDefaults[nameIndex])
					}

				}
			} else {
				name := fmt.Sprintf("%s/%s", rowName, column.Name)
				sInfo.Names = append(sInfo.Names, name)
			}
		}
	}

	return sInfo, nil
}

func (c *CassandraDB) WriteDir(si db.SourceInfoUncomp, src string) (err error) {
	glog.V(2).Infoln("Start directory mutation for: " + src)
	path, file := common.GetSrcComponents(src)
	row := &gossie.Row{[]byte("/" + path), nil}

	row.Columns = append(row.Columns, &gossie.Column{
		Name:  []byte(file),
		Value: dbcommon.SerializeSourceInfoUncomp(si),
	})

	err = c.pool.Writer().Insert(dbcommon.CFChildren, row).Run()
	glog.V(2).Infoln("Done directory mutation.")
	return
}

func (c *CassandraDB) DeleteDir(path, file string) (err error) {
	writer := c.pool.Writer()
	return writer.DeleteColumns(dbcommon.CFChildren, []byte("/"+path), [][]byte{[]byte(file)}).Run()
}

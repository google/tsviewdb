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
	"github.com/google/tsviewdb/src/db/dbcommon"
)

func (c *CassandraDB) DeleteRow(rowKey string) (err error) {
	writer := c.pool.Writer()
	return writer.
		Delete(dbcommon.CFAggregates, []byte(rowKey)).
		Delete(dbcommon.CFPoints, []byte(rowKey)).
		Delete(dbcommon.CFSource, []byte(rowKey)).
		Delete(dbcommon.CFConfigs, []byte(rowKey)).Run()
}

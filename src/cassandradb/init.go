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

// Package cassandradb provides a Cassandra implementation of db.DB.
package cassandradb

import (
	"flag"
	"github.com/adilhn/gossie/src/gossie"
	"github.com/golang/glog"
	"github.com/google/tsviewdb/src/db"
)

var keyspace = flag.String("cassandradb.keyspace", "perf", "Cassandra DB keyspace to use")
var poolSize = flag.Int("cassandradb.poolSize", 10, "Cassandra DB connection pool size.")

func New() db.DB {
	return &CassandraDB{}
}

type CassandraDB struct {
	pool gossie.ConnectionPool // DB Connections.
}

func (c *CassandraDB) Init() (err error) {
	glog.Infoln("Opening Cassandra DB connection..")
	c.pool, err = gossie.NewConnectionPool([]string{"localhost:9160"}, *keyspace,
		gossie.PoolOptions{Size: *poolSize, Timeout: 3000})
	return
}

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

package main

import (
	"flag"
	"fmt"
	"github.com/golang/glog"
	"net/http"
	"os"
	"runtime"
	"github.com/google/tsviewdb/src/cassandradb"
	"github.com/google/tsviewdb/src/db"
	"github.com/google/tsviewdb/src/handlers"
)

var servicePort = flag.Int("port", 8080, "API service port.")
var useDB = flag.String("useDB", "cassandra", "DB to use.")

func setup() {
	flag.Parse()
	numCPU := runtime.NumCPU()
	glog.Infoln("NumCPU", numCPU)
	if envMaxProcs := os.Getenv("GOMAXPROCS"); envMaxProcs == "" {
		if numCPU > 1 {
			// Consuming N-1 appears to greatly reduce per-request latency in loaded systems.
			runtime.GOMAXPROCS(numCPU - 1)
		}
	}
	glog.Infoln("GOMAXPROCS", runtime.GOMAXPROCS(0))

	var d db.DB
	switch *useDB {
	case "cassandra":
		d = cassandradb.New()
	default:
		glog.Fatalln("Unknown DB:", *useDB)
	}

	if err := d.Init(); err != nil {
		glog.Fatalln("An error occured Initializing the DB: ", err)
	}
	handlers.InitializeAndRegister(d)
}

func main() {
	setup()
	glog.Infof("Starting TSViewDB server on port %d\n", *servicePort)
	glog.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *servicePort), nil))
}

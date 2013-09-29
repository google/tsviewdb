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

// Package cachinghandler provides a handler-oriented interface to groupcache.
//
// Use like this:
//
// func main() {
// 	Initialize()
// 	myCreator := func(D db.DB, bf *bytes.Buffer, key string) error {
// 		time.Sleep(5 * time.Second) /* Generate expensive content here using D. */
// 		fmt.Fprintf(bf, "At %v Path is: %s", time.Now(), key)
// 	}
// 	RegisterCacheContentCreator(DB, "myGroup", myCreator, "text/html", true)
// 	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
// 		HandleWithCache(w, r, "myGroup", r.URL.Path)
// 	})
// 	http.ListenAndServe(":8080", nil)
// }
//
package cachinghandler

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/adilhn/groupcache-expiration"
	"github.com/golang/glog"
	"github.com/google/tsviewdb/src/db"
	"github.com/google/tsviewdb/src/gziphandler"
	"github.com/google/tsviewdb/src/handlers/handlerutils"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var defaultCacheExpiration, _ = time.ParseDuration("20s")
var defaultCacheStalePeriod, _ = time.ParseDuration("48h")
var defaultCacheStaleDeadline, _ = time.ParseDuration("50ms")

var cacheExpiration = flag.Duration("cacheExpiration", defaultCacheExpiration,
	"Duration before cached data is considered expired.")
var cacheStalePeriod = flag.Duration("cacheStalePeriod", defaultCacheStalePeriod,
	"Duration after cache expiration where stale data may be served.")
var cacheStaleDeadline = flag.Duration("cacheStaleDeadline", defaultCacheStaleDeadline,
	"Duration that we'll wait for a regen before returning stale data.  "+
		"Typically very short in milliseconds.")

var cacheSelfAddress = flag.String("cacheSelfAddress", "http://localhost:8080",
	"Address for self cache server.  Must include in -cachePeerAddresses "+
		"if that is set.")
var cachePeerAddresses peerAddresses

type peerAddresses []string // User-defined flag type.

func (p *peerAddresses) String() string {
	return fmt.Sprint(*p)
}

func (p *peerAddresses) Set(value string) error {
	for _, v := range strings.Split(value, ",") {
		*p = append(*p, v)
	}
	return nil
}

func init() {
	flag.Var(&cachePeerAddresses, "cachePeerAddresses",
		"Comma-separated list of cache peer addresses as valid URLs "+
			"(example: http://localhost:8091).  Must include self.")
}

const (
	cacheByteSize = 256 << 20
)

// Initialize should be called before any requests are serviced.
func Initialize() {
	if len(cachePeerAddresses) > 0 {
		groupcache.NewHTTPPool(*cacheSelfAddress).Set(cachePeerAddresses...)
	}
}

type genFunction func(bf *bytes.Buffer) error

func setCacheResult(dest groupcache.Sink, fill genFunction, zip bool) error {
	glog.V(2).Infoln("Start content generation.")
	var uncompressedContent *bytes.Buffer
	// TODO: Use variable-sized buffer pool for uncompressedContent and content.
	//       Already tried simple buffer bool for uncompressedContent without much
	//       apparent latency benefit, but with large memory overhead.
	uncompressedContent = &bytes.Buffer{}

	if err := fill(uncompressedContent); err != nil {
		return err
	}
	if zip {
		t1 := time.Now()
		// TODO: Use a buffer pool below.  This ultimately gets copied to the outgoing
		//       stream, and is more complicated to release because the outgoing
		//       stream bytes could also come from the cache.
		var content bytes.Buffer
		err := gziphandler.GzipContent(&content, uncompressedContent.Bytes())
		glog.V(3).Infof("PERF: gzip time: %v\n", time.Now().Sub(t1))
		if err != nil {
			return err
		}
		return dest.SetTimestampBytes(content.Bytes(), groupcache.GetTime())
	} else {
		return dest.SetTimestampBytes(uncompressedContent.Bytes(), groupcache.GetTime())
	}
}

// TODO: Make the DB interface generic using interface{} (though at the expense
//       of an added type assertion in the critical path).
type CacheContentCreator func(D db.DB, bf *bytes.Buffer, key string) error

type contentInfo struct {
	contentType string
	zip         bool // True if content should be compressed.
}

var (
	mu           sync.Mutex
	contentInfos = make(map[string]contentInfo)
)

// RegisterCacheContentCreator registers a content creation function so that it may
// be addressed later by HandleWithCache().  Must not be called concurrently with
// HandleWithCache().
func RegisterCacheContentCreator(D db.DB, group string, creator CacheContentCreator,
	contentType string, zip bool) {
	groupcache.NewGroup(group, cacheByteSize,
		groupcache.GetterFunc(func(gctx groupcache.Context, key string, dest groupcache.Sink) error {
			return setCacheResult(dest, func(bf *bytes.Buffer) error {
				return creator(D, bf, key)
			}, zip)
		})).SetExpiration(*cacheExpiration).SetStalePeriod(*cacheStalePeriod).SetStaleDeadline(*cacheStaleDeadline).
		SetDisableHotCache(true)

	mu.Lock()
	defer mu.Unlock()
	contentInfos[group] = contentInfo{contentType: contentType, zip: zip}
}

// HandleWithCache is called from a handler to either respond with a cached value
// or generate a new one.
func HandleWithCache(w http.ResponseWriter, r *http.Request, groupName, key string) {
	content, timestamp, err := getContentUsingCache(groupName, key)
	if err != nil {
		handlerutils.HttpError(w, err.Error(), http.StatusBadRequest)
		return
	}
	cInfo := contentInfos[groupName]
	returnWithContentType(w, r, content, cInfo.contentType, timestamp, cInfo.zip)
}

func getContentUsingCache(cacheGroup, key string) (content []byte, timestamp int64, err error) {
	var packedContent []byte
	group := groupcache.GetGroup(cacheGroup)

	if err := group.Get(nil, key, groupcache.AllocatingByteSliceSink(&packedContent)); err != nil {
		return content, timestamp, err
	}
	return groupcache.UnpackTimestamp(packedContent)
}

func returnWithContentType(w http.ResponseWriter, r *http.Request, content []byte,
	contentType string, timestamp int64, zip bool) {
	if handlerutils.EtagMatch(w, r, content) {
		return
	}
	if zip {
		w.Header().Set("Content-Encoding", "gzip")
	}
	setAge(w, timestamp)
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", fmt.Sprintf("private, max-age=%d", 20))
	w.Header().Set("Content-Length", strconv.Itoa(len(content)))
	t1 := time.Now()
	w.Write(content)
	glog.V(3).Infof("PERF: bytebuffer write time: %v\n", time.Now().Sub(t1))
}

func setAge(w http.ResponseWriter, timestamp int64) {
	userViewAge := time.Now().Unix() - timestamp
	glog.V(3).Infof("userViewAge = %d", userViewAge)
	w.Header().Set("Age", strconv.FormatInt(userViewAge, 10))
}

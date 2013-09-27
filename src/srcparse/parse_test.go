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

package srcparse

import (
	"github.com/davecgh/go-spew/spew"
	"testing"
)

type testCase struct {
	input string
	want  SrcResult
}

var testCases = []testCase{
	testCase{"src",
		SrcResult{Source: "src"}},
	testCase{"src:metric",
		SrcResult{Source: "src", Metric: "metric"}},
	testCase{"src:metric.aggregate",
		SrcResult{Source: "src", Metric: "metric", Aggregate: "aggregate"}},
	testCase{"src:*.aggregate",
		SrcResult{Source: "src", Aggregate: "aggregate"}},
	testCase{"src:metric.aggregate$key1",
		SrcResult{Source: "src", Metric: "metric", Aggregate: "aggregate", Configs: map[string]string{"key1": ""}}},
	testCase{"src$key1",
		SrcResult{Source: "src", Configs: map[string]string{"key1": ""}}},
	testCase{"src:metric$key1",
		SrcResult{Source: "src", Metric: "metric", Configs: map[string]string{"key1": ""}}},
	testCase{"src:metric.aggregate$key1=value1",
		SrcResult{Source: "src", Metric: "metric", Aggregate: "aggregate", Configs: map[string]string{"key1": "value1"}}},
	testCase{"src:metric.aggregate$key1=value1$key2=value2",
		SrcResult{Source: "src", Metric: "metric", Aggregate: "aggregate", Configs: map[string]string{"key1": "value1", "key2": "value2"}}},
	testCase{"src:metric.aggregate$key1$key2=value2",
		SrcResult{Source: "src", Metric: "metric", Aggregate: "aggregate", Configs: map[string]string{"key1": "", "key2": "value2"}}},
	testCase{"src:metric.aggregate$key1$key2",
		SrcResult{Source: "src", Metric: "metric", Aggregate: "aggregate", Configs: map[string]string{"key1": "", "key2": ""}}},
}

func TestAll(t *testing.T) {
	for i, tc := range testCases {
		got := Parse(tc.input)
		if !got.Equal(tc.want) {
			t.Errorf("TC:%d got: %s\nwant: %s\n", i, spew.Sdump(got), spew.Sdump(tc.want))
		}
	}
}

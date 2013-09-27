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

package regress

import (
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"strconv"
	"testing"
)

func makeTestTable(i []interface{}) (r []*[]*float64) {
	d := dataRow(i)
	r = make([]*[]*float64, len(d))
	for i := range d {
		row := make([]*float64, 1)
		row[0] = d[i]
		r[i] = &row
	}
	return
}

func dataRow(d []interface{}) (result []*float64) {
	result = make([]*float64, len(d))
	for i := range d {
		switch t := d[i].(type) {
		default:
			result[i] = nil
		case int:
			f := float64(t)
			result[i] = &f
		case float64:
			result[i] = &t
		}
	}

	return result
}

func run(r RegressionParams, input, want []interface{}) error {
	d := makeTestTable(input)
	rTable := regressTable(d)

	got := rTable.GetVerifiedRegression(0, r)
	cookedWant := dataRow(want)

	if len(got) != len(want) {
		return errors.New(fmt.Sprintf("len(want) != len(got); %d != %d", len(want), len(got)))
	}

	for i := range got {
		g := got[i]
		w := cookedWant[i]
		if g == nil && w == nil {
			continue
		}
		if g == nil || w == nil {
			return errors.New("Want:" + spew.Sdump(cookedWant) + "\nGot:" + spew.Sdump(got))
		}
		gStr := fmt.Sprintf("%.2f", *g)
		wStr := fmt.Sprintf("%.2f", *w)
		if gStr != wStr {
			return errors.New("Want:" + spew.Sdump(cookedWant) + "\nGot:" + spew.Sdump(got))
		}
	}
	return nil
}

type testCase struct {
	r     RegressionParams
	input []interface{}
	want  []interface{}
}

var pos0 = float64(0)
var pos1 = float64(1)
var pos10 = float64(10)
var neg0 = float64(0)
var neg1 = float64(-1)
var neg10 = float64(-10)

var testCases = []testCase{
	// Straight delta functionality.
	testCase{RegressionParams{Radius: 0, Window: 1},
		[]interface{}{1, 2, 3, 1, 2, 3, -1, -2, -3, 0},
		[]interface{}{nil, 1, 1, -2, 1, 1, -4, -1, -1, 3},
	},
	// Straight delta functionality, window > 1.
	testCase{RegressionParams{Radius: 0, Window: 2},
		[]interface{}{1, 2, 3, 1, 2, 3, -1, -2, -3, 0},
		[]interface{}{nil, nil, 0, -1, 0, 0, -3, -2, 0, nil},
	},
	// Straight delta functionality with percent output.
	testCase{RegressionParams{Radius: 0, Window: 1, UsePercent: true},
		[]interface{}{1, 2, 3, 1, 2, 3, -1, -2, -3, 0},
		[]interface{}{nil, 100, 50, -66.67, 100, 50, -133.33, -100, -50, 100},
	},
	// Straight delta functionality, window > 1, percent output.
	testCase{RegressionParams{Radius: 0, Window: 2, UsePercent: true},
		[]interface{}{1, 2, 3, 1, 2, 3, -1, -2, -3, 0},
		[]interface{}{nil, nil, 0, -50, 0, 0, -150, -66.67, 0, nil},
	},
	// Straight delta functionality with nil inputs.
	testCase{RegressionParams{Radius: 0, Window: 1},
		[]interface{}{nil, 2, 3, nil, 2, 3, -1, nil, -3, 0},
		[]interface{}{nil, nil, 1, nil, nil, 1, -4, nil, nil, 3},
	},
	// Basic use with small radius.
	testCase{RegressionParams{Radius: 1, Window: 1},
		[]interface{}{1, 1, 1, 1, 1, 10, 10, 10, 10, 10},
		[]interface{}{nil, nil, 0, 0, 0, 9, 0, 0, 0, nil},
	},
	// Basic use with large radius.
	testCase{RegressionParams{Radius: 4, Window: 1},
		[]interface{}{1, 1, 1, 1, 1, 10, 10, 10, 10, 10},
		[]interface{}{nil, nil, nil, nil, nil, 9, nil, nil, nil, nil},
	},
	// Basic use with small radius and window > 1.
	testCase{RegressionParams{Radius: 1, Window: 2},
		[]interface{}{1, 1, 1, 1, 1, 10, 10, 10, 10, 10},
		[]interface{}{nil, nil, nil, 0, 0, 9, 0, 0, nil, nil},
	},

	// Basic use with small radius and under positive threshold.
	testCase{RegressionParams{Radius: 1, Window: 1, Pos: &pos1},
		[]interface{}{1, 1, 1, 1, 1, 10, 10, 10, 10, 10},
		[]interface{}{nil, nil, 0, 0, 0, 9, 0, 0, 0, nil},
	},
	// Basic use with small radius and over positive threshold.
	testCase{RegressionParams{Radius: 1, Window: 1, Pos: &pos10},
		[]interface{}{1, 1, 1, 1, 1, 10, 10, 10, 10, 10},
		[]interface{}{},
	},
	// Basic use with small radius and under negative threshold.
	testCase{RegressionParams{Radius: 1, Window: 1, Neg: &neg1},
		[]interface{}{-1, -1, -1, -1, -1, -10, -10, -10, -10, -10},
		[]interface{}{nil, nil, 0, 0, 0, -9, 0, 0, 0, nil},
	},
	// Basic use with small radius and over negative threshold.
	testCase{RegressionParams{Radius: 1, Window: 1, Neg: &neg10},
		[]interface{}{-1, -1, -1, -1, -1, -10, -10, -10, -10, -10},
		[]interface{}{},
	},

	// Basic use with small radius and under positive threshold and return segments.
	testCase{RegressionParams{Radius: 1, Window: 1, Pos: &pos1, ReturnSegments: true},
		[]interface{}{1, 1, 1, 1, 1, 10, 10, 10, 10, 10},
		[]interface{}{nil, nil, nil, nil, 1, 10, nil, nil, nil, nil},
	},
	// Basic use with small radius and over positive threshold and return segments.
	testCase{RegressionParams{Radius: 1, Window: 1, Pos: &pos10, ReturnSegments: true},
		[]interface{}{1, 1, 1, 1, 1, 10, 10, 10, 10, 10},
		[]interface{}{},
	},
	// Basic use with small radius and under negative threshold and return segments.
	testCase{RegressionParams{Radius: 1, Window: 1, Neg: &neg1, ReturnSegments: true},
		[]interface{}{-1, -1, -1, -1, -1, -10, -10, -10, -10, -10},
		[]interface{}{nil, nil, nil, nil, -1, -10, nil, nil, nil, nil},
	},
	// Basic use with small radius and over negative threshold and return segments.
	testCase{RegressionParams{Radius: 1, Window: 1, Neg: &neg10, ReturnSegments: true},
		[]interface{}{-1, -1, -1, -1, -1, -10, -10, -10, -10, -10},
		[]interface{}{},
	},
}

func TestAll(t *testing.T) {
	for i := range testCases {
		tc := testCases[i]
		err := run(tc.r, tc.input, tc.want)
		if err != nil {
			t.Log("Test case: " + strconv.Itoa(i) + "\n" + err.Error())
			t.Fail()
		}
	}
}

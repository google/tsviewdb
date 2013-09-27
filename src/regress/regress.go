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
	"github.com/davecgh/go-spew/spew"
	"github.com/golang/glog"
	"math"
)

var _ = spew.Dump
var _ = glog.Infoln

// Not using math.Min and math.Max because need to be fast more than correct.
func minFloat64(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func maxFloat64(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

type RegressionParams struct {
	Selected       bool
	Pos            *float64 // Positive regression threshold.
	Neg            *float64 // Negative regression threshold.
	ReturnSegments bool     // Return regression segments, otherwise return regression function values.

	Radius     int     // Size of array ahead and behind of window to use to verify regression.
	Window     int     // Distance in records to use to calculate delta.  Must be >= 1.
	UsePercent bool    // Return values are in percent.
	IgnoreLT   float64 // Ignore absolute values < this amount if UsePercent is selected.
}

type regressTable []*[]*float64 // Not ptr because will never add rows.

func NewTable(t []*[]*float64) regressTable {
	newTable := regressTable(t)
	return newTable
}

func (t regressTable) get(row, col int) (val float64, ok bool) {
	valPtr := (*t[row])[col]
	if valPtr == nil {
		return 0, false
	}
	return *valPtr, true
}

func (t regressTable) getRaw(row, col int) *float64 {
	return (*t[row])[col]
}

func (t regressTable) GetVerifiedRegression(col int, r RegressionParams) (result []*float64) {
	result = t.computeVerifiedRegression(col, r)
	if (r.Pos == nil) && (r.Neg == nil) { // No threshold set.
		return result
	}

	posSet := r.Pos != nil
	negSet := r.Neg != nil

	var haveRegression bool
	for i := range result {
		if result[i] == nil {
			continue
		}
		if (posSet && (*result[i] > *r.Pos)) || (negSet && (*result[i] < *r.Neg)) {
			if r.ReturnSegments {
				result[i] = t.getRaw(i, col)
				if i > 0 { // Add the previous point if available.
					result[i-1] = t.getRaw(i-1, col)
				}
			}
			haveRegression = true
		} else {
			if r.ReturnSegments {
				result[i] = nil
			}
		}
	}

	if !haveRegression {
		return nil
	}
	return result
}

// computeVerifiedRegression uses these parameters:
//
//              + - \
//            /       /
//          /
//  - / \ /
//
// |r|r|r|b| | |c|r|r|r|
//  +-+-+  +----+ +-+-+
//  |      |      |
//  radius |      radius
//        window
//
func (t regressTable) computeVerifiedRegression(col int, r RegressionParams) (result []*float64) {
	result = make([]*float64, len(t))

	// Ensure the index is high-enough to contain our window and all the reverse
	// points, and that it is low enough to contain all our forward points.
	for n := r.Window + r.Radius; n <= len(t)-r.Radius-r.Window; n++ {
		nBack := n - r.Window // The "back" point index.
		back, ok := t.get(nBack, col)
		if !ok {
			continue
		}

		if r.UsePercent && (back == 0) { // Guard for ZeroDivisionError.
			continue
		}

		current, ok := t.get(n, col)
		if !ok {
			continue
		}
		if r.UsePercent && ((math.Abs(back) < r.IgnoreLT) || (math.Abs(current) < r.IgnoreLT)) {
			continue
		}
		backDelta := current - back

		var confDeltaPos, confDeltaNeg float64
		if r.Radius != 0 {
			confDeltaPos, confDeltaNeg, ok = t.getConfirmedDeltas(col, n, nBack, r.Radius)
			if !ok {
				continue
			}
		}

		var absResult float64
		if r.Window == 1 {
			if r.Radius == 0 {
				absResult = backDelta
			} else {
				if backDelta > 0 {
					absResult = minFloat64(backDelta, confDeltaPos)
				} else {
					absResult = maxFloat64(backDelta, confDeltaNeg)
				}
			}
			if r.UsePercent {
				percentAbsResult := (absResult / math.Abs(back)) * 100
				result[n] = &percentAbsResult
				continue
			}
			result[n] = &absResult
			continue
		}

		// fwd_delta is used to clean up artifacts if the window size is > 1.
		nFwdVal, ok := t.get(n+r.Window-1, col)
		if !ok {
			continue
		}
		nM1Val, ok := t.get(n-1, col)
		if !ok {
			continue
		}
		fwdDelta := nFwdVal - nM1Val

		if backDelta > 0 {
			if fwdDelta < 0 {
				fwdDelta = 0
			}
			if r.Radius == 0 {
				absResult = minFloat64(backDelta, fwdDelta)
			} else {
				absResult = minFloat64(minFloat64(backDelta, fwdDelta), confDeltaPos)
			}
		} else {
			if fwdDelta > 0 {
				fwdDelta = 0
			}
			if r.Radius == 0 {
				absResult = maxFloat64(backDelta, fwdDelta)
			} else {
				absResult = maxFloat64(maxFloat64(backDelta, fwdDelta), confDeltaNeg)
			}
		}

		if r.UsePercent {
			percentAbsResult := (absResult / math.Abs(back)) * 100
			result[n] = &percentAbsResult
		} else {
			result[n] = &absResult
		}
	}
	return result
}

func (t regressTable) getConfirmedDeltas(col, n, nBack, radius int) (confDeltaPos, confDeltaNeg float64, ok bool) {
	// Find min/max in array of num_back values starting before the back point.
	//maxBack := *(*t[nBack-radius])[col]
	maxBack, ok := t.get(nBack-radius, col)
	if !ok {
		return confDeltaPos, confDeltaNeg, false // Need all points valid
	}

	minBack := maxBack
	for i := 1; i < radius; i++ { // Skip first element because already set.
		//val := *(*t[nBack-radius+i])[col]
		val, ok := t.get(nBack-radius+i, col)
		if !ok {
			return confDeltaPos, confDeltaNeg, false // Need all points valid
		}
		if val > maxBack {
			maxBack = val
		} else if val < minBack {
			minBack = val
		}
	}

	// Find min/max in array of num_fwd values starting after the current point.
	maxFwd := *(*t[n+1])[col]
	minFwd := maxFwd
	for i := 1; i < radius; i++ { // Skip first element because already set.
		//val := *(*t[n+1+i])[col]
		val, ok := t.get(n+1+i, col)
		if !ok {
			return confDeltaPos, confDeltaNeg, false // Need all points valid
		}
		if val > maxFwd {
			maxFwd = val
		} else if val < minFwd {
			minFwd = val
		}
	}

	confDeltaPos = minFwd - maxBack
	if confDeltaPos < 0 { // Clamp.  Want to see positive only.
		confDeltaPos = 0
	}
	confDeltaNeg = maxFwd - minBack
	if confDeltaNeg > 0 { // Clamp.  Want to see negative only.
		confDeltaNeg = 0
	}
	return
}

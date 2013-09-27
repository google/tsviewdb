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

package tsviewdb

import (
	"sort"
)

func round(value float64) int64 {
	if value < 0.0 {
		value -= 0.5
	} else {
		value += 0.5
	}
	return int64(value)
}

type ptrDecFunc func(*int64) *float64
type ptrEncFunc func(*float64) *int64

type decFunc func(int64) float64
type encFunc func(float64) int64

func makePtrDecFunc(scale float64) ptrDecFunc {
	return func(i *int64) *float64 {
		if i != nil {
			f := new(float64)
			*f = float64(*i) / scale
			return f
		}
		return nil
	}
}

func makePtrEncFunc(scale float64) ptrEncFunc {
	return func(f *float64) *int64 {
		if f != nil {
			i := new(int64)
			*i = round(*f * scale)
			return i
		}
		return nil
	}
}

func makeDecFunc(scale float64) decFunc {
	return func(i int64) float64 {
		return float64(i) / scale
	}
}

func makeEncFunc(scale float64) encFunc {
	return func(f float64) int64 {
		return round(f * scale)
	}
}

var ptrDecFuncs = [4]ptrDecFunc{
	makePtrDecFunc(1.0),
	makePtrDecFunc(10.0),
	makePtrDecFunc(100.0),
	makePtrDecFunc(1000.0),
}

var ptrEncFuncs = [4]ptrEncFunc{
	makePtrEncFunc(1.0),
	makePtrEncFunc(10.0),
	makePtrEncFunc(100.0),
	makePtrEncFunc(1000.0),
}

var decFuncs = [4]decFunc{
	makeDecFunc(1.0),
	makeDecFunc(10.0),
	makeDecFunc(100.0),
	makeDecFunc(1000.0),
}

var encFuncs = [4]encFunc{
	makeEncFunc(1.0),
	makeEncFunc(10.0),
	makeEncFunc(100.0),
	makeEncFunc(1000.0),
}

// MakeDouble makes double fields from the scaled fields.  The type is not
// changed.  Memory is freed for the scaled field.
func (m *Aggregation) MakeDouble() {
	if m == nil {
		return
	}

	var dec ptrDecFunc

	switch m.GetType() {
	case DataType_INT64:
		dec = ptrDecFuncs[0]
	case DataType_SCALED1:
		dec = ptrDecFuncs[1]
	case DataType_SCALED2:
		dec = ptrDecFuncs[2]
	case DataType_SCALED3:
		dec = ptrDecFuncs[3]
	default:
		return
	}

	if m.Double == nil {
		m.Double = new(Aggregation_AggregationDouble)
	}

	// Convert.
	m.Double.Count = dec(m.Scaled.Count)
	m.Double.Min = dec(m.Scaled.Min)
	m.Double.Max = dec(m.Scaled.Max)
	m.Double.Mean = dec(m.Scaled.Mean)
	m.Double.Stdev = dec(m.Scaled.Stdev)
	m.Double.P99 = dec(m.Scaled.P99)
	m.Double.P95 = dec(m.Scaled.P95)
	m.Double.P90 = dec(m.Scaled.P90)
	m.Double.P85 = dec(m.Scaled.P85)
	m.Double.P80 = dec(m.Scaled.P80)
	m.Double.P75 = dec(m.Scaled.P75)
	m.Double.P70 = dec(m.Scaled.P70)
	m.Double.P65 = dec(m.Scaled.P65)
	m.Double.P60 = dec(m.Scaled.P60)
	m.Double.P55 = dec(m.Scaled.P55)
	m.Double.P50 = dec(m.Scaled.P50)
	m.Double.P45 = dec(m.Scaled.P45)
	m.Double.P40 = dec(m.Scaled.P40)
	m.Double.P35 = dec(m.Scaled.P35)
	m.Double.P30 = dec(m.Scaled.P30)
	m.Double.P25 = dec(m.Scaled.P25)
	m.Double.P20 = dec(m.Scaled.P20)
	m.Double.P15 = dec(m.Scaled.P15)
	m.Double.P10 = dec(m.Scaled.P10)
	m.Double.P5 = dec(m.Scaled.P5)
	m.Double.P1 = dec(m.Scaled.P1)

	m.Scaled = nil
}

// MakeScaled makes a scaled field from the double field.  The type is not
// changed.  Memory is freed for the double field.
func (m *Aggregation) MakeScaled(dataType DataType) {
	if m == nil {
		return
	}

	var enc ptrEncFunc

	switch dataType {
	case DataType_INT64:
		enc = ptrEncFuncs[0]
	case DataType_SCALED1:
		enc = ptrEncFuncs[1]
	case DataType_SCALED2:
		enc = ptrEncFuncs[2]
	case DataType_SCALED3:
		enc = ptrEncFuncs[3]
	default:
		return
	}

	if m.Scaled == nil {
		m.Scaled = new(Aggregation_AggregationScaled)
	}

	// Convert.
	m.Scaled.Count = enc(m.Double.Count)
	m.Scaled.Min = enc(m.Double.Min)
	m.Scaled.Max = enc(m.Double.Max)
	m.Scaled.Mean = enc(m.Double.Mean)
	m.Scaled.Stdev = enc(m.Double.Stdev)
	m.Scaled.P99 = enc(m.Double.P99)
	m.Scaled.P95 = enc(m.Double.P95)
	m.Scaled.P90 = enc(m.Double.P90)
	m.Scaled.P85 = enc(m.Double.P85)
	m.Scaled.P80 = enc(m.Double.P80)
	m.Scaled.P75 = enc(m.Double.P75)
	m.Scaled.P70 = enc(m.Double.P70)
	m.Scaled.P65 = enc(m.Double.P65)
	m.Scaled.P60 = enc(m.Double.P60)
	m.Scaled.P55 = enc(m.Double.P55)
	m.Scaled.P50 = enc(m.Double.P50)
	m.Scaled.P45 = enc(m.Double.P45)
	m.Scaled.P40 = enc(m.Double.P40)
	m.Scaled.P35 = enc(m.Double.P35)
	m.Scaled.P30 = enc(m.Double.P30)
	m.Scaled.P25 = enc(m.Double.P25)
	m.Scaled.P20 = enc(m.Double.P20)
	m.Scaled.P15 = enc(m.Double.P15)
	m.Scaled.P10 = enc(m.Double.P10)
	m.Scaled.P5 = enc(m.Double.P5)
	m.Scaled.P1 = enc(m.Double.P1)

	m.Double = nil
}

// SetDoubleField sets individual aggregates in the double field.
func (m *Aggregation) SetDoubleField(field string, value *float64) {
	d := m.Double

	switch field {
	case "count":
		d.Count = value
	case "min":
		d.Min = value
	case "max":
		d.Max = value
	case "mean":
		d.Mean = value
	case "stdev":
		d.Stdev = value
	case "p99":
		d.P99 = value
	case "p95":
		d.P95 = value
	case "p90":
		d.P90 = value
	case "p85":
		d.P85 = value
	case "p80":
		d.P80 = value
	case "p75":
		d.P75 = value
	case "p70":
		d.P70 = value
	case "p65":
		d.P65 = value
	case "p60":
		d.P60 = value
	case "p55":
		d.P55 = value
	case "p50":
		d.P50 = value
	case "p45":
		d.P45 = value
	case "p40":
		d.P40 = value
	case "p35":
		d.P35 = value
	case "p30":
		d.P30 = value
	case "p25":
		d.P25 = value
	case "p20":
		d.P20 = value
	case "p15":
		d.P15 = value
	case "p10":
		d.P10 = value
	case "p5":
		d.P5 = value
	case "p1":
		d.P1 = value
	}
}

type fieldsAndValues struct {
	fields []string
	values []*float64
}

func (m *fieldsAndValues) checkAndSet(value *float64, field string) {
	if value != nil {
		m.fields = append(m.fields, field)
		m.values = append(m.values, value)
	}
}

func GetDoubleFieldsAndValues(a *Aggregation) (fields []string, values []*float64) {
	d := a.Double
	fav := fieldsAndValues{}
	fav.checkAndSet(d.Count, "count")
	fav.checkAndSet(d.Min, "min")
	fav.checkAndSet(d.Max, "max")
	fav.checkAndSet(d.Mean, "mean")
	fav.checkAndSet(d.Stdev, "stdev")
	fav.checkAndSet(d.P99, "p99")
	fav.checkAndSet(d.P95, "p95")
	fav.checkAndSet(d.P90, "p90")
	fav.checkAndSet(d.P85, "p85")
	fav.checkAndSet(d.P80, "p80")
	fav.checkAndSet(d.P75, "p75")
	fav.checkAndSet(d.P70, "p70")
	fav.checkAndSet(d.P65, "p65")
	fav.checkAndSet(d.P60, "p60")
	fav.checkAndSet(d.P55, "p55")
	fav.checkAndSet(d.P50, "p50")
	fav.checkAndSet(d.P45, "p45")
	fav.checkAndSet(d.P40, "p40")
	fav.checkAndSet(d.P35, "p35")
	fav.checkAndSet(d.P30, "p30")
	fav.checkAndSet(d.P25, "p25")
	fav.checkAndSet(d.P20, "p20")
	fav.checkAndSet(d.P15, "p15")
	fav.checkAndSet(d.P10, "p10")
	fav.checkAndSet(d.P5, "p5")
	fav.checkAndSet(d.P1, "p1")
	return fav.fields, fav.values
}

type fieldsAndValuesFiltered struct {
	fields                []string
	values                []*float64
	filter                map[string]bool
	setAggregateIfMissing bool
}

func (m *fieldsAndValuesFiltered) checkAndSetFiltered(value *float64, field string) {
	if m.filter[field] {
		if m.setAggregateIfMissing || (!m.setAggregateIfMissing && value != nil) {
			m.fields = append(m.fields, field)
			m.values = append(m.values, value)
		}
	}
}

func GetDoubleFieldsAndValuesFiltered(a *Aggregation, filter map[string]bool,
	setAggregateIfMissing bool) (fields []string, values []*float64) {
	if filter == nil {
		return GetDoubleFieldsAndValues(a)
	}

	d := a.Double
	fav := fieldsAndValuesFiltered{filter: filter,
		setAggregateIfMissing: setAggregateIfMissing}
	fav.checkAndSetFiltered(d.Count, "count")
	fav.checkAndSetFiltered(d.Min, "min")
	fav.checkAndSetFiltered(d.Max, "max")
	fav.checkAndSetFiltered(d.Mean, "mean")
	fav.checkAndSetFiltered(d.Stdev, "stdev")
	fav.checkAndSetFiltered(d.P99, "p99")
	fav.checkAndSetFiltered(d.P95, "p95")
	fav.checkAndSetFiltered(d.P90, "p90")
	fav.checkAndSetFiltered(d.P85, "p85")
	fav.checkAndSetFiltered(d.P80, "p80")
	fav.checkAndSetFiltered(d.P75, "p75")
	fav.checkAndSetFiltered(d.P70, "p70")
	fav.checkAndSetFiltered(d.P65, "p65")
	fav.checkAndSetFiltered(d.P60, "p60")
	fav.checkAndSetFiltered(d.P55, "p55")
	fav.checkAndSetFiltered(d.P50, "p50")
	fav.checkAndSetFiltered(d.P45, "p45")
	fav.checkAndSetFiltered(d.P40, "p40")
	fav.checkAndSetFiltered(d.P35, "p35")
	fav.checkAndSetFiltered(d.P30, "p30")
	fav.checkAndSetFiltered(d.P25, "p25")
	fav.checkAndSetFiltered(d.P20, "p20")
	fav.checkAndSetFiltered(d.P15, "p15")
	fav.checkAndSetFiltered(d.P10, "p10")
	fav.checkAndSetFiltered(d.P5, "p5")
	fav.checkAndSetFiltered(d.P1, "p1")
	return fav.fields, fav.values
}

// Must ensure len(Data) > 0 before calling any lazyData methods!
type lazyData struct {
	Data []float64

	sorted     bool
	haveMinMax bool
	haveSum    bool

	min float64
	max float64
	sum float64
}

func (s *lazyData) createSorted() {
	if !s.sorted { // Create on demand.
		sort.Float64s(s.Data)
		s.sorted = true
	}
}

func (s *lazyData) createMinMax() {
	if !s.haveMinMax {
		s.min = s.Data[0] // Same for either case below.
		if s.sorted {
			s.max = s.Data[len(s.Data)-1]
		} else {
			s.max = s.Data[0]
			for _, val := range s.Data {
				if val < s.min {
					s.min = val
					continue
				}
				if val > s.max {
					s.max = val
				}
			}
		}
		s.haveMinMax = true
	}
}

func (s *lazyData) createSum() {
	if !s.haveSum {
		for _, val := range s.Data {
			s.sum += val
		}
		s.haveSum = true
	}
}

func (s *lazyData) Min() *float64 {
	s.createMinMax()
	return &s.min
}

func (s *lazyData) Max() *float64 {
	s.createMinMax()
	return &s.max
}

func (s *lazyData) Mean() *float64 {
	s.createSum()
	mean := s.sum / float64(len(s.Data))
	return &mean
}

// p is a fraction from 0 to 1 (o% to 100%)
func (s *lazyData) Percentile(p float32) *float64 {
	s.createSorted()
	return &s.Data[int(float32(len(s.Data))*p)]
}

func (m *Aggregation) CreateMissingDoubleAggregates(d []float64) {
	// Check each double field and, for each of a standard set of aggregates, if
	// not set, generate.
	if m.Double.Count == nil {
		count := float64(len(d))
		m.Double.Count = &count
	}

	if *m.Double.Count <= 0 {
		return // No other calculations make sense when no data.
	}

	l := lazyData{Data: d}

	// Check missing percentiles first, because can reuse sorted array in min/max
	// calculation below.
	if m.Double.P99 == nil {
		m.Double.P99 = l.Percentile(0.99)
	}
	if m.Double.P95 == nil {
		m.Double.P95 = l.Percentile(0.95)
	}
	if m.Double.P90 == nil {
		m.Double.P90 = l.Percentile(0.90)
	}
	if m.Double.P75 == nil {
		m.Double.P75 = l.Percentile(0.75)
	}
	if m.Double.P50 == nil {
		m.Double.P50 = l.Percentile(0.50)
	}
	if m.Double.P25 == nil {
		m.Double.P25 = l.Percentile(0.25)
	}
	if m.Double.P10 == nil {
		m.Double.P10 = l.Percentile(0.10)
	}
	if m.Double.P5 == nil {
		m.Double.P5 = l.Percentile(0.05)
	}
	if m.Double.P1 == nil {
		m.Double.P1 = l.Percentile(0.01)
	}

	if m.Double.Min == nil {
		m.Double.Min = l.Min()
	}
	if m.Double.Max == nil {
		m.Double.Max = l.Max()
	}
	if m.Double.Mean == nil {
		m.Double.Mean = l.Mean()
	}
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// MakeValuesDouble makes doubles from the delta scaled points.  The type is not
// changed.  Memory is freed for the scaled points.
func (m *Points) MakeValuesDouble() {
	if m == nil {
		return
	}

	var dec decFunc

	switch m.GetType() {
	case DataType_INT64:
		dec = decFuncs[0]
	case DataType_SCALED1:
		dec = decFuncs[1]
	case DataType_SCALED2:
		dec = decFuncs[2]
	case DataType_SCALED3:
		dec = decFuncs[3]
	default:
		return
	}

	// Convert.
	var valueScaled int64
	for _, deltaValueScaled := range m.DeltaValuesScaled {
		valueScaled += deltaValueScaled
		m.ValuesDouble = append(m.ValuesDouble, dec(valueScaled))
	}

	m.DeltaValuesScaled = nil
}

// MakeDeltaValuesScaled makes delta scaled points from the double points.  The
// type is not changed.  Memory is freed for the double points.
func (m *Points) MakeDeltaValuesScaled(dataType DataType) {
	if m == nil {
		return
	}

	var enc encFunc

	switch dataType {
	case DataType_INT64:
		enc = encFuncs[0]
	case DataType_SCALED1:
		enc = encFuncs[1]
	case DataType_SCALED2:
		enc = encFuncs[2]
	case DataType_SCALED3:
		enc = encFuncs[3]
	default:
		return
	}

	// Convert.
	var previousValueScaled int64
	for _, valueDouble := range m.ValuesDouble {
		valueScaled := enc(valueDouble)
		m.DeltaValuesScaled = append(m.DeltaValuesScaled, valueScaled-previousValueScaled)
		previousValueScaled = valueScaled
	}

	m.ValuesDouble = nil
}

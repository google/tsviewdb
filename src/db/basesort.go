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

package db

import (
	"errors"
	"github.com/google/tsviewdb/src/common"
	"sort"
)

/////////////////////////////////////////////////////////////////////////////

type StringSlice []string

func (s StringSlice) IndexForName(name string) (int, error) {
	var idx int
	for ; idx < len(s); idx++ {
		if s[idx] == name {
			break
		}
	}
	if idx == len(s) {
		return 0, errors.New("Non-existent name: " + name)
	}
	return idx, nil
}

/////////////////////////////////////////////////////////////////////////////

type parallelStringsFloats struct {
	names  []string   // Not ptr because will never add elements.
	floats []*float64 // Not ptr because will never add elements.
}

func (p *parallelStringsFloats) Len() int {
	return len(p.names)
}

func (p *parallelStringsFloats) Swap(i, j int) {
	p.names[i], p.names[j] = p.names[j], p.names[i]
	p.floats[i], p.floats[j] = p.floats[j], p.floats[i]
}

func (p *parallelStringsFloats) Less(i, j int) bool {
	return p.names[i] < p.names[j]
}

/////////////////////////////////////////////////////////////////////////////

type parallelStringsFloatTable struct {
	names *[]string     // Ptr because will add elements.
	data  []*[]*float64 // Not ptr because Will never add rows (only columns).
}

func (p *parallelStringsFloatTable) addRecordNumColumn() {
	p.names = common.PrependStringSlice(p.names, common.RecordNumName)
	for i := range p.data {
		recordNum := float64(i)
		p.data[i] = common.PrependFloatPtrSlice(p.data[i], &recordNum)
	}
}

// addColumn adds a column to the left-most row.
func (p *parallelStringsFloatTable) addColumn(colName string, col []*float64) {
	*p.names = append(*p.names, colName)
	for i := range p.data {
		*p.data[i] = append(*p.data[i], col[i])
	}
}

// // swapColumnWithArray swaps the left-most column with the given array.
// func (p *parallelStringsFloatTable) swapColumnWithArray(colName string, array []*float64) {
// 	(*p.names)[0] = colName
// 	for i := range p.data {
// 		array[i], (*p.data[i])[0] = (*p.data[i])[0], array[i]
// 	}
// }

// copyColumnToX copies the given column name to the 0 position and returns the
// original in an array.  This is used for changing the X axis to an arbitrary
// column and returning the original X axis as an array.
func (p *parallelStringsFloatTable) copyColumnToX(colIdx int, newColName string,
	savedX *[]*float64) {
	*savedX = make([]*float64, len(p.data))

	(*p.names)[0] = newColName
	for i := range p.data {
		(*savedX)[i] = (*p.data[i])[0]
		(*p.data[i])[0] = (*p.data[i])[colIdx]
	}
}

func deleteFloatPtrSliceItem(a []*float64, i int) []*float64 {
	copy(a[i:], a[i+1:])
	a[len(a)-1] = nil
	return a[:len(a)-1]
}

func deleteStrSliceItem(a []string, i int) []string {
	copy(a[i:], a[i+1:])
	a[len(a)-1] = ""
	return a[:len(a)-1]
}

func (p *parallelStringsFloatTable) deleteColumnbyIdx(colIdx int) {
	*p.names = deleteStrSliceItem(*p.names, colIdx)
	for i := range p.data {
		*p.data[i] = deleteFloatPtrSliceItem(*p.data[i], colIdx)
	}
}

func (p *parallelStringsFloatTable) deleteColumn(colName string) error {
	// Find which column contains colName.
	var idx int
	lenNames := len(*p.names)
	for ; idx < lenNames; idx++ {
		if (*p.names)[idx] == colName {
			break
		}
	}
	if idx == lenNames {
		return errors.New("Non-existent column name for deleteColumn: " + colName)
	}

	p.deleteColumnbyIdx(idx)
	return nil
}

func (p *parallelStringsFloatTable) FixRowLengths() {
	requiredLength := len(*p.names)
	for _, row := range p.data {
		if len(*row) != requiredLength {
			newRow := make([]*float64, requiredLength)
			copy(newRow, *row)
			*row = newRow // Free old row
		}
	}
}

func (p *parallelStringsFloatTable) SortDataColumns() {
	var ps positionString
	ps.Init(p.names)
	sort.Sort(&ps)
	ps.TemplateSortFloat(p.data)
}

/////////////////////////////////////////////////////////////////////////////

type parallelStringsStringTable struct {
	names []string
	data  []*[]*string
}

func (p *parallelStringsStringTable) FixRowLengths() {
	requiredLength := len(p.names)
	for _, row := range p.data {
		if len(*row) != requiredLength {
			newRow := make([]*string, requiredLength)
			copy(newRow, *row)
			*row = newRow // Free old row
		}
	}
}

func (p *parallelStringsStringTable) SortDataColumns() {
	var ps positionString
	ps.Init(&p.names)
	sort.Sort(&ps)
	ps.TemplateSortString(p.data)
}

/////////////////////////////////////////////////////////////////////////////

// tableSort attaches the methods of sort.Interface to []*[]*float64, sorting in
// increasing order by the first column.
type tableSort struct {
	v         []*[]*float64
	extraSwap func(i, j int) // Additional swap function.
	sortIdx   int
}

func (d tableSort) Len() int { return len(d.v) }
func (d tableSort) Less(i, j int) bool {
	di := d.v[i]
	dj := d.v[j]
	if di == nil {
		return true
	}
	if dj == nil {
		return false
	}
	di0 := (*di)[d.sortIdx]
	dj0 := (*dj)[d.sortIdx]
	if di0 == nil {
		return true
	}
	if dj0 == nil {
		return false
	}
	return *di0 < *dj0
}
func (d tableSort) Swap(i, j int) {
	d.v[i], d.v[j] = d.v[j], d.v[i]
	d.extraSwap(i, j)
}

func (d tableSort) ReverseRows() {
	for i, j := 0, len(d.v)-1; i < j; i, j = i+1, j-1 {
		d.Swap(i, j)
	}
}

/////////////////////////////////////////////////////////////////////////////

type positionString struct {
	strings   *[]string
	positions []int
}

func (p *positionString) Init(strings *[]string) {
	p.strings = strings

	p.positions = make([]int, len(*strings))
	for i := 0; i < len(*strings); i++ {
		p.positions[i] = i
	}
}

func (p *positionString) TemplateSortFloat(data []*[]*float64) {
	tmpRow := make([]*float64, len(*p.strings))
	for _, row := range data {
		// To avoid garbage, first copy to reused tmpRow putting in correct position.
		for i := 0; i < len(*row); i++ {
			tmpRow[i] = (*row)[p.positions[i]]
		}
		// Now copy back to original.
		for columnNum, value := range tmpRow {
			(*row)[columnNum] = value
		}
	}
}

func (p *positionString) TemplateSortString(data []*[]*string) {
	tmpRow := make([]*string, len(*p.strings))
	for _, row := range data {
		// To avoid garbage, first copy to reused tmpRow putting in correct position.
		for i := 0; i < len(*row); i++ {
			tmpRow[i] = (*row)[p.positions[i]]
		}
		// Now copy back to original.
		for columnNum, value := range tmpRow {
			(*row)[columnNum] = value
		}
	}
}

func (p *positionString) Len() int {
	return len(*p.strings)
}

func (p *positionString) Swap(i, j int) {
	s := *p.strings
	s[i], s[j] = s[j], s[i]
	p.positions[i], p.positions[j] = p.positions[j], p.positions[i]
}

func (p *positionString) Less(i, j int) bool {
	s := *p.strings
	return s[i] < s[j]
}

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

package rangecontent

import (
	"bytes"
	"errors"
	"github.com/google/tsviewdb/src/common"
	"github.com/google/tsviewdb/src/db"
	"github.com/google/tsviewdb/src/db/requests"
	"github.com/google/tsviewdb/src/png"
	"net/url"
	"strconv"
	"strings"
)

const (
	defaultPngWidth  = 5.0 // Inches
	defaultPngHeight = 3.0 // Inches
)

func SetFontDir(fontDir string) {
	png.SetFontDir(fontDir)
}

func MakeSrcsPngContent(d db.DB, b *bytes.Buffer, rawQuery string) error {
	dTable, err := getDataTable(d, rawQuery)
	if err != nil {
		return err
	}

	q, _ := url.ParseQuery(rawQuery)
	widthStr := q.Get("w")
	width := float64(defaultPngWidth)
	if widthStr != "" {
		var err error
		width, err = strconv.ParseFloat(widthStr, 64)
		if err != nil {
			return errors.New("Cannot parse width parameter w: " + err.Error())
		}
	}

	heightStr := q.Get("h")
	height := float64(defaultPngHeight)
	if heightStr != "" {
		var err error
		height, err = strconv.ParseFloat(heightStr, 64)
		if err != nil {
			return errors.New("Cannot parse height parameter h: " + err.Error())
		}
	}

	// TODO: Don't call this again since getDataTable() already did.
	req, err := requests.MakeRowRangeReqs(rawQuery)
	if err != nil {
		return err
	}

	// TODO: Write ticker functions to work for equalX, sortByColumn, and
	//       sortByConfig.  Use in dataTableToPng().
	xLabel := req.SortByColumn
	if req.EqualX {
		xLabel = common.RecordNumName
	} else if req.SortByConfig != "" {
		xLabel = req.SortByConfig
	}

	var srcs []string
	for _, fs := range req.FilteredSources {
		srcs = append(srcs, fs.Source)
	}
	title := strings.Join(srcs, ", ")

	png.DataTableToPng(b, dTable, title, width, height, xLabel)
	return nil
}

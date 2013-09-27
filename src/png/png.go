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

package png

import (
	"bytes"
	"code.google.com/p/plotinum/plot"
	"code.google.com/p/plotinum/plotter"
	"code.google.com/p/plotinum/vg"
	"code.google.com/p/plotinum/vg/vgimg"
	"fmt"
	"github.com/golang/glog"
	"github.com/google/tsviewdb/src/common"
	"github.com/google/tsviewdb/src/db"
	"github.com/lucasb-eyer/go-colorful"
	"image/color"
	"math"
	"strings"
	"time"
)

func drawPng(b *bytes.Buffer, p *plot.Plot, width, height float64) {
	w, h := vg.Inches(width), vg.Inches(height)
	c := vgimg.PngCanvas{Canvas: vgimg.New(w, h)}
	p.Draw(plot.MakeDrawArea(c))
	c.WriteTo(b)
}

func getHumanTime(epochTime int64, layout string) string {
	t := time.Unix(epochTime/1000, 0)
	return fmt.Sprintf("%v", t.UTC().Format(layout))
}

// getLayout gets a time layout given a diff in milliseconds.
func getLayout(min, max int64) string {
	diff := max - min
	switch {
	case diff > yearMillis:
		return layoutMonths
	case diff > monthMillis:
		return layoutDays
	// case diff > dayMillis:
	// 	return layoutHours
	case diff > hourMillis:
		return layoutMinutes
	default:
		return layoutSeconds
	}
}

func getHumanTimeScaled(epochTime int64, min, max int64) string {
	layout := getLayout(min, max)
	return getHumanTime(epochTime, layout)
}

const (
	hourMillis  = 3600000
	dayMillis   = 86400000
	monthMillis = dayMillis * 30
	yearMillis  = dayMillis * 365

	// Mon Jan 2 15:04:05 -0700 MST 2006
	layoutMonths = "Jan06"
	layoutDays   = "02Jan06"
	// layoutHours   = "02Jan06-15"
	layoutMinutes = "02Jan-15:04"
	layoutSeconds = "15:04:05"
)

// DefaultTicks is suitable for the Tick.Marker field of an Axis,
// it returns a resonable default set of tick marks.
func TimeTicks(min, max float64) (ticks []plot.Tick) {
	const SuggestedTicks = 3
	tens := math.Pow10(int(math.Floor(math.Log10(max - min))))
	n := (max - min) / tens
	for n < SuggestedTicks {
		tens /= 10
		n = (max - min) / tens
	}

	majorMult := int(n / SuggestedTicks)
	switch majorMult {
	case 7:
		majorMult = 6
	case 9:
		majorMult = 8
	}
	majorDelta := float64(majorMult) * tens
	val := math.Floor(min/majorDelta) * majorDelta
	for val <= max {
		if val >= min && val <= max {
			label := getHumanTimeScaled(int64(val), int64(min), int64(max))
			ticks = append(ticks, plot.Tick{Value: val, Label: label})
		}
		val += majorDelta
	}

	minorDelta := majorDelta / 2
	switch majorMult {
	case 3, 6:
		minorDelta = majorDelta / 3
	case 5:
		minorDelta = majorDelta / 5
	}

	val = math.Floor(min/minorDelta) * minorDelta
	for val <= max {
		found := false
		for _, t := range ticks {
			if t.Value == val {
				found = true
			}
		}
		if val >= min && val <= max && !found {
			ticks = append(ticks, plot.Tick{Value: val})
		}
		val += minorDelta
	}
	return
}

func DataTableToPng(b *bytes.Buffer, dt *db.DataTable, title string, width, height float64, xLabel string) error {
	p, err := plot.New()
	if err != nil {
		return err
	}

	p.Title.Text = title
	p.X.Label.Text = xLabel
	// p.X.Label.Text = "Record"
	p.Y.Label.Text = "msec"

	// FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX
	// FIX: need new ticker function to handle equalX (while keeping xLabel as selected)
	if xLabel == common.TimeName {
		p.X.Tick.Marker = TimeTicks
	}
	p.Legend.Top = true

	numColumns := len(dt.ColumnNames)
	lines := make([]plotter.XYs, numColumns-1) // Skip X column.

	for rowNum, dRow := range dt.Data {
		xp := (*dRow)[0]
		if xp != nil {
			for col := 1; col < numColumns; col++ { // Skip X column.
				yp := (*dRow)[col]
				if yp != nil {
					lines[col-1] = append(lines[col-1], struct{ X, Y float64 }{X: *xp, Y: *yp})
					_ = rowNum
					// lines[col-1] = append(lines[col-1], struct{ X, Y float64 }{X: float64(rowNum + 1), Y: *yp})
				}
			}
		}
	}

	colorList := getColors(numColumns - 1) // Skip X column.

	for i, line := range lines {
		columnName := dt.ColumnNames[i+1]
		l, err := plotter.NewLine(line)
		if err != nil {
			return err
		}
		if strings.Index(columnName, common.RegressNamePrefix) == 0 { // If regression value.
			l.LineStyle.Color = color.RGBA{255, 0, 0, 255}
			l.LineStyle.Width = vg.Points(2.0)
		} else {
			l.LineStyle.Color = colorList[i]
			l.LineStyle.Width = vg.Points(1.5)
		}
		// l.LineStyle.Dashes = []vg.Length{vg.Points(5), vg.Points(5)}
		p.Add(l)
		p.Legend.Add(columnName, l)
	}

	tPng := time.Now()
	drawPng(b, p, width, height)
	glog.V(3).Infof("PERF: makePng time: %v", time.Now().Sub(tPng))
	return nil
}

func getColors(n int) []color.Color {
	// From this discussion:
	//   http://martin.ankerl.com/2009/12/09/how-to-create-random-colors-programmatically/
	//
	goldenRatioC := float64(0.618033988749895)
	h := float64(0)

	colorList := make([]color.Color, n)
	s := 0.8
	v := 0.6
	for i := range colorList {
		h += goldenRatioC
		_, h := math.Modf(h)
		c := colorful.Hsv(h*360, s, v)
		colorList[i] = c
	}

	// gen_html {
	//   h += golden_ratio_conjugate
	//   h %= 1
	//   hsv_to_rgb(h, 0.5, 0.95)
	// }

	return colorList
}

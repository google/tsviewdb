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

'use strict';

/**
 * @fileoverview Date slider creation and handling code.  Depends on jQuery,
 * jQuery UI, and jQuery UI dragslider.
 */

// TODO: Make this code automatically create the required HTML & CSS.
/*
 * HTML should have this form:
 *
  <div id="date-control">
    <div id="date-selected-range">
      <input type="text" id="date-from" name="date-from">
      <input type="text" id="date-to" name="date-to">
    </div>

    <div id="date-slider"></div>

    <div id="date-scale-message">
      <input type="button" id="date-scale-requery" value="load relative dates">
    </div>

    <div id="date-scale-buttons">
      <input type="button" id="date-scale-left" value="&larr;">

      <input type="button" id="date-scale-right" value="&rarr;">
    </div>
  </div>

  <div id="date-slider-scale"></div><br>


  CSS should be:

  #date-control {
    margin-left: 10px;
    position: relative;
    width: 920px;
  }

  #date-selected-range {
    font-size: small;
    margin-top: -5px;
    float: left;
    width: 66px;
    height: 30px;
  }

  #date-slider {
    margin-left: 15px;
    height: 10px;
    width: 680px;
    float: left;
  }

  #date-scale-message {
    position: relative;
    font-size: small;
    left: 19px;
    top: -6px;
    float: left;
  }

  #date-scale-buttons {
    position: absolute;
    top: 18px;
    left: 804px;
    width: 90px;
  }

  #date-scale-left,
  #date-scale-right {
    display: inline-block;
    padding: 0px 8px 1px 8px;
  }

  #date-slider-scale {
    position: relative;
    font-size: small;
    margin-top: 7px;
    margin-left: 85px;
    top: 16px;
    height: 10px;
    width: 680px;
  }

  #date-from,
  #date-to {
    width: 56px;
  }

*/



/**
 * This class is used to create and manage a date slider widget which selects
 * date ranges and invokes a callback after range updates.
 * @constructor
 *  * @param {!Params} params A URL parameter storage and manipulation object.
 * @param {!Graphs} graphs A graphs manipulation object.
 * @param {!QueryChange} qchange A query change UI handling widget.
 * @param {jQuerySelector} elemSlider The date range slider div.
 * @param {jQuerySelector} elemSliderButtonL The slider date shift left button.
 * @param {jQuerySelector} elemSliderButtonR The slider date shift right button.
 * @param {jQuerySelector} elemSliderButtonRequery The slider requery button.
 * @param {jQuerySelector} elemSliderScale The slider scale div.
 * @param {jQuerySelector} elemDateFrom The calendar box "from" div.
 * @param {jQuerySelector} elemDateTo The calendar box "to" div.
 * @param {string} sliderHandleClass The class name for the slider handles.
 * @param {string} lightHandleClass The class name for the lighter handle CSS.
 * @param {function(boolean, boolean=)} graphLoadFunction A callback which
 *   performs a graph load.  Should take two booleans:
 *   forceUncached: If true then perform an uncached load and update caches
 *     along the way.
 *   opt_checkAndZoom: If true then update the date range window if this makes
 *     sense for the mode we're in.
 */
// TODO: Change the names of the variables which end up holding jQuery objects
// TODO: to something more representative of that fact.
function QuickDates(params, graphs, qchange, elemSlider, elemSliderButtonL,
    elemSliderButtonR, elemSliderButtonRequery, elemSliderScale, elemDateFrom,
    elemDateTo, sliderHandleClass, lightHandleClass, graphLoadFunction) {
  // TODO: Add docstrings to these member variables with type and visibility
  // TODO: annotations.
  this.params_ = params;
  this.graphs_ = graphs;
  this.qchange_ = qchange;
  // Clamp the date to the start of the day... if past that then use the next
  // day:
  this.currentDate = QuickDates.getDateCeilToDay_(new Date());
  this.currentTime = this.currentDate.getTime();

  // Low and high dates for currently selected dateslider widget settings.
  // Updated from the raw slider position after each position change.
  this.dateLow = null;
  this.dateHigh = null;

  this.scaleUnitOffset = 0;  // Move entire slider scale this many units.

  this.elemSlider = $(elemSlider);
  this.elemSliderButtonL = $(elemSliderButtonL);
  this.elemSliderButtonR = $(elemSliderButtonR);
  this.elemSliderButtonRequery = $(elemSliderButtonRequery);
  this.elemSliderScale = $(elemSliderScale);
  this.elemDateFrom = $(elemDateFrom);
  this.elemDateTo = $(elemDateTo);

  // Used to lighten the right slider handle color when sliding in and out of
  // the "now" position.
  this.sliderHandleClass = sliderHandleClass;
  this.lightHandleClass = lightHandleClass;
  this.rightSliderHandleElem = null;  // Set after sliderWidget creation.

  this.graphLoadFunction = graphLoadFunction;

  this.daysOfData = QuickDates.DEFAULT_SCALE_UNIT_RANGE;
  if (this.params_.frag.daysOfData) {
    this.daysOfData = parseInt(this.params_.frag.daysOfData);
  }
}


/**
 * Number of seconds in one scale unit.
 * @type {number}
 * @const
 */
QuickDates.SCALEUNITSECONDS = 86400;


/**
 * Number of scale units across the entire slider.  Must be evenly divisible by
 * SCALE_MOVE_AMOUNT.
 * @type {number}
 * @const
 */
QuickDates.MAX_SCALE_UNIT_RANGE = 180;


/**
 * The range in scale units when no range is selected.
 * @type {number}
 * @const
 */
QuickDates.DEFAULT_SCALE_UNIT_RANGE = 10;


/**
 * The minimum allowed range difference (in scale units).
 * @type {number}
 * @const
 */
QuickDates.MIN_RANGE = 0;


/**
 * The maximum allowed range difference (in scale units).
 * @type {number}
 * @const
 */
QuickDates.MAX_RANGE = 180;


/**
 * Number of scale units to adjust scale when using arrows.
 * @type {number}
 * @const
 */
QuickDates.SCALE_MOVE_AMOUNT = 30;


/**
 * Number of seconds in a day.  This is used for date calculations, and is
 * independent of SCALEUNITSECONDS.
 * @type {number}
 * @const
 */
QuickDates.DAYSECONDS = 86400;


/**
 * Format for dates shown in calendar popup from and to text boxes.
 * @type {string}
 * @const
 */
QuickDates.DATE_FORMAT = 'mm/dd/y';


/**
 * Return a new date object at an offset from a given one.
 * @param {!Date} d The input date.
 * @param {number} days Number of days to use.
 * @return {Object} A new date object "days" later than the given one.
 * @private
 */
QuickDates.getNewDatePlusDays_ = function(d, days) {
  var plusDate = new Date(d);
  plusDate.setDate(d.getDate() + days);
  return plusDate;
};


/**
 * Given two dates, return the older one.
 * @param {!Date} d1 A date object to compare.
 * @param {!Date} d2 A date object to compare.
 * @return {!Date} The older date object.
 * @private
 */
QuickDates.minimumDate_ = function(d1, d2) {
  return (d1.getTime() < d2.getTime()) ? d1 : d2;
};


/**
 * Return a date object for the oldest associated epoch day.
 * @param {!Date} d The input date.
 * @return {!Date} A date object set with the oldest associated epoch day from
 *   the time of the given date object.
 * @private
 */
QuickDates.getDateCeilToDay_ = function(d) {
  var epochDays = d.getTime() / (QuickDates.SCALEUNITSECONDS * 1000);
  return new Date(Math.ceil(epochDays) * QuickDates.SCALEUNITSECONDS * 1000);
};


/**
 * Get the local timezone difference in millis from UTC.
 * @param {!Date} d The input date.
 * @return {number} The timezone offset from UTC in millis.
 * @private
 */
QuickDates.getTZOffset_ = function(d) {
  return (24 - d.getTimezoneOffset() / 60) * 60 * 60 * 1000;
};


/**
 * Normalize time to UTC.
 * @param {!Date} d A date object which is *modified* by this function,
 *   and converted to UTC time.
 * @private
 */
QuickDates.addTZOffset_ = function(d) {
  d.setTime(d.getTime() + QuickDates.getTZOffset_(d));
};


/**
 * Create the date slider and date picker widgets.
 */
QuickDates.prototype.createWidget = function() {
  this.createWidgetDateSlider_(this.params_.frag.startDate,
      this.params_.frag.endDate);
  this.createWidgetDatePicker_();
};


/**
 * Create the date slider widget.
 * @param {string|undefined} startDate The initial start date to use.  In the
 *   format: YYYMMDD.  If startDate or endDate are not set then we use the
 *   default setting, which is DEFAULT_SCALE_UNIT_RANGE * SCALEUNITSECONDS
 *   seconds from the current time.
 * @param {string|undefined} endDate The initial end date to use.  In the
 *   format: YYYMMDD.  If startDate or endDate are not set then we use the
 *   default setting, which is DEFAULT_SCALE_UNIT_RANGE * SCALEUNITSECONDS
 *   seconds from the current time.
 * @private
 */
QuickDates.prototype.createWidgetDateSlider_ = function(startDate, endDate) {
  var self = this;

  this.elemSlider.dragslider({
    range: true,
    rangeDrag: true,
    // Since min/max is zero-based the actual number of scale units is
    // MAX_SCALE_UNIT_RANGE + 1.  This makes space for the "now" position.
    min: 0,
    max: QuickDates.MAX_SCALE_UNIT_RANGE,
    step: 1,
    values: [QuickDates.MAX_SCALE_UNIT_RANGE - this.daysOfData + 1,
             QuickDates.MAX_SCALE_UNIT_RANGE],
    create: function(event, ui) {
      if (startDate && endDate) {  // If query string dates are supplied.
        var startDateObj = QuickDates.getDateFromQSDate_(startDate);
        QuickDates.addTZOffset_(startDateObj);
        var endDateObj = QuickDates.getDateFromQSDate_(endDate);
        QuickDates.addTZOffset_(endDateObj);
        self.setDateSlider_(startDateObj, endDateObj);
      } else {
        self.setDateSliderRaw_(QuickDates.MAX_SCALE_UNIT_RANGE -
            self.daysOfData + 1,
            QuickDates.MAX_SCALE_UNIT_RANGE, true);
      }
    },
    slide: function(event, ui) {
      return self.setDateSliderRaw_(ui.values[0], ui.values[1]);},
    stop: function(event, ui) {self.updateURLFromSlider_(false);}
  });

  this.elemSliderButtonL.on('click', function(event) {
    var values = self.elemSlider.dragslider('values');
    if (values[1] == QuickDates.MAX_SCALE_UNIT_RANGE) {
      return;
    }
    self.scaleUnitOffset = self.scaleUnitOffset - QuickDates.SCALE_MOVE_AMOUNT;
    self.setDateSliderRaw_(values[0], values[1]);
    self.updateURLFromSlider_(false);
  });

  this.elemSliderButtonR.on('click', function(event) {
    var values = self.elemSlider.dragslider('values');
    if (values[1] == QuickDates.MAX_SCALE_UNIT_RANGE) {
      return;
    }
    self.scaleUnitOffset = self.scaleUnitOffset + QuickDates.SCALE_MOVE_AMOUNT;
    if (self.scaleUnitOffset > 0) {
      self.scaleUnitOffset = 0;
    } else {
      self.setDateSliderRaw_(values[0], values[1]);
      self.updateURLFromSlider_(false);
    }
  });

  this.elemSliderButtonRequery.on('click', function(event) {
    var forceUncachedLoad = event.ctrlKey || event.altKey || event.shiftKey;
    self.updateURLFromSlider_(forceUncachedLoad);
  });
};


/**
 * Create the calendar-based date selector widget.
 * @private
 */
QuickDates.prototype.createWidgetDatePicker_ = function() {
  var self = this;

  this.elemDateFrom.datepicker({
    defaultDate: this.dateLow,
    minDate:
        QuickDates.getNewDatePlusDays_(this.dateHigh,
            -QuickDates.MAX_RANGE + 1),
    maxDate:
        QuickDates.getNewDatePlusDays_(this.dateHigh, -QuickDates.MIN_RANGE),
    dateFormat: QuickDates.DATE_FORMAT,
    numberOfMonths: 4,
    duration: 0,  // Make it pop-up immediately.
    beforeShow: function(inputText, inst) {
      self.elemDateFrom.datepicker('option', {
        defaultDate: self.dateLow,
        minDate:
            QuickDates.getNewDatePlusDays_(self.dateHigh,
                -QuickDates.MAX_RANGE + 1),
        maxDate:
            QuickDates.getNewDatePlusDays_(self.dateHigh, -QuickDates.MIN_RANGE)
      });
    },
    onSelect: function(dateText, inst) {
      var newDateLow = new Date(dateText);
      QuickDates.addTZOffset_(newDateLow);
      self.setDateSlider_(newDateLow, self.dateHigh);
      self.updateURLFromSlider_(false);
    }
  });

  this.elemDateTo.datepicker({
    defaultDate: this.dateHigh,
    minDate: QuickDates.getNewDatePlusDays_(this.dateLow, QuickDates.MIN_RANGE),
    maxDate: QuickDates.minimumDate_(this.currentDate,
        QuickDates.getNewDatePlusDays_(this.dateLow, QuickDates.MAX_RANGE - 1)),
    dateFormat: QuickDates.DATE_FORMAT,
    numberOfMonths: 4,
    duration: 0,  // Make it pop-up immediately.
    beforeShow: function(inputText, inst) {
      self.elemDateTo.datepicker('option', {
        defaultDate: self.dateHigh,
        minDate:
            QuickDates.getNewDatePlusDays_(self.dateLow, QuickDates.MIN_RANGE),
        maxDate: QuickDates.minimumDate_(self.currentDate,
            QuickDates.getNewDatePlusDays_(self.dateLow, QuickDates.MAX_RANGE -
            1))
      });
    },
    onSelect: function(dateText, inst) {
      var newDateHigh = new Date(dateText);
      QuickDates.addTZOffset_(newDateHigh);
      self.setDateSlider_(self.dateLow, newDateHigh);
      self.updateURLFromSlider_(false);
    },
    onClose: function(dateText, inst) {
      // Keep "now" in the text box if it was already there and nothing was
      // selected.
      var values = self.elemSlider.dragslider('values');
      var useDaysOfData = values[1] == QuickDates.MAX_SCALE_UNIT_RANGE;
      if ((useDaysOfData) && (dateText == '')) {
        self.elemDateTo.val('now');
      }
    }
  });

};


/**
 * Update the URL with the slider position.
 * @param {boolean} forceUncachedLoad If true, force an uncached load.
 * @private
 */
QuickDates.prototype.updateURLFromSlider_ = function(forceUncachedLoad) {
  var values = this.elemSlider.dragslider('values');
  var difference = values[1] - values[0] + 1;
  var useDaysOfData = values[1] == QuickDates.MAX_SCALE_UNIT_RANGE;
  this.params_.updateDateRangeFragQS(useDaysOfData,
      QuickDates.formattedDateForQS_(this.dateLow),
      QuickDates.formattedDateForQS_(this.dateHigh), difference);
  if (this.qchange_.lastPts === 0) {
    this.graphLoadFunction(forceUncachedLoad, true);
  }
};


/**
 * @param {Object} dateL The low date object.
 * @param {Object} dateH The high date object.
 * @return {{low: number, high: number, scaleUnitOffset: number}} These
 *   properties are as follows:
 *     low: The low slider position.
 *     high: The high slider position.
 *     scaleUnitOffset: The amount the slider scale is shifted in days.
 * @private
 */
QuickDates.prototype.computeSliderPosFromDates_ = function(dateL, dateH) {
  // Update date range sliders from startDate and endDate QS parameters.
  // These are offsets in days from the current date:
  var startOffset = ((this.currentDate - dateL) /
          (1000 * QuickDates.DAYSECONDS)) + 1;
  var endOffset = ((this.currentDate - dateH) /
          (1000 * QuickDates.DAYSECONDS)) + 1;

  // startOffset will always be furthest away from "now."
  var startDateScale = startOffset / QuickDates.MAX_SCALE_UNIT_RANGE;
  var offset = 0;
  if (startDateScale > 1) {
    var startOffsetScaled = startOffset - QuickDates.MAX_SCALE_UNIT_RANGE;
    var localOffset = Math.ceil(startOffsetScaled /
        QuickDates.SCALE_MOVE_AMOUNT);
    offset = - localOffset * QuickDates.SCALE_MOVE_AMOUNT;
  }

  var high = QuickDates.MAX_SCALE_UNIT_RANGE - endOffset - offset;
  var low = QuickDates.MAX_SCALE_UNIT_RANGE - startOffset - offset;

  return {'low': low, 'high': high, 'scaleUnitOffset': offset};
};


/**
 * Set the date slider position.  Called when the slider is first created, or
 * when any dates are manually picked using the calendar date widgets.
 * @param {Object} dateL The low date object.
 * @param {Object} dateH The high date object.
 * @private
 */
QuickDates.prototype.setDateSlider_ = function(dateL, dateH) {
  var returnObj = this.computeSliderPosFromDates_(dateL, dateH);
  var low = returnObj.low;
  var high = returnObj.high;
  this.scaleUnitOffset = returnObj.scaleUnitOffset;
  this.elemSlider.dragslider('values', [low, high]);
  this.setDateSliderRaw_(low, high, true);
};


/**
 * The base routine called when the date slider position changes.  This updates
 * the graph (if needed), the date boxes, and enforces the min and max ranges.
 * Returns false to prevent a slide.
 * @param {number} valueLow The low slider value.
 * @param {number} valueHigh The high slider value.
 * @param {boolean=} opt_noZoom If true don't even check to see if we
 *   should rezoom the graph to date window, let alone do it.
 * @return {boolean} Returns false to prevent a slide.
 * @private
 */
QuickDates.prototype.setDateSliderRaw_ = function(valueLow, valueHigh,
    opt_noZoom) {
  var difference = valueHigh - valueLow + 1;
  if ((difference > QuickDates.MAX_RANGE) ||
          (difference < QuickDates.MIN_RANGE)) {
    return false;
  }

  if (this.rightSliderHandleElem === null) {
    this.rightSliderHandleElem = this.elemSlider.find(
        '.' + this.sliderHandleClass).eq(1);
  }

  if (valueHigh == QuickDates.MAX_SCALE_UNIT_RANGE) {
    this.rightSliderHandleElem.addClass(this.lightHandleClass);
    this.scaleUnitOffset = -1;
    this.paintDates_();
    this.elemSliderButtonL.hide();
    this.elemSliderButtonR.hide();
    this.elemSliderButtonRequery.attr('value',
        'load last ' + difference + ' days');
  } else {
    if (this.scaleUnitOffset == -1) {
      this.scaleUnitOffset = 0;
    }
    this.paintDates_();
    this.rightSliderHandleElem.removeClass(this.lightHandleClass);
    this.elemSliderButtonL.show();
    this.elemSliderButtonR.show();
    this.elemSliderButtonRequery.attr('value',
        'load exact ' + difference + ' days');
  }
  var low = valueLow - QuickDates.MAX_SCALE_UNIT_RANGE + 1;
  var high = valueHigh - QuickDates.MAX_SCALE_UNIT_RANGE + 1;
  this.updateDateRangeFromSliders_(low, high);
  if (!opt_noZoom) {
    this.zoomGraphToDateWindow_();
  }
  return true;
};


/**
 * Update the date boxes from the slider range.
 * @param {number} low The low slider range.
 * @param {number} high The high slider range.
 * @private
 */
QuickDates.prototype.updateDateRangeFromSliders_ = function(low, high) {
  var nowSelected = high == 1;
  this.dateLow = new Date(this.currentTime + (low + this.scaleUnitOffset) *
      QuickDates.SCALEUNITSECONDS * 1000);
  var dateLowStr = QuickDates.formattedDate_(this.dateLow);
  this.dateHigh = new Date(this.currentTime + (high + this.scaleUnitOffset) *
      QuickDates.SCALEUNITSECONDS * 1000);
  var dateHighStr = nowSelected ?
      'now' : QuickDates.formattedDate_(this.dateHigh);
  this.elemDateFrom.val(dateLowStr);
  this.elemDateTo.val(dateHighStr);
};


/**
 * Return the current date window specified by the sliders.  This is used to
 * initialize the graph zoom range after new data loads when we're in linear-
 * time mode (X-axis in variably-spaced units of time).
 * @return {Array.<number>} A low and high window range.  Returns null if we're
 *   not in linear time mode.
 */
QuickDates.prototype.getDateWindow = function() {
  // Get graph zoom window for the selected time range but only if we're in the
  // linear time mode, because for any other mode we won't know how to zoom.
  if (this.graphs_.timeXScale && !this.graphs_.equalXScale &&
          !this.qchange_.lastPts &&
          this.graphs_.aggData !== null) {  // Ensure initial data loaded.
    var windowLow = this.dateLow.getTime() - (QuickDates.DAYSECONDS * 1000);
    var windowHigh = this.dateHigh.getTime();
    return [windowLow, windowHigh];
  }
  return null;
};


/**
 * Zoom the graph to the current slider range if we're in linear time mode.
 * @private
 */
QuickDates.prototype.zoomGraphToDateWindow_ = function() {
  var dateWindow = this.getDateWindow();
  this.graphs_.aggChart.updateOptions({dateWindow: dateWindow});
};


/**
 * Draw the dates under the slider.
 * @private
 */
QuickDates.prototype.paintDates_ = function() {
  var months = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep',
                'Oct', 'Nov', 'Dec'];
  var newHtml = '';
  var scaleLength = this.elemSliderScale.width();
  var boxWidth = 80;  // Text box width.
  var sliderHandleWidth = this.rightSliderHandleElem.width();
  var scaleOffsetFrac = (boxWidth / 2 - sliderHandleWidth / 2) / scaleLength;
  var lastPointToDraw = QuickDates.MAX_SCALE_UNIT_RANGE;
  // If we're going to draw "now":
  if ((this.scaleUnitOffset == -1) || (this.scaleUnitOffset == 0)) {
    lastPointToDraw = lastPointToDraw -
        Math.round(QuickDates.MAX_SCALE_UNIT_RANGE * 0.04);
  }
  for (var i = 0; i < lastPointToDraw; i++) {
    var newDate = new Date(this.currentTime +
        (i + this.scaleUnitOffset - (QuickDates.MAX_SCALE_UNIT_RANGE - 1)) *
        QuickDates.SCALEUNITSECONDS * 1000);
    var month = months[newDate.getMonth()];
    var day = newDate.getDate();
    var year = (newDate.getFullYear() + '').substr(2, 2);
    if (day == 1) {
      newHtml = newHtml +
          '<div style="position:absolute; text-align: center; z-index:2; ' +
          'width: ' + boxWidth + 'px; top:4px; left:' + (i /
              (QuickDates.MAX_SCALE_UNIT_RANGE) - scaleOffsetFrac) * 100.0 +
          '%;">' + month + year + '</div>';
    }
  }
  if ((this.scaleUnitOffset == -1) || (this.scaleUnitOffset == 0)) {
    newHtml = newHtml +
        '<div style="position:absolute; text-align: center; z-index:2; ' +
        'width: ' + boxWidth + 'px; top:4px; left:' + (100 - scaleOffsetFrac *
            100) + '%;">now</div>';
  }
  this.elemSliderScale.html(newHtml);
};


/**
 * Return a formatted string representation of a Date() for browser display.
 * @param {!Date} d The input date.
 * @return {string} A formatted string representation of the given date.
 * @private
 */
QuickDates.formattedDate_ = function(d) {
  var month = Math.round(d.getMonth()) + 1;
  month = (month < 10) ? ('0' + month) : month;
  var day = d.getDate();
  day = (day < 10) ? ('0' + day) : day;
  var year = (d.getFullYear() + '').substr(2, 2);
  return month + '/' + day + '/' + year;
};


/**
 * Return a formatted string representation of a Date() for startDate and
 * endDate query string parameters.
 * @param {!Date} d The input date.
 * @return {string} A formatted string representation of the given date.
 * @private
 */
QuickDates.formattedDateForQS_ = function(d) {
  var month = Math.round(d.getMonth()) + 1;
  month = (month < 10) ? ('0' + month) : month;
  var day = d.getDate();
  day = (day < 10) ? ('0' + day) : day;
  var year = d.getFullYear() + '';  // To string to cause downstream typecasts.
  return year + month + day;
};


/**
 * Parse a string of the form YYYYMMDD and return a Date().
 * @param {string} qsDate A date in string form (see above).
 * @return {!Date} The output date.
 * @private
 */
QuickDates.getDateFromQSDate_ = function(qsDate) {
  var year = qsDate.substr(0, 4);
  var month = qsDate.substr(4, 2);
  var day = qsDate.substr(6, 2);
  var dateStr = month + '/' + day + '/' + year;
  return new Date(dateStr);
};

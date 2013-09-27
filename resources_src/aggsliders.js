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
 * @fileoverview Implements aggregate value selection slider widgets and
 * accompanying functionality.  This includes sliders to select the line and
 * shadow plot aggregates (max, mean, min, etc.), as well as functionality to
 * update the state in the URL.  Depends on jQuery and jQuery UI.
 */



/**
 * Aggregate slider creation and manipulation object.
 * @param {Params} params A URL parameter storage and manipulation object.
 * @param {graphs} graphs A graphs manipulation object.
 * @param {number} startLineIndex Line select start index (between 0 and 11).
 * @param {Array.<number>} startShadowIndices Shadow select start indices
 *   (between 0 and 11).
 * @constructor
 */
function AggSliders(params, graphs, startLineIndex, startShadowIndices) {
  this.params_ = params;
  this.graphs_ = graphs;

  // Default aggregate graph indices.
  this.lineSelect = startLineIndex;
  // Shadow select indexes: [lower, upper]:
  this.shadowSelect = startShadowIndices;

  if (params.frag.aggregates) {
    var tmp = params.frag.aggregates.split(',');
    if (tmp.length == 3) {
      // Determine if we have old-style or new-style agg selectors.
      if (isNaN(tmp[0])) {  // New style if not a number.
        this.lineSelect = AggSliders.MAP_PARAM[tmp[0]];
        this.shadowSelect[0] = AggSliders.MAP_PARAM[tmp[1]];
        this.shadowSelect[1] = AggSliders.MAP_PARAM[tmp[2]];
      } else {  // Old style.
        this.lineSelect = AggSliders.MAP_DEPRECATED[tmp[0]];
        this.shadowSelect[0] = AggSliders.MAP_DEPRECATED[tmp[1]];
        this.shadowSelect[1] = AggSliders.MAP_DEPRECATED[tmp[2]];
      }
    }
  }

  this.elemAllTextDivs = null;
  this.selectedLineClassName = null;
  this.selectedShadowClassName = null;
}


/**
 * The number of aggregate arrays in the data received from the server.  The
 * indices are:
 * [0]  = min
 * [1]  = 1st %-tile
 * [2]  = 5th %-tile
 * [3]  = 10th %-tile
 * [4]  = 25th %-tile
 * [5]  = 50th %-tile
 * [6]  = mean
 * [7]  = 75th %-tile
 * [8]  = 90th %-tile
 * [9]  = 95th %-tile
 * [10] = 99th %-tile
 * [11] = max
 */
AggSliders.NUM_AGGREGATES = 12;


/**
 * Deprecated to current map.
 */
AggSliders.MAP_DEPRECATED = {
        '0': 0,   // min
        '1': 2,   // p5
        '2': 3,   // p10
        '3': 4,   // DEPRECATED: p20 --> p25
        '4': 4,   // DEPRECATED: -95CI --> p25
        '5': 5,   // p50
        '6': 6,   // mean --> null (flag to enable mean)
        '7': 7,   // DEPRECATED: +95CI --> p75
        '8': 7,   // DEPRECATED: p80 --> p75
        '9': 8,   // p90
        '10': 9,  // p95
        '11': 11  // max
};


/**
 * Map of string params to indexes used internally.
 */
AggSliders.MAP_PARAM = {
        'min': 0,
        'p1': 1,
        'p5': 2,
        'p10': 3,
        'p25': 4,
        'p50': 5,
        'mean': 6,
        'p75': 7,
        'p90': 8,
        'p95': 9,
        'p99': 10,
        'max': 11
};


/**
 * Reverse of AggSliders.MAP_PARAM.
 */
AggSliders.MAP_INDEX = {
        '0': 'min',
        '1': 'p1',
        '2': 'p5',
        '3': 'p10',
        '4': 'p25',
        '5': 'p50',
        '6': 'mean',
        '7': 'p75',
        '8': 'p90',
        '9': 'p95',
        '10': 'p99',
        '11': 'max'
};


/**
 * Default line slider index.
 */
AggSliders.DEFAULT_LINE_INDEX = 6;  // Mean.


/**
 * Default shadow slider index.
 */
AggSliders.DEFAULT_SHADOW_INDICES = [2, 9];  // 5%-tile to 95%-tile


/**
 * Create aggregate sliders.
 * @param {jQuerySelector} lineSliderElem The line select slider element.
 * @param {jQuerySelector} shadowSliderElem The slider element div.
 * @param {jQuerySelector} sliderTextElem The slider scale text box.
 * @param {string} selectedLineClassName The class for tagging selected elements
 *   in the line slider.
 * @param {string} selectedShadowClassName The class name for tagging selected
 *   elements in the shadow slider.
 */
AggSliders.prototype.createSliders = function(lineSliderElem, shadowSliderElem,
    sliderTextElem, selectedLineClassName, selectedShadowClassName) {
  this.elemAllTextDivs = $(sliderTextElem).find('div');
  this.selectedLineClassName = selectedLineClassName;
  this.selectedShadowClassName = selectedShadowClassName;

  var self = this;

  $(lineSliderElem).slider({
    orientation: 'vertical',
    range: false,
    min: 0,
    max: AggSliders.NUM_AGGREGATES - 1,
    value: this.lineSelect,
    step: 1,
    create: function(event, ui) {
      self.updateLineSliderData(self.lineSelect);
    },
    slide: function(event, ui) {
      self.updateLineSliderData(ui.value);
    },
    stop: function() {self.updateAggSelectFragment();}
  });

  $(shadowSliderElem).slider({
    orientation: 'vertical',
    range: true,
    min: 0,
    max: AggSliders.NUM_AGGREGATES - 1,
    values: this.shadowSelect,
    step: 1,
    create: function(event, ui) {
      self.updateShadowSliderData(self.shadowSelect);
    },
    slide: function(event, ui) {
      self.updateShadowSliderData(ui.values);
    },
    stop: function() {self.updateAggSelectFragment();}
  });
};


/**
 * Updates the URL with the current slider position.  Called when the sliders
 * stop moving.
 */
AggSliders.prototype.updateAggSelectFragment = function() {
  var lineVal = AggSliders.MAP_INDEX[this.lineSelect];
  var shadowLowVal = AggSliders.MAP_INDEX[this.shadowSelect[0]];
  var shadowHighVal = AggSliders.MAP_INDEX[this.shadowSelect[1]];
  this.params_.updateFragment('aggregates', lineVal + ',' + shadowLowVal + ',' +
      shadowHighVal);
};


/**
 * Update the aggregate graph data for the line slider and shade the selected
 * slider text.
 * @param {number} index The currently selected line slider index.
 */
AggSliders.prototype.updateLineSliderData = function(index) {
  this.lineSelect = index;
  this.elemAllTextDivs.removeClass(this.selectedLineClassName);
  $(this.elemAllTextDivs[AggSliders.NUM_AGGREGATES - 1 - index]).addClass(
      this.selectedLineClassName);
  this.graphs_.updateAggregateGraph(this.lineSelect, this.shadowSelect);
};


/**
 * Update the aggregate graph data for the line slider and shade the selected
 * slider text.
 * @param {Array.<number>} values Value of the two currently selected shadow
 *   slider indices.
 */
AggSliders.prototype.updateShadowSliderData = function(values) {
  this.shadowSelect = values;
  this.elemAllTextDivs.removeClass(this.selectedShadowClassName);
  var upper = (AggSliders.NUM_AGGREGATES - 1) - values[0];
  var lower = (AggSliders.NUM_AGGREGATES - 1) - values[1];
  for (var i = lower; i <= upper; i++) {
    $(this.elemAllTextDivs[i]).addClass(this.selectedShadowClassName);
  }
  this.graphs_.updateAggregateGraph(this.lineSelect, this.shadowSelect);
};

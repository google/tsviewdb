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
 * @fileoverview Handlers for all the scale-change buttons (X-scale time/CL,
 *   Y-scale linear/log, etc.).  Depends on jQuery.
 */



/**
 * @constructor
 * @param {Params} params A URL parameter storage and manipulation object.
 * @param {Graphs} graphs A graphs manipulation object.
 * @param {boolean} timeXScale If True X-axis is in units of time, otherwise
 *   in units of CL.  This is the intial value.
 * @param {boolean} equalXScale If True X-axis is equally spaced, otherwise
 *   variably spaced.
 * @param {String} metricXScale The metric to use for the X-axis when timeXScale
 *   is false.
 * @param {jQuerySelector} elemLogScaleButton The first button in the log-scale
 *   button pair.
 * @param {jQuerySelector} elemTimeXScaleButton The first button in the time/CL
 *   button pair.
 * @param {jQuerySelector} elemEqualXScaleButton The first button in the
 *   equal/variable button pair.
 * @param {jQuerySelector} elemLogScaleButtonOpposite The second button in the
 *   log-scale button pair.
 * @param {jQuerySelector} elemTimeXScaleButtonOpposite The second button in the
 *   time/CL button pair.
 * @param {jQuerySelector} elemEqualXScaleButtonOpposite The second button in
 *   the equal/variable pair.
 * @param {jQuerySelector} elemTimeXScaleButtonCl Button used to reset the
 *   X-scale to units of CL.
 */
function Scales(params, graphs, timeXScale, equalXScale, metricXScale,
    elemLogScaleButton, elemTimeXScaleButton, elemEqualXScaleButton,
    elemLogScaleButtonOpposite, elemTimeXScaleButtonOpposite,
    elemEqualXScaleButtonOpposite, elemTimeXScaleButtonCl) {
  this.params_ = params;
  this.graphs_ = graphs;

  this.timeXScale = timeXScale;
  this.equalXScale = equalXScale;
  this.metricXScale = metricXScale;

  this.elemLogScaleButton = $(elemLogScaleButton);
  this.elemTimeXScaleButton = $(elemTimeXScaleButton);
  this.elemEqualXScaleButton = $(elemEqualXScaleButton);
  this.elemLogScaleButtonOpposite = $(elemLogScaleButtonOpposite);
  this.elemTimeXScaleButtonOpposite = $(elemTimeXScaleButtonOpposite);
  this.elemEqualXScaleButtonOpposite = $(elemEqualXScaleButtonOpposite);
  this.elemTimeXScaleButtonCl = $(elemTimeXScaleButtonCl);

  this.logYScale = this.params_.frag.logscale == '1';
  this.graphLoadFunction_ = null;
}


/**
 * Set button value based on timeXScale value.
 * @private
 */
Scales.prototype.applyTimeXScaleButtonChange_ = function() {
  if (this.timeXScale) {
    this.elemTimeXScaleButton.addClass('button-pressed');
    this.elemTimeXScaleButtonOpposite.removeClass('button-pressed');
  } else {
    this.elemTimeXScaleButton.removeClass('button-pressed');
    this.elemTimeXScaleButtonOpposite.addClass('button-pressed');
  }
};


/**
 * Apply new metricXScale value.
 * @param {String} newMetricXScale The new metricXScale value to apply.
 */
Scales.prototype.applyNewMetricXScale = function(newMetricXScale) {
  this.metricXScale = newMetricXScale;
  this.elemTimeXScaleButtonOpposite.val(this.metricXScale);
  this.timeXScale = false;
  this.params_.addToFragQS('time_xscale', 0);
  this.params_.addToFragQS('metric_xscale', this.metricXScale);
  this.applyTimeXScaleButtonChange_();
  this.elemTimeXScaleButtonCl.show();
  this.graphLoadFunction_(false, true);
};


/**
 * Set button value based on equalXScale value.
 * @private
 */
Scales.prototype.applyEqualXScaleButtonChange_ = function() {
  if (this.equalXScale) {
    this.elemEqualXScaleButton.removeClass('button-pressed');
    this.elemEqualXScaleButtonOpposite.addClass('button-pressed');
  } else {
    this.elemEqualXScaleButton.addClass('button-pressed');
    this.elemEqualXScaleButtonOpposite.removeClass('button-pressed');
  }
};


/**
 * Change logscale for buttons, graph, and URL, based on logYScale value.
 * @private
 * @param {boolean} noUpdate If true then don't repaint the graph.
 */
Scales.prototype.applyYScaleChange_ = function(noUpdate) {
  if (this.logYScale) {
    this.elemLogScaleButton.removeClass('button-pressed');
    this.elemLogScaleButtonOpposite.addClass('button-pressed');
    if (this.graphs_.aggData && !noUpdate) {
      this.graphs_.aggChart.updateOptions({includeZero: false, logscale: true});
    }
    this.params_.updateFragment('logscale', '1');
  } else {
    this.elemLogScaleButton.addClass('button-pressed');
    this.elemLogScaleButtonOpposite.removeClass('button-pressed');
    if (this.graphs_.aggData && !noUpdate) {
      this.graphs_.aggChart.updateOptions({includeZero: true, logscale: false});
    }
    this.params_.deleteFromFragment('logscale', true);
  }
};


/**
 * Update all button state.  Called on initial UI load.
 */
Scales.prototype.updateButtons = function() {
  if (this.metricXScale == Constants.CL_METRIC_X_AXIS_TOKEN) {
    this.elemTimeXScaleButtonCl.hide();
    this.elemTimeXScaleButtonOpposite.val('CL');
  } else {
    this.elemTimeXScaleButtonCl.show();
    this.elemTimeXScaleButtonOpposite.val(this.metricXScale);
  }
  // Don't update options since we've already done so in the load update:
  this.applyYScaleChange_(true);
  this.applyTimeXScaleButtonChange_();
  this.applyEqualXScaleButtonChange_();
};


/**
 * Register handlers for all buttons.
 * @param {Function} graphLoadFunction A callback which performs a graph load.
 *   Should take two booleans: forceUncached, opt_checkAndZoom.
 *   forceUncached: If True then perform an uncached load and update caches
 *     along the way.
 *   opt_checkAndZoom: If true then update the date range window if this makes
 *     sense for the mode we're in.
 */
Scales.prototype.registerHandlers = function(graphLoadFunction) {
  this.graphLoadFunction_ = graphLoadFunction;
  var self = this;
  var logHandler = function(event) {
    event.preventDefault();
    self.logYScale = !self.logYScale;
    self.applyYScaleChange_();
  };
  var timeXHandler = function(event) {
    event.preventDefault();
    self.timeXScale = !self.timeXScale;
    self.params_.addToFragQS('time_xscale', self.timeXScale ? 1 : 0);
    self.applyTimeXScaleButtonChange_();
    graphLoadFunction(false, true);
  };
  var timeXClHandler = function(event) {
    event.preventDefault();
    self.elemTimeXScaleButtonCl.hide();
    self.timeXScale = false;
    self.elemTimeXScaleButtonOpposite.val('CL');
    self.metricXScale = Constants.CL_METRIC_X_AXIS_TOKEN;
    self.params_.addToFragQS('time_xscale', 0);
    self.params_.addToFragQS('metric_xscale', Constants.CL_METRIC_X_AXIS_TOKEN);
    self.applyTimeXScaleButtonChange_();
    graphLoadFunction(false, true);
  };
  var equalXHandler = function(event) {
    event.preventDefault();
    self.equalXScale = !self.equalXScale;
    self.params_.addToFragQS('equal_xscale', self.equalXScale ? 1 : 0);
    self.applyEqualXScaleButtonChange_();
    graphLoadFunction(false, true);
  };

  this.elemLogScaleButton.on('click', logHandler);
  this.elemLogScaleButtonOpposite.on('click', logHandler);

  this.elemTimeXScaleButton.on('click', timeXHandler);
  this.elemTimeXScaleButtonOpposite.on('click', timeXHandler);
  this.elemTimeXScaleButtonCl.on('click', timeXClHandler);

  this.elemEqualXScaleButton.on('click', equalXHandler);
  this.elemEqualXScaleButtonOpposite.on('click', equalXHandler);
};

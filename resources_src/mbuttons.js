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
 * @fileoverview Create and handle metric buttons.  Depends on jQuery.
 */



/**
 * @constructor
 * @param {Graphs} graphs A graphs manipulation object.
 * @param {Params} params A URL parameter storage and manipulation object.
 * @param {AggSliders} aggsliders An aggregate slider widget object.
 * @param {Scales} scales A scale button handler object.
 * @param {jQuerySelector} stackButtonElem First button in the stacked/separated
 *   button pair.
 * @param {jQuerySelector} stackButtonOppositeElem Second button in the
 *   stacked/separated button pair.
 * @param {jQuerySelector} shadowSliderElem The shadow slider element.  This is
 *   greyed-out when we select stacked graphs.
 * @param {jQuerySelector} metricSearchInputBoxElem The metric search box.
 * @param {string} buttonPressedClassName The class name associated with pressed
 *   button state.
 * @param {string} greyedOutClassName The class name associated with the
 *   greyed-out state.
 * @param {string} metricSelectorIdPrefix The id prefix for the metric buttons.
 */
function MButtons(graphs, params, aggsliders, scales, stackButtonElem,
    stackButtonOppositeElem, shadowSliderElem, metricSearchInputBoxElem,
    buttonPressedClassName, greyedOutClassName, metricSelectorIdPrefix) {
  this.graphs_ = graphs;
  this.params_ = params;
  this.aggsliders_ = aggsliders;  // For lineSelect and shadowSelect.
  this.scales_ = scales;  // For metricXScale.
  this.checkAllButtonElem = null;
  this.stackButtonElem = $(stackButtonElem);
  this.stackButtonOppositeElem = $(stackButtonOppositeElem);
  this.shadowSliderElem = $(shadowSliderElem);
  this.metricSearchInputBoxElem = $(metricSearchInputBoxElem);
  this.buttonPressedClassName = buttonPressedClassName;
  this.greyedOutClassName = greyedOutClassName;
  this.metricSelectorIdPrefix = metricSelectorIdPrefix;
}


/**
 * The time in milliseconds under which two simultaneous clicks on a metric in
 * the metric list are considered a double-click.
 */
MButtons.DOUBLE_CLICK_TIMEOUT = 250;


/**
 * Handler for clicks on the metric buttons.  Attached to the parent div.
 * @param {Object} e The event handler object.
 * @private
 */
MButtons.prototype.metricClickHandler_ = function(e) {
  var elem = e.target;
  var labelsIndex = parseInt(elem.id);
  if (isNaN(labelsIndex)) { return; }
  if (e.ctrlKey || e.altKey || e.shiftKey) {  // Select ONLY selected element.
    this.graphs_.setVisibilityForAllGraphs(false);
    elem.checked = true;
    this.checkAllButtonElem.attr('checked', false);
  }
  var previousValue = elem.checked;
  var self = this;
  setTimeout(function() {
    if (elem.checked != previousValue) {
      var newRawMetricXScale = self.graphs_.labelsArray[parseInt(elem.id)];
      var newMetricXScale = newRawMetricXScale.split('/').pop();
      self.scales_.applyNewMetricXScale(newMetricXScale);
    }
  }, MButtons.DOUBLE_CLICK_TIMEOUT);

  this.graphs_.setVisibilityForMetric(labelsIndex, elem.checked);
  if (!elem.checked) {
    this.unHighlightAll_();
  }
  this.updateDisplayFrag();
};


/**
 * Update the URL for metric visibility.  Should be called only after a metric
 * selector button or the metric check-all button is clicked.  Is also called
 * when old-style URL metric selectors are converted to the new style.
 */
MButtons.prototype.updateDisplayFrag = function() {
  var fragStringDisplay = '';
  for (var hash in this.graphs_.displayMap) {
    fragStringDisplay += hash;
  }
  this.params_.updateFragment('display', fragStringDisplay);
};


/**
 * Either select all graphs or select none.
 * @param {boolean} shouldCheckAll If True select all graphs, otherwise none.
 * @private
 */
MButtons.prototype.selectAllOrNone_ = function(shouldCheckAll) {
  this.graphs_.setVisibilityForAllGraphs(shouldCheckAll);
  this.updateDisplayFrag();
};


/**
 * Handle clicks on the select-all button.
 * @param {Object} e The event handler object.
 * @private
 */
MButtons.prototype.selectAllClickHandler_ = function(e) {
  this.selectAllOrNone_(Boolean(e.target.checked));
};


/**
 * Change stack property for graph and URL based on stackAgg value.
 * @param {boolean=} opt_noGraphUpdate If true don't redraw the graph.  By
 *   default we redraw the graph after updating the aforementioned state.
 */
MButtons.prototype.applyStackChange = function(opt_noGraphUpdate) {
  if (this.graphs_.stackAgg) {
    this.stackButtonElem.removeClass(this.buttonPressedClassName);
    this.stackButtonOppositeElem.addClass(this.buttonPressedClassName);
    this.shadowSliderElem.addClass(this.greyedOutClassName);
    this.params_.updateFragment('stack', '1');
  } else {
    this.stackButtonElem.addClass(this.buttonPressedClassName);
    this.stackButtonOppositeElem.removeClass(this.buttonPressedClassName);
    this.shadowSliderElem.removeClass(this.greyedOutClassName);
    this.params_.deleteFromFragment('stack', true);
  }

  if (!opt_noGraphUpdate) {
    this.graphs_.updateAggregateGraph(this.aggsliders_.lineSelect,
        this.aggsliders_.shadowSelect);
  }
};


/**
 * Register handler for the stacked/separated button pair.
 */
MButtons.prototype.registerStackChangeHandler = function() {
  var self = this;
  var handler = function(event) {
    event.preventDefault();
    self.graphs_.stackAgg = !self.graphs_.stackAgg;
    self.applyStackChange();  // Call explicitly because we don't reload.
  };
  this.stackButtonElem.on('click', handler);
  this.stackButtonOppositeElem.on('click', handler);
};


/**
 * Graph options to use when selecting highlighted lines.
 */
MButtons.HIGHLIGHT_OPTIONS = {
  highlightSeriesOpts: {
    strokeWidth: 3,
    strokeBorderWidth: 1,
    highlightCircleSize: 5
  },
  showLabelsOnHighlight: false
};


/**
 * Graph options to use when unhighlighting lines.
 */
MButtons.UNHIGHLIGHT_OPTIONS = {
  showLabelsOnHighlight: true
};


/**
 * Highlight a series for a given graph.
 * @param {Dygraph} g A Dygraph graph object.
 * @param {string} label A label.
 * @private
 */
MButtons.highlight_ = function(g, label) {
  g.updateOptions(MButtons.HIGHLIGHT_OPTIONS, true);
  g.setSelection(false, label);
};


/**
 * Unhighlight all series for a given graph.
 * @param {Dygraph} g A Dygraph graph object.
 * @private
 */
MButtons.unHighlight_ = function(g) {
  g.clearSelection();
  delete g.user_attrs_.highlightSeriesOpts;
  g.updateOptions(MButtons.UNHIGHLIGHT_OPTIONS, true);
};


/**
 * Highlight a series in all graphs.
 * @param {string} label A label.
 * @private
 */
MButtons.prototype.highlightAll_ = function(label) {
  MButtons.highlight_(this.graphs_.aggChart, label);
  MButtons.highlight_(this.graphs_.rawChart, label);
  MButtons.highlight_(this.graphs_.histoChart, label);
};


/**
 * Unhighlight all series in all graphs.
 * @private
 */
MButtons.prototype.unHighlightAll_ = function() {
  MButtons.unHighlight_(this.graphs_.aggChart);
  MButtons.unHighlight_(this.graphs_.rawChart);
  MButtons.unHighlight_(this.graphs_.histoChart);
};


/**
 * Handle metric box hovers.  Here we want to highlight a series if we mouseover
 * its label.
 * @param {Object} e An event object.
 * @private
 */
MButtons.prototype.hoverHandler_ = function(e) {
  e.preventDefault();
  var elem = e.target;
  var elemId = elem.id;
  if (elemId.indexOf('check') !== 0) { return; }
  var labelsIndex = elemId.slice(5);
  if ($('#' + labelsIndex).is(':checked')) {
    labelsIndex = Math.floor(labelsIndex);
    var label = this.graphs_.labelsArray[labelsIndex];
    this.highlightAll_(label);
  } else {
    this.unHighlightAll_();
  }
};


/**
 * Register handlers which we can attach before any loads.
 * @param {jQuerySelector} metricSearchBoxElem The metric search box div.
 */
MButtons.prototype.registerPreLoadHandlers = function(metricSearchBoxElem) {
  metricSearchBoxElem = $(metricSearchBoxElem);
  metricSearchBoxElem.focus();
  var self = this;
  $(metricSearchBoxElem).keyup(
      function(e) {self.metricSearchBoxHandler_(e);}
  );

  // Prevent <Enter> in the metric search box from submitting.
  metricSearchBoxElem.keydown(function(event) {
    if (event.which == 13) {  // Enter.
      event.preventDefault();
    }
  });
};


/**
 * Register handlers which we can only attach after we load data.
 * @param {jQuerySelector} metricButtonDivElem The metric button container div.
 * @param {jQuerySelector} checkAllButtonElem The metric button check-all button
 *   div.
 */
MButtons.prototype.registerPostLoadHandlers = function(metricButtonDivElem,
    checkAllButtonElem) {
  // Install these handlers only after a load is successful and metric
  // select buttons are present.
  metricButtonDivElem = $(metricButtonDivElem);
  this.checkAllButtonElem = $(checkAllButtonElem);
  var self = this;
  metricButtonDivElem.click(function(e) {self.metricClickHandler_(e);});
  metricButtonDivElem.mouseover(
      function(e) {setTimeout(self.hoverHandler_(e), 0);});
  metricButtonDivElem.mouseleave(
      function(e) {setTimeout(self.unHighlightAll_(), 0);});
  this.checkAllButtonElem.click(function(e) {self.selectAllClickHandler_(e);});
};


/**
 * Given text in the metric filter box, selectively show metrics which contain
 * at least one match.
 */
MButtons.prototype.filterMetricsUsingBoxText = function() {
  if (!this.graphs_.labelsArrayFull) {  // Make sure graph data is loaded.
    return;
  }
  var searchValue = $.trim(this.metricSearchInputBoxElem.val());
  var showAll = (searchValue == '');
  searchValue = searchValue.toLowerCase();
  for (var i = 0; i < this.graphs_.labelsArrayFull.length; i++) {
    var currentLabelArray = this.graphs_.labelsArrayFull[i].split('/');
    var metric = currentLabelArray.pop();
    if (showAll || (metric.toLowerCase().indexOf(searchValue) != -1)) {
      $('#' + this.metricSelectorIdPrefix + i).show();
    } else {
      $('#' + this.metricSelectorIdPrefix + i).hide();
    }
  }
};


/**
 * Handle key events in the metric search box.  Typing text causes the metric
 * list to be filtered as you type.  Pressing return will toggle selecting
 * everything in the filtered list.
 * @param {Object} event An event object.
 * @private
 */
MButtons.prototype.metricSearchBoxHandler_ = function(event) {
  this.filterMetricsUsingBoxText();
  if ((event.keyCode == 13) && (this.checkAllButtonElem !== null)) {
    var shouldCheckAll = !this.checkAllButtonElem.is(':checked');
    this.checkAllButtonElem.prop('checked', shouldCheckAll);
    this.selectAllOrNone_(shouldCheckAll);
  }
};

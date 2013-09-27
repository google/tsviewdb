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
 * @fileoverview Implement aggregate graph resize functionality.  Specifically
 *   the capability to toggle between only displaying the aggregate graph, or
 *   displaying all 3 graphs.  This functionality is overridden whenever any
 *   embedded graphs are selected.  This file depends on jQuery.
 */



/**
 * @constructor
 * @param {Params} params A URL parameter storage and manipulation object.
 * @param {boolean} aggResize True if aggregate graph should be resized (set
 *   from query string params typically).
 * @param {number} numEmbedGraphs The number of embedded graphs selected.
 */
function AggResize(params, aggResize, numEmbedGraphs) {
  this.params_ = params;
  this.aggResize_ = aggResize;
  this.numEmbedGraphs_ = numEmbedGraphs;

  // Computed on demand:
  this.aggResizeButtonElem_ = null;
  this.hideFor1EmbedElems_ = null;
  this.lowerThirdsElem_ = null;
}


/**
 * Make the aggregate graph CSS changes based on the button selection.
 */
AggResize.prototype.applyAggResizeChange = function() {
  if (this.numEmbedGraphs_ > 0) {  // No change when there are embedded graphs.
    return;
  }
  if (this.aggResize_) {
    this.hideFor1EmbedElems_.hide();
    this.aggResizeButtonElem_.val('-');
    this.lowerThirdsElem_.css('margin-top', '-10px');
  } else {
    this.hideFor1EmbedElems_.show();
    this.aggResizeButtonElem_.val('+');
    this.lowerThirdsElem_.css('margin-top', '0px');
  }
};


/**
 * Register handlers for resizing the aggregate graph, and call the graphs
 * module to update necessary state.
 * @param {jQuerySelector} aggResizeButtonElem The button responsible for
 *   toggling the aggregate graph resize functionality.
 * @param {jQuerySelector} hideFor1EmbedClass The elements which should be
 *   hidden when only the aggregate graph is displayed.
 * @param {jQuerySelector} lowerThirdsElem The bottom, non-graph portion of the
 *   page.
 * @param {Function} graphResizeCallback A callback which accepts a boolean
 *   argument whether or not aggregate graph resize has been selected.
 */
AggResize.prototype.registerHandlers = function(aggResizeButtonElem,
    hideFor1EmbedClass, lowerThirdsElem, graphResizeCallback) {
  this.aggResizeButtonElem_ = $(aggResizeButtonElem);
  this.hideFor1EmbedElems_ = $(hideFor1EmbedClass);
  this.lowerThirdsElem_ = $(lowerThirdsElem);

  var self = this;
  this.aggResizeButtonElem_.on('click', function(event) {
    event.preventDefault();

    if (self.aggResize_) {
      self.params_.deleteFromFragment('resize', true);
    } else {
      self.params_.updateFragment('resize', '1');
    }

    self.aggResize_ = !self.aggResize_;
    graphResizeCallback(self.aggResize_);
    self.applyAggResizeChange();
  });
};

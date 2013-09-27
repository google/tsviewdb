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
 * @fileoverview Code to handle query change functionality (last points vs.
 *   range loads).  This file depends on jQuery.
 */



/**
 * @constructor
 * @param {Params} params A URL parameter storage and manipulation object.
 * @param {number} lastPts The number of points to load.  If 0 then not in last
 *   points mode, so load range instead of last points.
 * @param {numEmbedGraphs} numEmbedGraphs The number of embedded graphs chosen
 *   if any.
 */
function QueryChange(params, lastPts, numEmbedGraphs) {
  this.params_ = params;
  this.lastPts = lastPts;
  this.numEmbedGraphs = numEmbedGraphs;

  // Computed on demand:
  this.queryChangeButtonElem = null;
  this.queryChangeButtonOppositeElem = null;
  this.lastPtsElems = null;
  this.lastPtsTitleElem = null;
  this.hideForLastPtsElems = null;
}


/**
 * Number of records to load for the "last N points" button.
 */
QueryChange.DEFAULT_LAST_PTS = 15;


/**
 * Set button value based on lastPts value.
 */
QueryChange.prototype.applyQueryChangeButtonChange = function() {
  if (this.lastPts == 0) {
    this.queryChangeButtonOppositeElem
      .val('last ' + QueryChange.DEFAULT_LAST_PTS)
      .removeClass('button-pressed');
    this.queryChangeButtonElem.addClass('button-pressed');
    this.hideForLastPtsElems.css('visibility', '');
    this.lastPtsElems.hide();
  } else {
    this.queryChangeButtonOppositeElem
      .val('last ' + this.lastPts)
      .addClass('button-pressed');
    this.queryChangeButtonElem.removeClass('button-pressed');
    this.lastPtsTitleElem.text('Last ' + this.lastPts + ' Records');
    this.hideForLastPtsElems.css('visibility', 'hidden');
    if (this.numEmbedGraphs == 0) {
      this.lastPtsElems.show();
    }
  }
};


/**
 * Register button handlers.
 * @param {Function} graphLoadFunction A callback which performs a graph load.
 *   Should take two booleans: forceUncached, opt_checkAndZoom.
 *   forceUncached: If True then perform an uncached load and update caches
 *     along the way.
 *   opt_checkAndZoom: If true then update the date range window if this makes
 *     sense for the mode we're in.
 * @param {jQuerySelector} queryChangeButtonElem The button which selects range
 *   querying (the default).
 * @param {jQuerySelector} queryChangeButtonOppositeElem The button which
 *   selects last points querying.
 * @param {jQuerySelector} lastPtsClassElems Any elements which should be shown
 *   when in last points mode.
 * @param {jQuerySelector} lastPtsTitleElem The last points title text div.
 * @param {jQuerySelector} hideForLastPtsClassElems Any elements which should be
 *   hidden when in last points mode.
 */
QueryChange.prototype.registerHandlers = function(graphLoadFunction,
    queryChangeButtonElem, queryChangeButtonOppositeElem, lastPtsClassElems,
    lastPtsTitleElem, hideForLastPtsClassElems) {
  this.queryChangeButtonElem = $(queryChangeButtonElem);
  this.queryChangeButtonOppositeElem = $(queryChangeButtonOppositeElem);
  this.lastPtsElems = $(lastPtsClassElems);
  this.lastPtsTitleElem = $(lastPtsTitleElem);
  this.hideForLastPtsElems = $(hideForLastPtsClassElems);

  var self = this;
  var handler = function(event) {
    event.preventDefault();
    if (self.lastPts == 0) {
      self.lastPts = QueryChange.DEFAULT_LAST_PTS;
      self.params_.addToFragQS('last_pts', QueryChange.DEFAULT_LAST_PTS);
    } else {
      self.lastPts = 0;
      self.params_.deleteFromFragment('last_pts');
    }
    self.applyQueryChangeButtonChange();
    graphLoadFunction(false, true);
  };

  this.queryChangeButtonElem.on('click', handler);
  this.queryChangeButtonOppositeElem.on('click', handler);
};

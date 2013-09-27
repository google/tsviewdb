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
 * @fileoverview Encapsulate embedded graph functionality.  Depends on
 * jQuery.
 */



/**
 * @constructor
 * @param {Params} params A Params object used to access the URL.
 */
function Embed(params) {
  this.params_ = params;
  this.numEmbedGraphs = 0;  // Number of embedded graphs to display.
  if (this.params_.qs.embed) {  // params.qs is always set.
    this.numEmbedGraphs = parseInt(this.params_.qs.embed);
  }
}


/**
 * Register handlers to create embedded graphs.
 *
 * @param {string} embedOneElem An element to which we want to attach the
 *   handler that will create a one-graph embedded graph.
 * @param {string} embedThreeElem An element to which we want to attach the
 *   handler that will create a three-graph embedded graph.
 */
Embed.prototype.registerHandlers = function(embedOneElem, embedThreeElem) {
  var self = this;
  $(embedOneElem).on('click', function(event) {
    event.preventDefault();
    self.params_.addToQS('embed', '1');
  });

  $(embedThreeElem).on('click', function(event) {
    event.preventDefault();
    self.params_.addToQS('embed', '3');
  });
};


/**
 * Hide and Fix UI elements if embedded graphs are requested.
 *
 * @param {string} hideForAllEmbedClass Class for elements which should be
 *   hidden in any embedded graph.
 * @param {string} hideFor1EmbedClass Class for elements which should be
 *   hidden for only the one embedded graph case.
 * @param {string} graphsContainerElem Id Element containing all graphs.
 * @param {string} outerGraphElem Element containing the aggregate graph.
 */
Embed.prototype.FixUI = function(hideForAllEmbedClass, hideFor1EmbedClass,
    graphsContainerElem, outerGraphElem) {
  if (this.numEmbedGraphs > 0) {
    $(hideForAllEmbedClass).hide();
    $(graphsContainerElem).css({'margin-left': 'auto',
                              'margin-right': 'auto'});
    $(outerGraphElem).css('left', 'auto');
    if (this.numEmbedGraphs == 1) {
      $(hideFor1EmbedClass).hide();
      $('body').css('overflow-y', 'hidden');
    } else {  // 3 graphs.
      $(hideFor1EmbedClass).css('left', '0px');
    }
  }
};

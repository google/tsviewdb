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
 * @fileoverview Generate histograms from a table of data.
 */



/**
 * @constructor
 * @param {Array.<Array.<number>>} data Table of data.
 */
function Histogram(data) {
  this.data = data;
}


/**
 * Number of histogram buckets.
 */
Histogram.BUCKETS = 101;


/**
 * Compute a histogram.
 * @param {Array.<boolean>} visibility The visibility array.
 * @return {Array.<Array.<number>>} A histogram result.
 */
Histogram.prototype.compute = function(visibility) {
  var newData = [];

  if (this.data.length < 1) {
    return newData;
  }

  var minMax = this.minMax(visibility);
  var minimum = minMax[0];
  var maximum = minMax[1];

  var range = maximum - minimum;
  var bucketWidth = range / (Histogram.BUCKETS - 1);
  if (bucketWidth < 0.1) {  // Clamp to prevent division by zero later.
    bucketWidth = 0.1;
  }

  var rowLength = this.data[0].length;  // All rows same length.

  // +1 to add 0 row at end to allow seeing maximum histogram bucket.
  var buckets = new Array(Histogram.BUCKETS + 1);
  for (var i = 0; i < (Histogram.BUCKETS + 1); i++) {
    buckets[i] = new Array(rowLength);  // Make space for X-axis as well.
    var bucketMinVal = minimum + i * bucketWidth;
    buckets[i][0] = bucketMinVal;  // X-axis bucket #.
    for (var j = 1; j < rowLength; j++) {
      buckets[i][j] = 0;  // Initialize.
    }
  }

  for (var i = 0; i < this.data.length; i++) {  // Stepping through rows.
    for (var j = 1; j < rowLength; j++) {  // Stepping through metrics.
      if (visibility[j - 1]) {
        val = this.data[i][j];
        if (val == null) {
          continue;
        }
        var bucketNum = Math.round((val - minimum) / bucketWidth);
        // window.console.log('bucketNum:' + bucketNum);
        var bucketVal = buckets[bucketNum][j];
        buckets[bucketNum][j]++;
      }
    }
  }

  return buckets;
};


/**
 * Compute the min and max of the visible series for the member data table.
 * @param {Array.<boolean>} visibility The visibility array.
 * @return {Array.<number>} Min and max results.
 */
Histogram.prototype.minMax = function(visibility) {
  var rowLength = this.data[0].length;  // All rows same length.

  var minimum = null;
  var maximum = null;
  for (var i = 0; i < this.data.length; i++) {  // Stepping through rows.
    for (var j = 1; j < rowLength; j++) {  // Stepping through metrics.
      if (visibility[j - 1]) {
        var val = this.data[i][j];
        if (val == null) {
          continue;
        }
        if ((minimum == null) || (val < minimum)) {
          minimum = val;
        }
        if ((maximum == null) || (val > maximum)) {
          maximum = val;
        }
      }
    }
  }
  return [minimum, maximum];
};

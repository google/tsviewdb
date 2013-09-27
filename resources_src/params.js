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
 * @fileoverview URL parameter processing class.  Depends on jQuery and jQuery
 * plugin BBQ (url below).
 *
 * http://benalman.com/code/projects/jquery-bbq/docs/files/jquery-ba-bbq-js.html
 */



/**
 * @constructor
 */
function Params() {
  this.queryString = $.param.querystring();
  this.qs = $.deparam.querystring();
  this.frag = $.deparam.fragment();  // Must be updated if URL frag is modified.
}


/**
 * Older parameters which need to be supported.
 * @private
 */
Params.LEGACY_AJAX_PARAMS_ = ['benchmark', 'experiment'];


/**
 * Parameters used by the TSView JSON Ajax request handler.  This list is used
 * by FixLegacyURl(), hasAjaxChange(), and getAjaxQS().
 * @private
 */
Params.AJAX_PARAMS_ = ['src', 'cl', 'startDate', 'endDate', 'daysOfData',
                       'time_xscale', 'equal_xscale', 'metric_xscale',
                       'max_pts', 'config', 'ufunc',
                       'last_pts'].concat(Params.LEGACY_AJAX_PARAMS_);


/**
 * Rewrite potentially old-style URL parameters for new format.  May cause a
 * reload.
 *
 * @return {boolean} True if this function is going to cause a page reload.
 *   Used to prevent further network requests or other expensive resource use,
 *   as the page reload is asynchronous.
 */
Params.prototype.fixLegacyURL = function() {
  // Can't modify this.qs because we're looping over it, so accumulate a new
  // object.
  var newQS = {};
  var willReload = false;  // If going to reload, make sure using new path.
  for (var parameter in this.qs) {
    var value = this.qs[parameter];
    if (Params.AJAX_PARAMS_.indexOf(parameter) >= 0) {
      this.frag[parameter] = value;
      willReload = true;
    } else {
      newQS[parameter] = value;
    }
  }

  var foundOldSrc = false;
  if (this.frag.benchmark) {  // Handle very old style src specification.
    this.frag.src = this.frag.benchmark + '/' + this.frag.experiment;
    delete this.frag.benchmark;
    delete this.frag.experiment;
    foundOldSrc = true;
  }

  if (willReload || foundOldSrc) {
    this.loadFragQS(this.frag);
  }
  if (willReload) {
    Params.loadWithQS(newQS, '/v', true);
  }

  return willReload;
};


/**
 * Return a QS object suitable for use in a JSON Ajax request.
 * @return {Object.<string, string>} A QS object.
 */
Params.prototype.getAjaxQS = function() {
  var ajaxQSObj = {};
  for (var parameter in this.frag) {
    if (Params.AJAX_PARAMS_.indexOf(parameter) >= 0) {
      ajaxQSObj[parameter] = this.frag[parameter];
    }
  }
  return ajaxQSObj;
};


/**
 * Given a QS Object load a new URL.
 *
 * @protected
 * @param {Object.<string, string>} qsObj Query string object returned from
 *   $.deparam.querystring() call.
 * @param {string=} opt_pathname Pathname to use instead of current.
 * @param {boolean=} opt_replace If true replace the URL without adding history.
 */
Params.loadWithQS = function(qsObj, opt_pathname, opt_replace) {
  var newQS = $.param(qsObj, true);
  var pathname = opt_pathname || window.location.pathname;
  var locationToLoad = '//' + window.location.host +
      pathname + '?' + newQS + window.location.hash;
  if (opt_replace) {
    window.location.replace(locationToLoad);
  } else {
    window.location.href = locationToLoad;
  }
};


/**
 * Given a QS Object load a new hash.
 *
 * @protected
 * @param {Object.<string, string>} qsObj Query string object returned from
 *   $.deparam.fragment() call.
 * @param {boolean=} opt_replace If true replace the URL without adding history.
 */
Params.prototype.loadFragQS = function(qsObj, opt_replace) {
  var newQS = $.param(qsObj, true);
  if (opt_replace) {
    window.location.replace('#' + newQS);
  } else {
    window.location.hash = newQS;
  }
  // No page reload in the previous line, so we need to update our fragment
  // state manually:
  this.frag = qsObj;
};


/**
 * Add name/value pair to URL QS, and reload.  Replaces existing names.
 *
 * @param {string} name Name part of query parameter.
 * @param {string} value Value part of query parameter.
 */
Params.prototype.addToQS = function(name, value) {
  var qsObj = $.extend({}, this.qs);
  qsObj[name] = value;
  Params.loadWithQS(qsObj);
};


/**
 * Add name/value pair to URL frag QS.  Replaces existing names.
 *
 * @param {string} name Name part of query parameter.
 * @param {string} value Value part of query parameter.
 */
Params.prototype.addToFragQS = function(name, value) {
  var qsObj = $.extend({}, this.frag);
  qsObj[name] = value;
  this.loadFragQS(qsObj);
};


/**
 * Merge name/value pair with existing name as list, and reload.
 *
 * @param {string} name Name part of query parameter.
 * @param {string} value Value part of query parameter.
 */
Params.prototype.mergeAddToQS = function(name, value) {
  var valueArray = [].concat(value);  // Make sure is an Array.
  var qsObj = $.extend({}, this.qs);
  var currentValue = qsObj[name];
  if (currentValue !== undefined) {  // Merge current with new.
    valueArray = valueArray.concat(currentValue);
  }
  qsObj[name] = valueArray;
  Params.loadWithQS(qsObj);
};


/**
 * Merge name/value pair with existing name as list, and update frag.
 *
 * @param {string} name Name part of query parameter.
 * @param {string} value Value part of query parameter.
 */
Params.prototype.mergeAddToFragQS = function(name, value) {
  var valueArray = [].concat(value);  // Make sure is an Array.
  var qsObj = $.extend({}, this.frag);
  var currentValue = qsObj[name];
  if (currentValue !== undefined) {  // Merge current with new.
    valueArray = valueArray.concat(currentValue);
  }
  qsObj[name] = valueArray;
  this.loadFragQS(qsObj);
};


/**
 * Remove name from URL QS, and reload.
 *
 * @param {string} name Name part of query parameter.
 */
Params.prototype.deleteFromQS = function(name) {
  var qsObj = $.extend({}, this.qs);
  delete qsObj[name];
  Params.loadWithQS(qsObj);
};


/**
 * Update URL fragment with name/value pair.  Replacing existing name.
 *
 * @param {string} name Name part of query parameter.
 * @param {string} value Value part of query parameter.
 * @param {boolean=} opt_noReplace If true replaces the URL with adding history.
 */
Params.prototype.updateFragment = function(name, value, opt_noReplace) {
  var newFragment = $.param.fragment(window.location.hash, name + '=' + value);
  if (opt_noReplace) {
    window.location.hash = newFragment;
  } else {
    window.location.replace(newFragment);
  }
  // No page reload in the previous line, so we need to update our fragment
  // state manually:
  this.frag = $.deparam.fragment();
};


/**
 * Remove name from fragment.
 *
 * @param {string} name Name part of query parameter.
 * @param {boolean=} opt_replace If true replaces the URL without adding
 *   history.
 */
Params.prototype.deleteFromFragment = function(name, opt_replace) {
  var fragObj = $.extend({}, this.frag);
  delete fragObj[name];
  this.loadFragQS(fragObj, opt_replace);
};


/**
 * Update date range Frag QS in URL.
 *
 * @param {boolean} useDaysOfData If true use newDaysOfData to update URL
 *   instead of startDate and endDate.
 * @param {string} startDate A human readable date in the format: YYYYMMDD.
 * @param {string} endDate A human readable date in the format: YYYYMMDD.
 * @param {number} newDaysOfData Number of days of data to load from current
 *   day.
 */
Params.prototype.updateDateRangeFragQS = function(useDaysOfData, startDate,
    endDate, newDaysOfData) {
  var qsObj = $.extend({}, this.frag);
  if (useDaysOfData) {
    qsObj.daysOfData = newDaysOfData;
    delete qsObj.startDate;
    delete qsObj.endDate;
  } else {
    qsObj.startDate = startDate;
    qsObj.endDate = endDate;
    delete qsObj.daysOfData;
  }
  this.loadFragQS(qsObj);
};


/**
 * Registers a hash change handler to handle the back and forward buttons.
 *
 * @param {boolean} willReload A signal from the legacy URL handler which tells
 *   us whether or not the URL will be reloaded.  Used to prevent further
 *   network requests because these reloads happen asynchronously.
 */
Params.prototype.registerHashChangeHandler = function(willReload) {
  var self = this;
  // On back button, reload pages which would send a different Ajax request than
  // the page we're on.
  $(window).bind('hashchange', function(e) {
    if (self.hasAjaxChange() &&
        !willReload) {  // Don't reload if by chance we're redirecting.
      window.location.reload();
    }
  });
};


/**
 * Used by hashchange handler to decide if we need to perform an Ajax reload.
 * This function compares the current URL fragment with the fragment state
 * (the previous fragment).  If they differ this means that we should reload
 * the page because the underlying Ajax call has different parameters than the
 * one we're now looking at.
 *
 * @return {boolean} True if the current hash params which are used in the
 *   Ajax data request are different from those same params in the current
 *   state.  This happens if we push the back or forward buttons.
 */
Params.prototype.hasAjaxChange = function() {
  var fragActual = $.deparam.fragment();
  var fragState = $.extend({}, this.frag);  // Don't modify original with sort.
  var paramsLength = Params.AJAX_PARAMS_.length;
  for (var i = 0; i < paramsLength; i++) {
    var parameter = Params.AJAX_PARAMS_[i];
    var valueActual = [].concat(fragActual[parameter]).sort().join();
    var valueState = [].concat(fragState[parameter]).sort().join();
    if (valueActual != valueState) {
      return true;
    }
  }
  return false;
};

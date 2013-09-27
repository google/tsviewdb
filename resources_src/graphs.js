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
 * @fileoverview The graph handling object.  Interfaces with Dygraph.  Depends
 * on Dygraph and jQuery.
 */

goog.require('goog.crypt.Md5');



/**
 * @constructor
 * @param {jQuerySelector} aggChartElem The div to use to create the aggregate
 *   graph.
 * @param {jQuerySelector} rawChartElem The div to use to create the raw/samples
 *   graph.
 * @param {jQuerySelector} histoChartElem The div to use to create the histogram
 *   graph.
 * @param {number} numEmbedGraphs The number of embedded graphs requested.
 * @param {boolean} aggResize If true then we want only the aggregate graph
 *   instead of all three.
 * @param {jQuerySelector} instantInfoElem The div to use to update with the
 *   mouseover extra info.  For now, this is just the config data for each
 *   record.
 * @param {jQuerySelector} buttonColumnElem The div for the button column.
 * @param {jQuerySelector} loadingSpinnerElem The busy spinner element.  By
 *   default this is hidden.
 * @param {boolean} stackAgg The initial value for the stacking/separated state.
 *   If true we want stacked graphs; separated otherwise.
 * @param {jQuerySelector} headerLinksElem The top header links div (embedded
 *   graphs links, help, etc.).
 * @param {jQuerySelector} graphsContainerElem The outer graph containing div.
 * @param {jQuerySelector} dataRangeWidgetsElem The outer container for any data
 *   range selection widgets.  This includes the *date* range widget, last
 *   points selection (which is just a title right now), or anything else we may
 *   add.
 * @param {jQuerySelector} aggregateResizeButtonDivElem The div-based button
 *   used to toggle between different aggregate graph sizes.
 * @param {jQuerySelector} lowerThirdsElem The outer div for the lower-thirds
 *   section of the UI (the user function and help divs).
 * @param {jQuerySelector} metricButtonsElem The containing div for the metric
 *   buttons and all other related UI elements (search box, select-all button,
 *   etc.).
 * @param {string} metricSelectorIdPrefix The id prefix for the individual
 *   metric buttons.  These are dynamically created.  So this must be a string.
 * @param {string} metricButtonsDivId The closest containing div for the
 *   individual metric buttons.
 * @param {string} checkAllButtonId The id for the check-all button.  This is
 *   dynamically created.  So this must be a string.
 */
function Graphs(aggChartElem, rawChartElem, histoChartElem, numEmbedGraphs,
    aggResize, instantInfoElem, buttonColumnElem, loadingSpinnerElem, stackAgg,
    headerLinksElem, graphsContainerElem, dataRangeWidgetsElem,
    aggregateResizeButtonDivElem, lowerThirdsElem, metricButtonsElem,
    metricSelectorIdPrefix, metricButtonsDivId, checkAllButtonId) {
  this.aggChartElem_ = $(aggChartElem);
  this.rawChartElem_ = $(rawChartElem);
  this.histoChartElem_ = $(histoChartElem);
  this.srcList = null;
  this.numEmbedGraphs = numEmbedGraphs;
  this.aggResize = aggResize;  // Updated in a session by this.makeAggResize().
  this.instantInfoElem = $(instantInfoElem);
  this.buttonColumnElem = $(buttonColumnElem);

  this.loadingSpinnerElem = $(loadingSpinnerElem);

  // For page resizing only.
  this.headerLinksElem = $(headerLinksElem);
  this.graphsContainerElem = $(graphsContainerElem);
  this.dataRangeWidgetsElem = $(dataRangeWidgetsElem);
  this.aggregateResizeButtonDivElem = $(aggregateResizeButtonDivElem);
  this.lowerThirdsElem = $(lowerThirdsElem);

  this.metricButtonsElem = $(metricButtonsElem);
  this.metricSelectorIdPrefix = metricSelectorIdPrefix;
  this.metricButtonsDivId = metricButtonsDivId;
  this.checkAllButtonId = checkAllButtonId;

  this.visibilityArray = null;
  // Old-style metrics-to-display list from URL.
  this.urlVisibilityArray = null;

  this.aggChart = null;
  this.rawChart = null;
  this.histoChart = null;

  this.rawChartHasRealData = null;
  this.histoChartHasRealData = null;

  this.aggData = null;  // The raw, precomputed aggregate data.
  // This object is used in generating the scale and mouseover data; indexed by
  // actual ordinal; returning [display_ordinal, timestamp_in_millis, cl, ID,
  // srcId].  src_id is a key into srcDict.
  this.clId = {0: [0, 0, 0, 0]};
  this.srcDict = {};  // Used to map srcId to src string.
  this.labelsArray = null;
  this.labelsArrayFull = null;  // With full paths (set by Ajax call).
  this.numSrcIds = null;
  this.configsArray = null;  // Holds current configs.
  this.unit = null;  // Holds unit to display on aggregate graph.

  this.histogram = null;  // Holds the histogram object.

  // Mapping visibilityArray index to base-64 48-bit md5 hash.  This is a list
  // of hashes for all loaded metrics.
  this.labelsHashArray = null;
  // Metrics selected by user.  Key is base-64 md5 hash.  Selected if present.
  this.displayMap = null;

  // Remap aggChart indices to lower chart indices, because lower charts
  // can include any subset of the aggChart metrics.
  // We can access attributes before they're set, and for speed we don't have
  // checks in the code which uses them.  Faster to set a default object here.
  this.indexRemap = {};

  // Each element of the cache object is:
  //  {data: <JSON data for invocation ID>,
  //   colors: [remappedColors],
  //   remap: indexRemapped, configs_string}
  this.cache = null;

  // Set only for equalXscale: used to map the actual data X value to a value
  // which can be used as an index into clId.  This is created simply by sorting
  // the clId keys numerically.
  this.equalXMap = null;

  this.stackAgg = stackAgg;  // If true, produce a stacked graph.

  // The maximum wrapped label length for the aggregate graph.  Used to adjust
  // Dygraph labelsDivWidth.
  this.maxLabelLength = null;

  // These two are needed for the mouseover highlight handler:
  this.timeXScale = null;
  this.equalXScale = null;
}


/**
 * Initial graph integer data.  Used to initialize graphs before user data is
 * loaded.
 */
Graphs.INIT_DATA_INT = [[0, [0, 0, 0]],
                        [1, [0, 0, 0]]];


/**
 * Initial upper graph label.  Used before user data is loaded.
 */
Graphs.INIT_LABELS_UPPER = ['x', 'Please wait... loading.'];


/**
 * Initial lower graph labels.  Used before user data is loaded.
 */
Graphs.INIT_LABELS_LOWER = ['x', 'No data loaded yet.'];


/**
 * The aggregate graph factor for the common new height calculation function.
 */
Graphs.AG_HEIGHT_SCALE = 1.5;


/**
 * The raw graph factor for the common new height calculation function.
 */
Graphs.RAW_HEIGHT_SCALE = 0.75;


/**
 * The histogram graph factor for the common new height calculation function.
 */
Graphs.HISTO_HEIGHT_SCALE = 0.5;


/**
 * The number by which a non-time X-axis value should be divided to yield
 * the actual X-axis value as the integer result.  This is done because
 * lower digits are used to uniquify identical X-axis values.
 */
Graphs.UNIQUIFYING_DIVISOR = Math.pow(10, Constants.UNIQUIFYING_DIGITS);


/**
 * Calculate a graph width based on environment parameters.
 * @param {number} numEmbedGraphs The number of embedded graphs selected.
 * @return {number} A width to use for all graphs.
 * @private
 */
Graphs.prototype.calcNewWidth_ = function(numEmbedGraphs) {
  // TODO: Get the constants used here from UI elements.
  var windowWidth = $(window).width();
  if (numEmbedGraphs > 0) {
    var embedWidth = windowWidth - 40;
    var widthLimit = numEmbedGraphs == 1 ? 200 : 400;
    embedWidth = (embedWidth <= widthLimit) ? widthLimit : embedWidth;
    return embedWidth;
  }
  var width = windowWidth - (windowWidth < 900 ? 40 : 210);
  return width;
};


/**
 * Calculate a graph height based on environment parameters.
 * @param {number} numEmbedGraphs The number of embedded graphs selected.
 * @param {boolean} aggResize If true we want only only an aggregate graph
 *   when in non-embedded mode.  This is ignored when numEmbedGraphs is > 0.
 * @return {number} A common height for graphs to be multiplied by the specific
 *   scale factors for each graph.
 * @private
 */
Graphs.calcNewHeight_ = function(numEmbedGraphs, aggResize) {
  // TODO: Get the constants used here from UI elements.
  var height;
  if (numEmbedGraphs == 1) {
    height = $(window).height() - 65;
  } else if ((numEmbedGraphs == 0) && aggResize) {
    height = $(window).height() - 170;
  } else {
    var lowerThirdsFactor = (numEmbedGraphs > 0) ? 40 : 65;
    height = Math.floor(($(window).height() / 3) - lowerThirdsFactor);
  }
  height = (height <= 125) ? 125 : height;
  return height;
};


/**
 * Call the graph height and width calculation routines and resize all graphs.
 */
Graphs.prototype.resizeGraphs = function() {
  var width = this.calcNewWidth_(this.numEmbedGraphs);
  var height = Graphs.calcNewHeight_(this.numEmbedGraphs, this.aggResize);
  if ((this.numEmbedGraphs == 1) ||
          ((this.numEmbedGraphs == 0) && (this.aggResize))) {
    this.aggChart.resize(width, height);
  } else {
    this.aggChart.resize(width, Math.round(height * Graphs.AG_HEIGHT_SCALE));
  }
  this.rawChart.resize(width, Math.round(height * Graphs.RAW_HEIGHT_SCALE));
  this.histoChart.resize(width,
      Math.round(height * Graphs.HISTO_HEIGHT_SCALE));
};


/**
 * Update page elements for the current page width.  Specifically we hide the
 * left-side toolbar when the page size is < 900px, and show it otherwise.  This
 * function has no effect when embedded graphs are selected.
 */
Graphs.prototype.resizePage = function() {
  if (this.numEmbedGraphs > 0) {
    return;
  }
  var windowWidth = $(window).width();
  if (windowWidth < 900) {
    this.buttonColumnElem.hide();
    this.headerLinksElem.css('visibility', 'hidden');
    this.graphsContainerElem.css('margin-left', 'auto');
    this.dataRangeWidgetsElem.css('margin-left', 'auto');
    this.aggregateResizeButtonDivElem.css('left', '10px');
    this.lowerThirdsElem.css('margin-left', '20px');
  } else {
    this.buttonColumnElem.show();
    this.graphsContainerElem.css('margin-left', '165px');
    this.dataRangeWidgetsElem.css('margin-left', '165px');
    this.aggregateResizeButtonDivElem.css('left', '175px');
    this.lowerThirdsElem.css('margin-left', '185px');
    this.headerLinksElem.css('visibility', 'visible');
  }
};


/**
 * Initialize our custom interaction model.
 * @private
 */
Graphs.customInteractionModel_ = $.extend({}, Dygraph.Interaction.defaultModel);


/**
 * Make the mousewheel zoom.  This comes from one of the examples on
 * http://dygraphs.com
 * @param {Object} event An event object.
 * @param {Object} g The associated Dygraph graph object.
 * @param {Object} context The associated canvas context.
 */
Graphs.customInteractionModel_.mousewheel = function(event, g, context) {
  var normal = event.detail ? event.detail * -1 : event.wheelDelta / 40;
  var percentage = normal / 40;
  var axis = g.xAxisRange();
  var xOffset = g.toDomCoords(axis[0], null)[0];
  var x = event.offsetX - xOffset;
  var w = g.toDomCoords(axis[1], null)[0] - xOffset;
  var xPct = w == 0 ? 0 : (x / w);

  var delta = axis[1] - axis[0];
  var increment = delta * percentage;
  var foo = [increment * xPct, increment * (1 - xPct)];
  var dateWindow = [axis[0] + foo[0], axis[1] - foo[1]];

  g.updateOptions({
    dateWindow: dateWindow
  });
  Dygraph.cancelEvent(event);
};


/**
 * Parameters common to all graphs.
 */
Graphs.COMMON_DYGRAPH_PARAMS = {
  rollPeriod: 1,
  showRoller: true,
  strokeWidth: 1.5,
  highlightCircleSize: 4,
  includeZero: true,
  rightGap: 0,
  highlightSeriesBackgroundAlpha: 1,
  animatedZooms: true
};


/**
 * Create all three graphs given the div's specified in the constructor.  Does
 * not load user data.
 */
Graphs.prototype.createGraphs = function() {
  var newWidth = this.calcNewWidth_(this.numEmbedGraphs);
  var newHeight = Graphs.calcNewHeight_(this.numEmbedGraphs, this.aggResize);

  var aggHeight = null;
  if ((this.numEmbedGraphs == 1) ||
          ((this.numEmbedGraphs == 0) && (this.aggResize))) {
    aggHeight = newHeight;
  } else {
    aggHeight = Math.round(newHeight * Graphs.AG_HEIGHT_SCALE);
  }

  var aggChartParams = $.extend({
    title: this.numEmbedGraphs == 1 ? null :
        "<span class='graph-label-title'>Aggregate</span>",
    labels: Graphs.INIT_LABELS_UPPER,
    width: newWidth,
    height: aggHeight,
    fillAlpha: 0.25,
    customBars: !this.stackAgg,
    interactionModel: this.numEmbedGraphs != 1 ?
        Graphs.customInteractionModel_ : Dygraph.Interaction.defaultModel,
    connectSeparatedPoints: true,
    stackedGraph: this.stackAgg
  },
  Graphs.COMMON_DYGRAPH_PARAMS);

  var rawChartParams = $.extend({
    labels: Graphs.INIT_LABELS_LOWER,
    width: newWidth,
    height: Math.round(newHeight * Graphs.RAW_HEIGHT_SCALE),
    title: "<span class='graph-label-title'>Samples</span>",
    drawPoints: true,
    connectSeparatedPoints: true,
    xlabel: 'Time',
    axes: {
      x: {  // We know x is an epoch date, so set the appropriate options.
        valueFormatter: Dygraph.dateString_,
        axisLabelFormatter: Dygraph.dateAxisFormatter,
        ticker: Dygraph.dateTicker
      }
    }
  },
  Graphs.COMMON_DYGRAPH_PARAMS);

  var histoChartParams = $.extend({
    labels: Graphs.INIT_LABELS_LOWER,
    width: newWidth,
    height: Math.round(newHeight * Graphs.HISTO_HEIGHT_SCALE),
    fillAlpha: 0.5,
    title: "<span class='graph-label-title'>Histogram of Samples</span>",
    ylabel: 'Count',
    stepPlot: true,
    fillGraph: true
  },
  Graphs.COMMON_DYGRAPH_PARAMS);

  this.aggChart = new Dygraph(this.aggChartElem_.get(0),
      Graphs.INIT_DATA_INT, aggChartParams);
  this.rawChart = new Dygraph(this.rawChartElem_.get(0),
      Graphs.INIT_DATA_INT, rawChartParams);
  this.histoChart = new Dygraph(this.histoChartElem_.get(0),
      Graphs.INIT_DATA_INT, histoChartParams);
  var self = this;
  $(window).resize(function() {
    self.resizeGraphs();
    self.resizePage();
  });
  // This is destroyed when we load new data.  But set in case data is not
  // loaded because of a graph loading error.
  this.addRollingAvgHoverText_();
};


/**
 * Show error dialog for graph loading errors.
 * @param {string=} opt_message A user message.  If not set a default error
 *   message is used.
 * @private
 */
Graphs.showErrorDialog_ = function(opt_message) {
  if (!opt_message) {
    opt_message = 'Graph loading error.';
  }
  var dialogElem = $('<div></div>')
      .text(opt_message)
      .dialog({
            autoOpen: false,
            resizable: false,
            modal: false,
            buttons: {Dismiss: function() { $(this).dialog('close'); }}
          });
  dialogElem.dialog('open');
};


/**
 * Convert API return format into legacy data format.
 * @param {Object} data The object returned from the API call.
 * @return {Object} Data in the legacy format.
 */
Graphs.makeLegacyDataFormat = function(data) {
  var names = data.AggregatesColumnNames;
  for (var i = 0; i < names.length; i++) {

  }
  return [];
};


/**
 * Make names list from flat list which is in the form <metric>.<aggregate>, ...
 * @param {Array.<string>} flatList The input list of aggregate names from the
 *   DB.
 * @return {Array.<string>} An output list of aggregate names minus the
 *   .aggregate suffix.
 */
Graphs.getNamesFromFlatList = function(flatList) {
  var returnList = [];
  var previousMetric;
  for (var i = 0; i < flatList.length; i++) {
    var element = flatList[i];
    var metric = element.substring(0, element.lastIndexOf('.'));
    if (previousMetric != metric) {
      returnList.push(metric);
    }
    previousMetric = metric;
  }
  return returnList;
};


/**
 * Load aggregate graph data.
 * @param {string} queryStringObj The query string object to use for the AJAX
 *   load request.
 * @param {number} lineSelect The line select index.
 * @param {number} shadowSelect The shadow select index.
 * @param {boolean} timeXScale If true use X-axis in units of time, otherwise
 *   use units of CL.
 * @param {boolean} equalXScale If true use an equally spaced X-axis, otherwise
 *   use variable spacing.
 * @param {string} metricXScale Metric to use for X-axis if timeXScale is
 *   false.
 * @param {boolean} logYScale If true use a log scale Y-axis, otherwise use
 *   a linear scale.
 * @param {Function} newDateWindowFunction A function which returns a two-
 *   element array which represents the current date window.
 * @param {Function} postAggLoadFunction A function to be called after a
 *   successful graph data load.
 * @param {boolean} forceUncached If true force an uncached load.
 */
Graphs.prototype.loadAggregateData = function(queryStringObj, lineSelect,
    shadowSelect, timeXScale, equalXScale, metricXScale, logYScale,
    newDateWindowFunction, postAggLoadFunction, forceUncached) {
  // These two are used in the highlight handler:
  this.timeXScale = timeXScale;
  this.equalXScale = equalXScale;

  this.srcList = queryStringObj.src;  // Always present
  // delete queryStringObj.src;

  queryStringObj['returnConfigs'] = 1;
  queryStringObj['returnIds'] = 1;
  queryStringObj['aggregates'] =
      // TODO: autogenerate this list.
      'min,p1,p5,p10,p25,p50,mean,p75,p90,p95,p99,max';
  queryStringObj['setAggregateIfMissing'] = 1;

  if (queryStringObj.last_pts > 0) {
    delete queryStringObj.startDate;
    delete queryStringObj.endDate;
    delete queryStringObj.daysOfData;
    queryStringObj['maxResults'] = queryStringObj.last_pts;
  }

  if (forceUncached) {
    queryStringObj['force_uncached'] = 1;
  }
  var queryString = $.param(queryStringObj, true);

  this.loadingSpinnerElem.show();
  var self = this;
  var jqxhr = $.get('srcs/v1?' + queryString,
      function(data) {
        self.loadingSpinnerElem.hide();

        //////////////////////////////////////////////////////////////////////
        // Handle errors.

        var errorMessage = data.message;
        if (errorMessage && errorMessage.length > 0) {
          Graphs.showErrorDialog_(errorMessage);
          return;
        }

        //////////////////////////////////////////////////////////////////////
        // Read incoming aggregate data.

        // self.aggData = data.agg_graph.file;
        // self.aggData = Graphs.makeLegacyDataFormat(data);
        self.aggData = data.aggregates;

        if (self.aggData.length < 1) {
          Graphs.showErrorDialog_('Graph loading error: No data to plot!');
          return;
        }

        //////////////////////////////////////////////////////////////////////
        // Reset state.

        self.rawChartHasRealData = false;
        self.histoChartHasRealData = false;
        self.cache = {};  // Clear the mouseover cache.

        //////////////////////////////////////////////////////////////////////
        // Read incoming labels, cl/id map, configs array, and unit.

        // var labelsArrayWithX = self.processLabels_(
        //     data.agg_graph.labels.slice());
        var labelsArrayWithX = self.processLabels_(
            Graphs.getNamesFromFlatList(data.aggregatesColumnNames));

        // FIX FIX FIX  FIX FIX FIX  FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX

        // BUILD CLID HERE.

        // clid[actual_ordinal] = [display_ordinal, timestamp_in_millis, cl, ID, srcId]

        // actual_ordinal: always use X-axis
        // display_ordinal: if time, use X-axis,
        //                  if sortByColumn use X-axis
        //                  if equal&time, use timestamp array
        //                  if equal&sortByColumn, use column
        // timestamp_in_millis: if time: use X-axis, otherwise use timestamp array
        // cl: skip... fill with 0... remove this functionality as it will be in configs
        // ID: get from id array
        // srcId: get rid of this... comes already in mouseover data.

        for (var rowCount = 0; rowCount < self.aggData.length; rowCount++) {
          var actualOrdinal = self.aggData[rowCount][0];
          var rowId = data.ids[rowCount];
          self.clId[actualOrdinal] = [actualOrdinal, actualOrdinal, 0, rowId, 0];
        }
        self.srcDict = {'0': self.srcList};

        // self.clId = data.clid;
        // self.srcDict = data.src_dict;
        if (self.equalXScale) {
          self.equalXMap = Object.keys(self.clId).sort(
              function(a, b) {return a - b});
        } else {
          self.equalXMap = null;
        }
        // FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX
        // self.configsArray = data.configs;
        // self.configsArray = data.configs[0];
        self.configsArray = Graphs.processConfigsArray_(data.configsColumnNames,
          data.configs);
        // FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX
        self.unit = '';

        //////////////////////////////////////////////////////////////////////
        // Select mode-specific graph options.

        var xAxisLabelFormatter = null;
        var xValueFormatter = null;
        var xPixelsPerLabel = null;
        var xTicker = null;

        var maxNumberWidth = null;
        var xAxisLabelWidth = null;

        if (self.timeXScale) {
          maxNumberWidth = 6;
          if (!self.equalXScale) {
            xValueFormatter = Dygraph.dateString_;
            xAxisLabelFormatter = Dygraph.dateAxisFormatter;
            xTicker = Dygraph.dateTicker;
            xPixelsPerLabel = 60;
            xAxisLabelWidth = 50;
          } else {
            xValueFormatter = function(x) {
              var newDate = new Date(Math.floor(self.equalXMap[x]));
              return Dygraph.dateString_(newDate);
            };
            xAxisLabelFormatter = function(x, granularity) {
              var remappedX = self.equalXMap[x];
              if (remappedX === undefined) { return '';}
              // Example result: '01Apr 16:51:37'
              return new Date(self.clId[remappedX][1]).strftime('%d%b %T');
            };
            xTicker = Dygraph.numericLinearTicks;
            xPixelsPerLabel = 120;
            xAxisLabelWidth = 120;
          }
        } else {
          if (self.equalXScale) {
            xValueFormatter = function(x) {
              var remappedX = self.equalXMap[x];
              return new Date(self.clId[remappedX][1]);
            };
            xAxisLabelFormatter = function(x, granularity, opts, g) {
              // The x value passed into axisLabelFormatter is interpolated, so
              // may not match an index in clId.  In the case of equal X-axis
              // we can't interpolate because the axis is non-linear.  Instead
              // we speculatively look up the mapping in clId.
              var remappedX = self.equalXMap[x];
              if (remappedX === undefined) { return '';}
              return String(self.clId[remappedX][0]);
            };
          } else {
            xValueFormatter = function(x) {
              return new Date(self.clId[x][1]);
            };
            xAxisLabelFormatter = function(x, granularity, opts, g) {
              // The x value passed into axisLabelFormatter is interpolated.
              // In the case of variably-spaced X-axis the actual value is in
              // the top digits (lower 5 digits store the uniquifying counter),
              // so we can just return the top digits.
              var upperX = Math.floor(Math.abs(x / Graphs.UNIQUIFYING_DIVISOR));
              return String(upperX);
            };
          }

          xTicker = Dygraph.numericLinearTicks;
          xPixelsPerLabel = 80;
          maxNumberWidth = 30;
          xAxisLabelWidth = 110;
        }

        //////////////////////////////////////////////////////////////////////
        // Load the data and redraw.

        self.aggChart.updateOptions({
          file: self.getGraphDataFromAggData_(lineSelect, shadowSelect,
              self.stackAgg),
          labels: labelsArrayWithX,
          visibility: self.getVisibilityArrayAgg_(),
          dateWindow: newDateWindowFunction(),
          includeZero: !logYScale,
          logscale: logYScale,
          title: self.numEmbedGraphs == 1 ? null :
              "<span class='graph-label-title'>Aggregate: </span>" +
              "<span class='graph-label-text'>" + self.srcList + '<span>',
          xlabel: self.timeXScale ? 'Time' :
              (metricXScale == Constants.CL_METRIC_X_AXIS_TOKEN ?
                  'CL' : metricXScale),
          ylabel: self.unit,
          labelsDivWidth: self.getWidthForLabelLength_(),
          colors: Graphs.getColors_(self.aggChart, labelsArrayWithX.length - 1),
          highlightCallback: ((self.numEmbedGraphs == 1) ||
              ((self.numEmbedGraphs == 0) && (self.aggResize))) ?
              null :
              function(e, x, pts, rows) {
                self.aggChartHighlightHandler_(e, x, pts, rows);},
          maxNumberWidth: maxNumberWidth,
          xAxisLabelWidth: xAxisLabelWidth,
          axes: {
            x: {
              axisLabelFormatter: xAxisLabelFormatter,
              valueFormatter: xValueFormatter,
              ticker: xTicker,
              pixelsPerLabel: xPixelsPerLabel}
          }
        });

        //////////////////////////////////////////////////////////////////////
        // Handle post-load functions (draw buttons, resize graphs, etc.)

        self.drawMetricButtons_();
        self.resizeGraphs();
        self.addRollingAvgHoverText_();
        postAggLoadFunction();
      },
      'json');  // end $.get()

  jqxhr.fail(function(jqXHr, textStatus, errorThrown) {
    self.loadingSpinnerElem.hide();
    if (textStatus == 'timeout') {
      Graphs.showErrorDialog_(
          'Graph loading timeout: Try reloading ' +
          '(perhaps the database backends are slow).');
    } else {
      // If we get a 0 status and 'error' statusText, then we've likely hit
      // a redirect from an authentication server (but could also occur when
      // we're not connected via a proxy and our service goes down).
      if ((jqXHr.status == 0) && (jqXHr.statusText == 'error')) {
        Graphs.showErrorDialog_(
            'Please try reloading (perhaps your authentication needs ' +
            'refreshing).');
      } else {
        Graphs.showErrorDialog_('Graph loading error: ' +
            (jqXHr.statusText == 'OK' ? textStatus :
            '(' + jqXHr.status + '): ' + jqXHr.statusText));
      }
    }
  });
};  // end getAggregateGraphData()


/**
 * Turn array of configs into list of name=value pairs.
 * @param {Array.<string>} configsColumnNames Names of config columns.
 * @param {Array.<Array.<string>} configs Data table.
 * @return {Array.<string>} List of name=value pairs representing union of all
 *   configs.
 * @private
 */
Graphs.processConfigsArray_ = function(configsColumnNames, configs) {
  if (configsColumnNames === undefined) {
    return [];
  }

  // First
  var resultList = [];
  var kvSet = {};  // Used to prevent duplicates.
  var rowSize = configs.length;
  var colSize = configsColumnNames.length;
  for (var i = 0; i < rowSize; i++) {
    var row = configs[i];
    if (row === null) {
      continue;
    }
    for (var j = 0; j < colSize; j++) {
      var configVal = row[j];
      if (configVal === null) {
        continue;
      }
      var k = configsColumnNames[j];
      var kv = k + '=' + configVal;

      if (kvSet[kv]) {
        continue;
      }
      resultList.push(kv);
      kvSet[kv] = true;
    }
  }
  resultList.sort();
  return resultList;
};


/**
 * Generate a truncated MD5 hash, then encode to base64.
 * @param {string} bytes Input to use for MD5 hash input.
 * @return {string} URL-safe base-64 encoded 48-bit truncated MD5 hash.
 */
Graphs.prototype.asciiHash = function(bytes) {
  var md5 = new goog.crypt.Md5();
  md5.update(bytes);
  var hash48Array = md5.digest().slice(0, 6);
  var hash48String = String.fromCharCode.apply(this, hash48Array);
  var ascii = window.btoa(hash48String);
  // Finally make URL safe.
  return ascii.replace(/\+/g, '-').replace(/\//g, '_');
};


/**
 * Process labels: set labels arrays, calculate max length, set number of
 * source ids, and generate label hashes for use in metric selection.
 * @param {Array.<string>} rawLabelsArray The raw labels array from the JSON
 *   data.
 * @return {Array.<string>} A processed labels array suitable for use in the
 *   Dygraph updateOptions() call.
 * @private
 */
Graphs.prototype.processLabels_ = function(rawLabelsArray) {
  this.visibilityArray = [];
  this.labelsArrayFull = rawLabelsArray;
  this.labelsArray = [];
  this.labelsHashArray = [];
  this.numSrcIds = 0;
  this.maxLabelLength = 0;
  var previousSrcIdParent = null;
  for (var i = 1; i < this.labelsArrayFull.length; i++) {
    // Fix user function names, which have "~" prepended (so they sort last).
    this.labelsArrayFull[i] = this.labelsArrayFull[i].replace(/~/, '');

    var labelHash = this.asciiHash(this.labelsArrayFull[i]);
    this.labelsHashArray[i - 1] = labelHash;
    if (this.urlVisibilityArray[i - 1]) {  // For compatibility.
      this.displayMap[labelHash] = null;
    }
    this.visibilityArray[i - 1] = labelHash in this.displayMap;

    var tempFullLabel = this.labelsArrayFull[i].split('/');
    var metric = tempFullLabel.pop();
    var metric_length = metric.length;
    if (metric_length > this.maxLabelLength) {
      this.maxLabelLength = metric_length;
    }
    this.labelsArray[i] = metric;  // Speculatively use short names.
    var srcIdParent = tempFullLabel.join('/');
    if (srcIdParent != previousSrcIdParent) {
      this.numSrcIds++;
    }
    previousSrcIdParent = srcIdParent;
  }
  this.urlVisibilityArray = [];  // Clear.

  // Fix if we guessed wrong.
  if (this.numSrcIds > 1) {
    this.maxLabelLength = 0;  // Reset this first (declared in outer scope).
    for (var i = 0; i < this.labelsArrayFull.length; i++) {
      this.labelsArray[i] = this.labelsArrayFull[i];
      var metric_length = this.labelsArray[i].length;
      if (metric_length > this.maxLabelLength) {
        this.maxLabelLength = metric_length;
      }
    }
  }

  var labelsArrayWithX = this.labelsArray.slice();  // Copy of the original.
  // Remove first elements which are always: "x"
  this.labelsArrayFull.splice(0, 1);
  this.labelsArray.splice(0, 1);
  return labelsArrayWithX;
};


/**
 * Update graph for aggregate graph resize.
 * @param {boolean} aggResize True if we should show only the aggregate graph.
 */
Graphs.prototype.updateForAggResize = function(aggResize) {
  this.aggResize = aggResize;
  this.resizeGraphs();

  var self = this;
  this.aggChart.updateOptions({
    highlightCallback: ((this.numEmbedGraphs == 1) ||
            ((this.numEmbedGraphs == 0) && (this.aggResize))) ?
        null :
        function(e, x, pts, rows) {
          self.aggChartHighlightHandler_(e, x, pts, rows);}
  });
};


/**
 * Get the selected metrics from the URL.  Must be called before
 * loadAggregateData()
 * @param {Params} params A URL parameter storage and manipulation object.
 */
Graphs.prototype.getSelectedMetricsFromUrl = function(params) {
  // Load the visibility array from the URL if the old metric selection
  // parameter is used.  Otherwise just reset the visibility array.
  var tmpVisibilityArray = [];
  var visibilityStr = params.frag.visibility;
  if (visibilityStr) {
    // 'http:blah/blah#visibility=101' becomes [true, false, true]
    for (var i = 0; i < visibilityStr.length; i++) {
      tmpVisibilityArray[i] = visibilityStr.charAt(i) === '1';
    }
  }
  this.urlVisibilityArray = tmpVisibilityArray;

  // Load the selected metrics from the URL into displayMap.
  this.displayMap = {};
  var displayStr = params.frag.display;
  if (displayStr) {
    for (var i = 0; i * 8 < displayStr.length; i++) {
      var hash = displayStr.substr(i * 8, 8);
      this.displayMap[hash] = null;
    }
  }
};


/**
 * Calculate series colors.  We need this function because our three graphs must
 * match colors for particular metrics, but some graphs may not have all the
 * same series.  So we calculate our own colors based on remapped series index
 * values.  This function is a code snippet from Dygraph.prototype.getColors_().
 * @param {Object} g A graph object.
 * @param {number} num An graph series number.
 * @return {Array.<string>} A list of colors strings.
 * @private
 */
Graphs.getColors_ = function(g, num) {
  var colors = [];
  var sat = g.getOption('colorSaturation') || 1.0;
  var val = g.getOption('colorValue') || 0.5;
  var half = Math.ceil(num / 2);
  for (var i = 1; i <= num; i++) {
    // Colors are alternated to maximize contrast.
    var idx = i % 2 ? Math.ceil(i / 2) : (half + i / 2);
    var hue = (1.0 * idx / (1 + num));
    colors.push(Dygraph.hsvToRGB(hue, sat, val));
  }
  return colors;
};


/**
 * Create the metric buttons based on labels data.  Called after a successful
 * graph data load.
 * @private
 */
Graphs.prototype.drawMetricButtons_ = function() {
  // Accumulate innerHTML code.
  var innerHTML = '<input type="checkbox" id="' + this.checkAllButtonId + '">';
  innerHTML += '<label for="check-all" ';
  innerHTML += 'class="hi-on-hover">check all filtered</label>';
  innerHTML += '<br><div id="' + this.metricButtonsDivId + '">';
  var previousSrcIdParent = null;
  var firstOne = true;
  var colorsArray = Graphs.getColors_(this.aggChart,
      this.labelsArrayFull.length);
  for (var i = 0; i < this.labelsArrayFull.length; i++) {
    var currentLabelArray = this.labelsArrayFull[i].split('/');
    var metric = currentLabelArray.pop();
    var srcIdParent = currentLabelArray.join('/');
    if (srcIdParent != previousSrcIdParent) {
      if (!firstOne) {
        innerHTML += '</div><br>';
      } else {
        firstOne = false;
      }
      if (this.numSrcIds > 1) {
        innerHTML += '<div class="src-id-title">' + srcIdParent + ':</div>';
      }
      innerHTML += '<div class="metric-buttons-subsection">';
    }
    var isChosen = this.visibilityArray[i];
    var checked = isChosen ? 'checked=""' : '';
    var color = colorsArray[i];
    innerHTML += '<span id="metric-selector' + i +
                 '"><input type="checkbox" id="' + i + '" ' +
                 checked + '> <label id=' + '"check' + i + '" for="' +
                 i + '" class="hi-on-hover" style="color:' + color + '">' +
                 metric + '</label><br></span>';
    previousSrcIdParent = srcIdParent;
  }
  innerHTML += '</div><br></div>';
  this.metricButtonsElem.html(innerHTML);
};


/**
 * Return a representation of our internal visibility array for the aggregate
 * graph suitable for using to set the aggregate graph visibility array.
 * @return {Array.<boolean>} The visibility array for the aggregate graph.
 * @private
 */
Graphs.prototype.getVisibilityArrayAgg_ = function() {
  var newVisibility = [];
  for (var i = 0; i < this.labelsArray.length; i++) {
    // this.visibilityArray may not be set from URL params, so set it here.
    newVisibility[i] = Boolean(this.visibilityArray[i]);
  }
  return newVisibility;
};


/**
 * Return a representation of our internal visibility array for the lower graphs
 * suitable for using to set their visibilty arrays.  A single array is returned
 * since both graphs always have identical visibility.
 * @return {Array.<boolean>} A visibilty array for the lower graphs.
 * @private
 */
Graphs.prototype.getVisibilityArrayLower_ = function() {
  var newVisibility = [];
  for (var i = 0; i < this.labelsArray.length; i++) {
    var idx = this.indexRemap[i];
    if (idx !== undefined) {
      newVisibility[idx] = this.visibilityArray[i];
    }
  }
  return newVisibility;
};


/**
 * Invoked by external callers to update the aggregate graph given selected
 * line and shadow aggregate indices.
 * @param {number} lineSelect The line select index.
 * @param {Array<number>} shadowSelect The shadow graph select indices.
 */
Graphs.prototype.updateAggregateGraph = function(lineSelect, shadowSelect) {
  if (!this.aggData) return;
  if (this.aggData.length < 1) return;
  var file_data = this.getGraphDataFromAggData_(lineSelect, shadowSelect,
      this.stackAgg);
  this.aggChart.updateOptions({
    file: file_data,
    customBars: !this.stackAgg,
    stackedGraph: this.stackAgg ? true : null,
    fillGraph: this.stackAgg ? true : null
  });
};


/**
 * Map of sorted param order to required param order.
 */
Graphs.MAP_AGG = [
  2,  // min
  3,  // p1
  6,  // p5
  4,  // p10
  5,  // p25
  7,  // p50
  1,  // mean
  8,  // p75
  9,  // p90
  10, // p95
  11, // p99
  0   // max
];

/**
 * Extract graphable data from the aggregate bulk data.  The bulk data contains
 * an array of precomputed graphable data, where each element is the data for an
 * aggregate.  Given line and shadow select indices, we can extract the data
 * for the selected aggregates.
 * @param {number} lineSelect The line select index.
 * @param {Array.<number>} shadowSelect The shadow select indices, high and low.
 * @param {boolean} isStacked If true a stacked graph is requested, otherwise
 *   a separated graph.
 * @return {Array.<Array.<number>>} An array of graphable data.  See the Dygraph
 *   native data format spec.
 *
 *   http://dygraphs.com/data.html#array
 * @private
 */
// Graphs.prototype.getGraphDataFromAggData_ = function(lineSelect, shadowSelect,
//                                                      isStacked) {
//   var newData = [];
//   for (var i = 0; i < this.aggData.length; i++) {
//     var newLine = [];
//     var line = this.aggData[i];
//     newLine[0] = line[0];
//     for (var j = 1; j < line.length; j++) {
//       if (line[j] !== null) {
//         if (isStacked) {
//           newLine[j] = line[j][lineSelect];
//         } else {
//           newLine[j] = [line[j][shadowSelect[0]],
//                         line[j][lineSelect],
//                         line[j][shadowSelect[1]]];
//         }
//       } else {
//         newLine[j] = [null, null, null];
//       }
//     }
//     newData[i] = newLine;
//   }
//   return newData;
// };
Graphs.prototype.getGraphDataFromAggData_ = function(lineSelect, shadowSelect,
                                                     isStacked) {
  var newData = [];
  var dividend = this.aggData[0].length - 1;
  // TODO: Automatically generate the divisor.
  if ((dividend % 12) != 0) {
    window.console.log('ERROR: malformed data.');
    return;
  }

  // Remap to known order from DB.
  var shadow0Sel = Graphs.MAP_AGG[shadowSelect[0]];
  var lineSel = Graphs.MAP_AGG[lineSelect];
  var shadow1Sel = Graphs.MAP_AGG[shadowSelect[1]];
  var numMetrics = dividend / 12;

  for (var i = 0; i < this.aggData.length; i++) {  // Stepping through records.
    var newLine = [];
    var line = this.aggData[i];
    newLine[0] = line[0];
    for (var j = 0; j < numMetrics; j++) {  // Stepping through metrics.
      var offset = j * 12 + 1;
      if (isStacked) {
        newLine[j + 1] = line[offset + lineSel];
      } else {
        newLine[j + 1] = [line[offset + shadow0Sel],
                          line[offset + lineSel],
                          line[offset + shadow1Sel]];
      }
    }
    newData[i] = newLine;
  }

  return newData;
};


/**
 * Given two arrays of sorted labels, the first with more elements than the,
 * second create a mapping between the smaller one to the larger one.  This
 * function is used to create a mapping between the labels in the aggregate
 * graph and the labels in the lower graphs.  This mapping is created on
 * mouseover and is cached.  It's used to synchronize the selection of metrics
 * between the graphs when the lower graphs may not have all the same metrics
 * as the aggregate graph.
 * @param {Array.<string>} labelBig The larger array of labels.
 * @param {Array.<string>} labelSmall The smaller array of labels.
 * @return {Object<number, number>} An index remapping between labels sets.
 * @private
 */
Graphs.createIndexRemap_ = function(labelBig, labelSmall) {
  var remap = {};
  var y = 1;
  for (var i = 1; i < labelSmall.length; i++) {  // Because labels[0] == X
    while (labelSmall[i] != labelBig[y]) {
      y++;
      if (y >= labelBig.length) {
        throw new Error("Didn't find a label match for index remapping!");
      }
    }
    remap[y - 1] = (i - 1);
  }
  return remap;
};


/**
 * Handle mouseover of the aggregate graph.
 * @param {Object} e An element object.
 * @param {number} x The x-coordinate of the highlighted point.
 * @param {number} pts An array of highlighted points.  Only valid if
 *   highlightSeriesOpts is set.
 * @param {number} row Index of highlighted row.  Only valid if
 *   highlightSeriesOpts is set.
 * @param {string} seriesName Name of the highlighted series  Only valid if
 *   highlightSeriesOpts is set.
 * @private
 */
Graphs.prototype.aggChartHighlightHandler_ = function(e, x, pts, row,
    seriesName) {
  // Gate with this.equalXMap to prevent TypeError if we mouseover while a new
  // graph loads.
  var remappedX = this.equalXScale && this.equalXMap ? this.equalXMap[x] : x;
  var clIdVector = this.clId[remappedX];
  // Prevent TypeError if we mouseover before a new graph loads.
  if (!clIdVector) {return;}
  var cl = clIdVector[2];
  var id = clIdVector[3];
  var srcId = clIdVector[4];
  var srcString = this.srcDict[srcId];

  var cached = this.cache[id];
  if (cached) {
    this.indexRemap = cached.remap;
    this.updateLowerGraphs_(cached.data, cached.colors, cl, cached.htmlConfigs,
        cached.unit, srcString);
  } else {
    var self = this;
    $.get('record/v1/' + id + '?noReturnAggregates=1',
        function(data) {
          ///////////////////////////////////////////////////////////////////
          // Process labels

          // Raw and histo labels are identical, so histo labels are not sent.
          // var raw_labels = data.raw_graph.labels;  // An alias.
          var raw_labels = data.pointsColumnNames;  // An alias.

          // Fix label names.
          // Fully-qualify the metric names if we have more than one src.
          if (self.numSrcIds > 1) {
            for (var i = 0; i < raw_labels.length; i++) {
              raw_labels[i] = srcString + '/' + raw_labels[i];
            }
          }

          ///////////////////////////////////////////////////////////////////
          // Create color list and index remapping.

          self.indexRemap = Graphs.createIndexRemap_(
            self.aggChart.getOption('labels'), raw_labels);
          var remappedColors = self.getRemappedColorsFromAggChart_();

          ///////////////////////////////////////////////////////////////////
          // Allocate in cache.

          // var rawConfigs = data.configs.join(', ');
          var htmlConfigs = Graphs.processConfigPairs_(data.configPairs);

          // FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX  FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX
          // var unit = data.unit;
          var unit = 'unknown';

          self.cache[id] = {
            data: $.extend({}, data),  // Shallow copy.
            colors: remappedColors,
            remap: self.indexRemap,
            htmlConfigs: htmlConfigs,
            unit: unit};

          ///////////////////////////////////////////////////////////////////
          // Update graphs and redraw.

          self.updateLowerGraphs_(data, remappedColors, cl, htmlConfigs, unit,
            srcString);
          self.rawChartHasRealData = true;
          self.histoChartHasRealData = true;
        }  // function(data)
    );  // $.get()
  }  // else
};


/**
 * Returns a sanitized HTML representation of the configs pairs.
 * @param {Object} configPairs An object representing config name/value pairs.
 * @return {string} A sanitized HTML representation of the config pairs.
 * @private
 */
Graphs.processConfigPairs_ = function(configPairs) {
  if (configPairs === undefined) {
    return '';
  }
  var resultHTMLarray = [];
  $.each(configPairs, function(k, v) {
    var escapedK = $('<div/>').text(k).html();
    if ((v.substring(0, 7) == 'http://') || (v.substring(0, 8) == 'https://')) {
      var noSchemeV = v.substr(v.indexOf('//') + 2);
      var escapedV = $('<a/>').text(noSchemeV).attr('href', v)
        .attr('target', '_blank')[0].outerHTML;
    } else {
      var escapedV = $('<div/>').text(v).html();
    }
    resultHTMLarray.push(escapedK + '=' + escapedV);
  });

  return resultHTMLarray.join(', ');
};


/**
 * Given new data, update the raw and histo charts.
 * @param {Object} data An object initialized from JSON data.
 * @param {Array<string>} remappedColors An array of colors to set.
 * @param {number} cl The cl to display.
 * @param {string} htmlConfigs Sanitized HTML configs to display.
 * @param {string} unit Unit to display for Y-axis.
 * @param {string} srcString Data source path.
 * @private
 */
Graphs.prototype.updateLowerGraphs_ = function(data, remappedColors, cl,
    htmlConfigs, unit, srcString) {
  // var raw_file_data = data.raw_graph.file;
  // var raw_labels = data.raw_graph.labels;
  var raw_file_data = data.points;
  var raw_labels = data.pointsColumnNames;

  this.histogram = new Histogram(data.points);

// FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIX FIXFIX FIX FIX FIX FIX FIXFIX FIX FIX FIX FIX FIXFIX FIX FIX FIX FIX FIX
  // var histo_file_data = data.histo_graph.file;
  var lowerVisibilityArray = this.getVisibilityArrayLower_();
  var histo_file_data = this.histogram.compute(lowerVisibilityArray);
  // var histo_file_data = data.points;

  this.rawChart.updateOptions({
    file: raw_file_data.length > 0 ? raw_file_data : null,
    labels: raw_labels,
    labelsDivWidth: this.getWidthForLabelLength_(),
    colors: remappedColors,
    title: "<span class='graph-label-title'>Points: </span>" +
        "<span class='graph-label-text'>" + srcString + '</span>',
    ylabel: unit
  });

  this.histoChart.updateOptions({
    file: histo_file_data.length > 0 ? histo_file_data : null,
    labels: raw_labels,
    labelsDivWidth: this.getWidthForLabelLength_(),
    colors: remappedColors,
    title: "<span class='graph-label-title'>Histogram of Points: </span>" +
        "<span class='graph-label-text'>" + srcString + '</span>',
    xlabel: unit
  });

  // Draw all at once after updating visibility for everything:
  this.rawChart.attrs_.visibility = lowerVisibilityArray;
  this.histoChart.attrs_.visibility = lowerVisibilityArray;
  // Force a redraw in the Dygraph library.
  this.rawChart.updateOptions({unused: null});
  this.histoChart.updateOptions({unused: null});

  this.updateInstantInfo_(htmlConfigs);
};


/**
 * Write data into the instant info div.
 * @param {string} htmlConfigs A string representing the configs for the
 *   point we are mouseing over.  Must be sanitized HTML.
 * @private
 */
Graphs.prototype.updateInstantInfo_ = function(htmlConfigs) {
  if (htmlConfigs.length != 0) {
    this.instantInfoElem.html('Configs: ' + htmlConfigs);
  } else {
    this.instantInfoElem.text('');
  }
};


/**
 * Create a list of colors for the lower graphs from the aggregate graph.  Must
 * be called only after aggChart colors are set.
 * @return {Array.<string>} A list of colors to set.
 * @private
 */
Graphs.prototype.getRemappedColorsFromAggChart_ = function() {
  var colors = [];
  var agColors = this.aggChart.getOption('colors');
  for (var i = 0; i < agColors.length; i++) {
    if (this.indexRemap[i] !== undefined) {
      colors.push(agColors[i]);
    }
  }
  return colors;
};


/**
 * Given a maximum label length for all metrics, return a Dygraph legend div
 * width that can be used as a value for the Dygraph option: legendLabelsWidth.
 * @return {number} The dygraph legend div width.
 * @private
 */
Graphs.prototype.getWidthForLabelLength_ = function() {
  if (this.maxLabelLength > 40) {
    return (this.maxLabelLength - 40) * 5 + 250;
  } else {
    return 250;
  }
};


/**
 * Set the displayMap to the given value.
 * @param {boolean} visibilityArrayIndex The index in the visibility array.
 * @param {boolean} value The value to set.
 * @private
 */
Graphs.prototype.setDisplayMap_ = function(visibilityArrayIndex, value) {
  var hash = this.labelsHashArray[visibilityArrayIndex];
  if (value) {
    this.displayMap[hash] = null;
  } else {
    delete this.displayMap[hash];
  }
};


/**
 * Set the visibility of all values for all graphs to the given value.
 * @param {boolean} value If true show all graphs, otherwise clear all graph.
 */
Graphs.prototype.setVisibilityForAllGraphs = function(value) {
  for (var i = 0; i < this.labelsArray.length; i++) {
    if ($('#' + this.metricSelectorIdPrefix + i).is(':visible')) {
      var mbutton = document.getElementById(i.toString());
      mbutton.checked = value;
      this.setVisibilityForMetric(i, value, true);  // No redraw.
      this.visibilityArray[i] = value;
      this.setDisplayMap_(i, value);
    }
  }
  this.redrawGraphs_();
};


/**
 * Redraw all graphs.
 * @private
 */
Graphs.prototype.redrawGraphs_ = function() {
  // Force a redraw in the Dygraph library.
  this.aggChart.updateOptions({unused: null});
  this.rawChart.updateOptions({unused: null});
  this.histoChart.updateOptions({unused: null});

  //////////////////////////////////////////////////////////////////////////////////
  if (this.histogram != null) {
    var lowerVisibilityArray = this.getVisibilityArrayLower_();
    var histo_file_data = this.histogram.compute(lowerVisibilityArray);
    this.histoChart.updateOptions({
      file: histo_file_data.length > 0 ? histo_file_data : null
    });
  }
  //////////////////////////////////////////////////////////////////////////////////

};


/**
 * Set the visibility for one metric across all graphs.
 * @param {number} metricNum The series number of the metric whose visibility to
 *   modify.
 * @param {boolean} value The visibility value.  True means display the series.
 * @param {boolean=} opt_noRedraw If True don't redraw the graphs.
 */
Graphs.prototype.setVisibilityForMetric = function(metricNum, value,
    opt_noRedraw)
    {
  var aggVisibilityArray = this.aggChart.visibility();
  aggVisibilityArray[metricNum] = value;

  var idx = this.indexRemap[metricNum];
  if (idx !== undefined) {
    if (this.rawChartHasRealData) {
      var rawVisibilityArray = this.rawChart.visibility();
      rawVisibilityArray[idx] = value;
    }
    if (this.histoChartHasRealData) {
      var histoVisibilityArray = this.histoChart.visibility();
      histoVisibilityArray[idx] = value;
    }
  }
  this.visibilityArray[metricNum] = value;
  this.setDisplayMap_(metricNum, value);

  if (!opt_noRedraw) {
    this.redrawGraphs_();
  }
};


/**
 * Add hover text to the rolling average boxes for all graphs.
 * @private
 */
Graphs.prototype.addRollingAvgHoverText_ = function() {
  var graphElems = [this.aggChartElem_, this.rawChartElem_,
                    this.histoChartElem_];
  for (var i in graphElems) {
    var jqObj = graphElems[i].find('input[type=text]');
    jqObj.attr('title', 'Enter number of points over which to calculate ' +
        'rolling average.');
  }
};

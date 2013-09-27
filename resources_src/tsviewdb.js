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
 * @fileoverview Time-series visualization client main file.  Depends on the
 *   TSView Javascript library and jQuery.
 */

(function() {
  /////////////////////////////////////////////////////////////////////////////
  // Read URL params and possibly fix a legacy URL.

  var params = new Params();  // A URL parameter storage and manipulation obj.
  // True if, as a result of fixing a legacy URL, we will cause a page reload.
  var willReload = params.fixLegacyURL();  // May cause a reload.

  /////////////////////////////////////////////////////////////////////////////
  // Set initial conditions from the URL.

  // True for X-axis in units of time; false for X-axis in units of CL's.
  var timeXScale = (params.frag.time_xscale == '1') ||
                   (params.frag.time_xscale === undefined);
  // True for X-axis in equally spaced increments; false for variable spacing.
  var equalXScale = params.frag.equal_xscale == '1';
  // The metric to use for the X-axis when timeXScale is false.
  var metricXScale = params.frag.metric_xscale === undefined ?
      Constants.CL_METRIC_X_AXIS_TOKEN : params.frag.metric_xscale;
  // True for stacked graph; false for unstacked.
  var stackAgg = params.frag.stack == '1';
  // True to show only aggregate graph (no samples and histogram graph).
  var aggResize = params.frag.resize == '1';
  // Number of most recent records to load.  If zero, query by time range
  // instead.
  var lastPts = params.frag.last_pts === undefined ? 0 :
      parseInt(params.frag.last_pts);

  // Forcing an uncached load should be a one-shot event.  Subsequent loads or
  // shares of the URL shouldn't inherit this directive.
  var forceUncached = params.frag.force_uncached == '1';
  if (forceUncached) {
    params.deleteFromFragment('force_uncached');
  }

  var embed = new Embed(params);  // An embedded graph handling object.

  /////////////////////////////////////////////////////////////////////////////
  // AFTER THIS LINE THE DOM MUST BE READY!
  /////////////////////////////////////////////////////////////////////////////

  /////////////////////////////////////////////////////////////////////////////
  // Fix the UI if an embedded graph is requested.

  embed.FixUI('.hide-for-all-embed', '.hide-for-1-embed', '#graphs-container',
      '#outer-graph');

  /////////////////////////////////////////////////////////////////////////////
  // Create the graphs (no data yet).

  var graphs = new Graphs('#agg-chart', '#raw-chart', '#histo-chart',
      embed.numEmbedGraphs, aggResize, '#instant-info', '#button-column',
      '#loading-spinner', stackAgg, '#headerlinks', '#graphs-container',
      '#data-range-widgets', '#aggregate-resize-button-div', '#lower-thirds',
      '#metric-buttons', 'metric-selector', 'metric-buttons-selectors',
      'check-all');

  /////////////////////////////////////////////////////////////////////////////
  // Send the AJAX data request.

  // Aggregate line and shadow sliders object.
  var aggsliders = new AggSliders(params, graphs,
      AggSliders.DEFAULT_LINE_INDEX, AggSliders.DEFAULT_SHADOW_INDICES);

  // Scale control buttons object.
  var scales = new Scales(params, graphs, timeXScale, equalXScale, metricXScale,
      '#logscale-button', '#time-xscale-button', '#equal-xscale-button',
      '#logscale-button-opposite', '#time-xscale-button-opposite',
      '#equal-xscale-button-opposite', '#time-xscale-button-cl');

  // Graph load closure (called on initial load and by various handlers).
  var graphLoad = function(opt_forceUncachedLocal, opt_checkAndZoom) {
    $('.ui-dialog-content').dialog('close');
    var newDateWindowFunction = function() {
      return opt_checkAndZoom ? qdates.getDateWindow() : null;
    };
    var fragStringObj = params.getAjaxQS();
    graphs.loadAggregateData(fragStringObj, aggsliders.lineSelect,
        aggsliders.shadowSelect, scales.timeXScale, scales.equalXScale,
        scales.metricXScale, scales.logYScale, newDateWindowFunction,
        function() {  // Post aggregate-graph load function: called on success.
          qchange.applyQueryChangeButtonChange();
          config.handleConfigsArray(graphs.configsArray);
          mbuttons.registerPostLoadHandlers('#metric-buttons-selectors',
              '#check-all');
          // In case there's already text waiting:
          mbuttons.filterMetricsUsingBoxText();
          // Convert old-style metric selectors to new-style.
          if (params.frag.visibility) {
            mbuttons.updateDisplayFrag();
            params.deleteFromFragment('visibility');
          }
        },
        opt_forceUncachedLocal);
  };

  $.ajaxSetup({timeout: (60 * 1000)});
  if (!willReload) {  // If we're redirecting, don't send the initial load.
    graphLoad(forceUncached,
        true);  // Zoom to exact selected range if we're in linear-time mode.
  }

  graphs.resizePage();
  graphs.getSelectedMetricsFromUrl(params);
  graphs.createGraphs();

  /////////////////////////////////////////////////////////////////////////////
  // Create the UI controls and attach handlers.

  // These are ordered to minimize perceived latency.

  params.registerHashChangeHandler(willReload);

  // Aggregate graph resize object.  Handles UI for aggregate-graph only viewing
  // mode: no samples or histogram graphs shown.
  var aresize = new AggResize(params, aggResize, embed.numEmbedGraphs);
  aresize.registerHandlers('#aggregate-resize-button', '.hide-for-1-embed',
      '#lower-thirds', function(x) {graphs.updateForAggResize(x);});
  aresize.applyAggResizeChange();

  // Metric buttons UI handling object.
  var mbuttons = new MButtons(graphs, params, aggsliders, scales,
      '#stack-button', '#stack-button-opposite', '#shadow-slider',
      '#metric-search-input-box', 'button-pressed', 'greyed-out',
      'metric-selector');
  mbuttons.registerPreLoadHandlers('#metric-search-input-box');
  mbuttons.registerStackChangeHandler();
  mbuttons.applyStackChange(false);  // Update state; don't repaint the graph.

  // Object to handle query by range or last points.
  var qchange = new QueryChange(params, lastPts, embed.numEmbedGraphs);
  qchange.registerHandlers(graphLoad, '#query-change-button',
      '#query-change-button-opposite', '.last-pts', '#last-pts-title',
      '.hide-for-last-pts');
  qchange.applyQueryChangeButtonChange();

  // Create the date slider UI.
  var qdates = new QuickDates(params, graphs, qchange, '#date-slider',
      '#date-scale-left', '#date-scale-right', '#date-scale-requery',
      '#date-slider-scale', '#date-from', '#date-to', 'ui-slider-handle',
      'date-slider-handle-light', graphLoad);
  qdates.createWidget();

  // Update the scale buttons for pressed/no-pressed states.
  scales.updateButtons();
  scales.registerHandlers(graphLoad);

  // Create the aggregate graph line and shadow sliders.
  aggsliders.createSliders('#line-slider', '#shadow-slider',
      '#slider-text', 'selected-aggregate-line', 'selected-aggregate-shadow');

  // General search box handler.
  var autocompleteCache = {};
  $('#search-input-box').autocomplete(
      {source: function(request, response) {
        var term = request.term;
        if (term in autocompleteCache) {
          response(autocompleteCache[term]);
          return;
        }
        $.getJSON('dir/v1/' + term + '*',
            function(data, status, xhr) {
              autocompleteCache[term] = data.names;
              response(data.names);
            });
      },
      delay: 100,  // In millis.
      select: function(event, ui) {
        $('#search-input-box').val(ui.item.value);
        $('#search-form').submit();
      }
      });
  $('#search-input-box').keypress(function(event) {
    if (event.keyCode == 13) {  // Enter.
      event.preventDefault();
      var searchValue = $.trim($('#search-input-box').val());
      if ((searchValue.charAt(0) == '+') && (searchValue.length > 2)) {
        $('#search-input-box').val('');
        var srcArray = searchValue.substr(1).split(',');
        for (var i = 0; i < srcArray.length; i++) {
          srcArray[i] = $.trim(srcArray[i]);
        }
        params.mergeAddToFragQS('src', srcArray);
        params.mergeAddToFragQS('last_pts', 15);
        graphLoad();
        return;
      }
      $('#spinner').show();
      $('#search-form').submit();
    }
  });

  // The configs filtering UI object.
  var config = new Config(params, graphLoad);
  config.registerHandlers('#configs-div', '#configs-list', '#enabled-configs',
      '#config-search', '#selected-filters', '#only-if-have-list',
      '#clear-url-config-button', '#submit-config-button',
      '#clear-config-button');

  embed.registerHandlers('#embed-1', '#embed-3');
})();

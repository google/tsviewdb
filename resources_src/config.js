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
 * @fileoverview Config filter handling code.  This file depends on jQuery.
 */



/**
 * @constructor
 * @param {Params} params A URL parameter storage and manipulation object.
 * @param {Function} graphLoadFunction A callback which performs a graph load.
 */
function Config(params, graphLoadFunction) {
  this.params_ = params;
  this.graphLoadFunction = graphLoadFunction;

  // Set in this.registerHandlers().
  this.configsDivElem = null;
  this.configsListElem = null;
  this.enabledConfigsElem = null;
  this.configSearchElem = null;
}


/**
 * Update the UI with URL and JSON configs data.  Call after a successful load.
 * The currently selected config filters are specified via the URL.
 * @param {Array.<string>} configsArray Configuration pairs loaded from the
 *   JSON data request.
 */
Config.prototype.handleConfigsArray = function(configsArray) {
  if (configsArray.length > 0) {
    if (this.params_.frag.config !== undefined) {
      var enabledConfigsArray = [].concat(this.params_.frag.config);
      var newUl = document.createElement('ul');
      for (var i = 0; i < enabledConfigsArray.length; i++) {
        var newLi = document.createElement('li');
        newLi.className = 'tiny-text';
        var newText = document.createTextNode(enabledConfigsArray[i]);
        newLi.appendChild(newText);
        newUl.appendChild(newLi);
      }
      this.configsListElem.html('');
      this.configsListElem.append(newUl);
      this.enabledConfigsElem.show();
    } else {
      this.enabledConfigsElem.hide();
    }
    this.configsDivElem.show();
    this.configSearchElem.autocomplete('option', 'source', configsArray);
  } else {
    this.configsListElem.hide();
    this.configsDivElem.hide();
  }
};


/**
 * Register configs related handlers.
 * @param {jQuerySelector} configsDivElem The outer config containing div.
 * @param {jQuerySelector} configsListElem The selected-but-not-applied config
 *   list containing div.
 * @param {jQuerySelector} enabledConfigsElem The enabled config list containing
 *   div.
 * @param {jQuerySelector} configSearchElem The config search box element.
 * @param {jQuerySelector} selectedFiltersElem The selected config list
 *   containing div.
 * @param {jQuerySelector} onlyIfHaveListElem The container for the selected
 *   divs and associated buttons.
 * @param {jQuerySelector} clearUrlConfigButtonElem The button to clear enabled
 *   configs.
 * @param {jQuerySelector} submitConfigButtonElem The button to submit selected
 *   configs.
 * @param {jQuerySelector} clearConfigButtonElem The button to clear selected
 *   configs.
 */
Config.prototype.registerHandlers = function(configsDivElem, configsListElem,
    enabledConfigsElem, configSearchElem, selectedFiltersElem,
    onlyIfHaveListElem, clearUrlConfigButtonElem, submitConfigButtonElem,
    clearConfigButtonElem) {
  // Also used in this.handleConfigsArray().
  this.configsDivElem = $(configsDivElem);
  this.configsListElem = $(configsListElem);
  this.enabledConfigsElem = $(enabledConfigsElem);
  this.configSearchElem = $(configSearchElem);

  selectedFiltersElem = $(selectedFiltersElem);
  onlyIfHaveListElem = $(onlyIfHaveListElem);
  clearUrlConfigButtonElem = $(clearUrlConfigButtonElem);
  submitConfigButtonElem = $(submitConfigButtonElem);
  clearConfigButtonElem = $(clearConfigButtonElem);

  var self = this;
  function addConfigs_(configPair) {
    var localConfigPair = $.trim(configPair);
    if (localConfigPair.length == 0) {  // Don't add null filters.
      return;}

    var numMatchingFiltered = self.configsListElem.find('li').filter(
        function() {return $(this).text() == localConfigPair}).length;
    if (numMatchingFiltered > 0) {  // Don't add filters already invoked.
      return;}

    var selectedFiltersItems = selectedFiltersElem.find('li');
    var numMatchingSelected = selectedFiltersItems.filter(function() {
      return $(this).text() == localConfigPair}).length;
    if (numMatchingSelected > 0) {  // Don't add filters in new list.
      return;}

    if (selectedFiltersItems.length == 0) {
      selectedFiltersElem.append($('<ul>'));
      onlyIfHaveListElem.show();
    }
    selectedFiltersElem.find('ul').append($('<li>').text(localConfigPair));
  };

  // Register these two only once, unconditionally (even though the target
  // element could be hidden).
  this.configSearchElem.autocomplete({
    // 'source' is set in handleConfigsArray().
    delay: 0,
    autoFocus: true,
    select: function(event, ui) {
      event.preventDefault();
      addConfigs_(ui.item.value);
      self.configSearchElem.val('');
    }});
  this.configSearchElem.keypress(function(event) {
    if (event.keyCode == 13) {
      event.preventDefault();
      addConfigs_(event.currentTarget.value);
      event.currentTarget.value = '';
    }
  });

  clearUrlConfigButtonElem.on('click', function(event) {
    event.preventDefault();
    self.params_.deleteFromFragment('config');
    self.enabledConfigsElem.hide();
    self.configsListElem.html('');
    onlyIfHaveListElem.hide();
    selectedFiltersElem.html('');  // Clear accumulated new configs.
    self.graphLoadFunction(false, true);
  });

  submitConfigButtonElem.on('click', function(event) {
    var localConfigs = [];
    selectedFiltersElem.find('li').each(function() {
      localConfigs.push($.trim($(this).text()));
    });
    if (localConfigs.length == 0) {
      event.preventDefault();
      return;
    }
    event.preventDefault();
    // Now merge with config already in QS.
    self.params_.mergeAddToFragQS('config', localConfigs);
    onlyIfHaveListElem.hide();
    selectedFiltersElem.html('');  // Clear accumulated new configs.
    self.graphLoadFunction(false, true);
  });

  clearConfigButtonElem.on('click', function(event) {
    onlyIfHaveListElem.hide();
    selectedFiltersElem.html('');
    event.preventDefault();
  });
};

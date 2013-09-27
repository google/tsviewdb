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
 * @fileoverview Handle search box for TSView landing page.  Depends on jQuery.
 */

(function() {
  var autocompleteCache = {};

  $(document).ready(function() {
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
    $('#search-input-box').focus();
    $('#search-input-box').keypress(function(event) {
      if (event.keyCode == 13) {
        event.preventDefault();

        var searchValue = $.trim($('#search-input-box').val());
        if ((searchValue.charAt(0) == '+') && (searchValue.length > 2)) {
          var srcArray = searchValue.substr(1).split(',');
          for (var i = 0; i < srcArray.length; i++) {
            srcArray[i] = encodeURIComponent($.trim(srcArray[i]));
          }
          var srcComponent = srcArray.join('&src=');
          var locationToLoad = '//' + window.location.host + '/v#src=' +
              srcComponent + '&last_pts=15&visibility=1';
          window.location.href = locationToLoad;
          return;
        }

        $('#spinner').show();
        $('#search-form').submit();
      }
    });
  });
})();

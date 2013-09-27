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
 * @fileoverview TSView constants.  Must be kept synced with
 *   tsview/constants.py.
 */


/**
 * Namespace for TSView constants.
 */
var Constants = {};


/**
 * Special metric value used to signify "use CL" when changing X-axis
 * units.
 * @type {string}
 * @const
 */
Constants.CL_METRIC_X_AXIS_TOKEN = '__CL__';


/**
 * Number of base-10 digits to use for uniquifying identical X-axis
 * values when X-axis is not in units of time.  We append extra
 * digits to the the X-axis value.  This is so the JavaScript
 * graphing code orders the values correctly.  We remap the X-axis
 * back to the actual digits at every point we need to display
 * these values back to the user.  To summarize:
 *
 *  XX.Xyy.y <-- X axis value.
 *  ^^^  ^^^
 *   |     + Uniquifying digits.
 *   + Actual digits
 *
 * @type {number}
 * @const
 */
Constants.UNIQUIFYING_DIGITS = 5;

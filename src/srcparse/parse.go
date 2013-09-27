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

package srcparse

import (
	"strings"
)

type SrcResult struct {
	Source    string
	Metric    string
	Aggregate string
	Configs   map[string]string
}

func (a SrcResult) Equal(b SrcResult) bool {
	configEqual := len(a.Configs) == len(b.Configs)
	for k := range a.Configs {
		if a.Configs[k] != b.Configs[k] {
			configEqual = false
			break
		}
	}
	return (a.Source == b.Source) && (a.Metric == b.Metric) &&
		(a.Aggregate == b.Aggregate) && configEqual
}

// Parse takes inputs in these forms:
//
// Input                 Source  Metric  Aggregate
// -------               ------- ------- ----------
// src                   src     *        *
// src:metric            src     metric   *
// src:*.aggregate       src     *        aggregate
// src:metric.aggregate  src     mertric   aggregate
//
// and similar forms with:
// src:metric.aggregate$key1=value1
func Parse(fullSrc string) (r SrcResult) {
	result := strings.Split(fullSrc, "$")
	processed := result[0]
	if len(result) > 1 {
		r.Configs = make(map[string]string)
		for _, kv := range result[1:] {
			kvArray := strings.SplitN(kv, "=", 2)
			k := kvArray[0]
			var v string
			if len(kvArray) == 2 {
				v = kvArray[1]
			}
			r.Configs[k] = v
		}
	}

	result = strings.SplitN(processed, ":", 2)
	r.Source = result[0]
	if len(result) != 2 {
		return
	}

	result = strings.SplitN(result[1], ".", 2)
	if result[0] != "*" {
		r.Metric = result[0]
	}
	if len(result) == 2 {
		r.Aggregate = result[1]
	}

	return
}

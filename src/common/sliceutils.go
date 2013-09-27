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

package common

func ReverseStringSlice(a []string) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

func PrependStringSlice(a *[]string, element string) *[]string {
	*a = append(*a, element)
	ReverseStringSlice(*a)
	return a
}

func ReverseFloatPtrSlice(a []*float64) {
	for i, j := 0, len(a)-1; i < j; i, j = i+1, j-1 {
		a[i], a[j] = a[j], a[i]
	}
}

func PrependFloatPtrSlice(a *[]*float64, element *float64) *[]*float64 {
	*a = append(*a, element)
	ReverseFloatPtrSlice(*a)
	return a
}

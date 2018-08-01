// Copyright 2018 The ksonnet authors
//
//
//    Licensed under the Apache License, Version 2.0 (the "License");
//    you may not use this file except in compliance with the License.
//    You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
//    Unless required by applicable law or agreed to in writing, software
//    distributed under the License is distributed on an "AS IS" BASIS,
//    WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//    See the License for the specific language governing permissions and
//    limitations under the License.

package cluster

import (
	"sort"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// UnstructuredSlice is a sortable slice of k8s unstructured.Unstructured objects
type UnstructuredSlice []*unstructured.Unstructured

// Sort sorts an UnstructuredSlice
func (u UnstructuredSlice) Sort() {
	sort.Stable(u)
}
func (u UnstructuredSlice) Len() int {
	return len(u)
}
func (u UnstructuredSlice) Swap(i, j int) {
	u[i], u[j] = u[j], u[i]
}
func (u UnstructuredSlice) Less(i, j int) bool {
	// Ordered sort key extractors
	keyFuncs := []func(*unstructured.Unstructured) string{
		func(o *unstructured.Unstructured) string {
			return o.GetNamespace()
		},
		func(o *unstructured.Unstructured) string {
			return o.GroupVersionKind().String()
		},
		func(o *unstructured.Unstructured) string {
			return o.GetName()
		},
		func(o *unstructured.Unstructured) string {
			return o.GetGenerateName()
		},
		func(o *unstructured.Unstructured) string {
			return string(o.GetUID())
		},
	}

	a := u[i]
	b := u[j]

	switch {
	case a == nil:
		return true
	case b == nil:
		return false
	}

	for _, f := range keyFuncs {
		vA, vB := f(a), f(b)
		switch {
		case vA < vB:
			return true
		case vA > vB:
			return false
		}
	}

	return false
}

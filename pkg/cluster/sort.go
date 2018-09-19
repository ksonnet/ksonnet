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
	sorters := []func(a, b *unstructured.Unstructured) (less bool, keepGoing bool){
		func(a, b *unstructured.Unstructured) (bool, bool) {
			return lessString(a.GetNamespace(), b.GetNamespace())
		},
		func(a, b *unstructured.Unstructured) (bool, bool) {
			return lessKind(a.GetKind(), b.GetKind())
		},
		func(a, b *unstructured.Unstructured) (bool, bool) {
			return lessString(
				a.GroupVersionKind().GroupVersion().String(),
				b.GroupVersionKind().GroupVersion().String(),
			)
		},
		func(a, b *unstructured.Unstructured) (bool, bool) {
			return lessString(a.GetName(), b.GetName())
		},
		func(a, b *unstructured.Unstructured) (bool, bool) {
			return lessString(a.GetGenerateName(), b.GetGenerateName())
		},
		func(a, b *unstructured.Unstructured) (bool, bool) {
			return lessString(string(a.GetUID()), string(b.GetUID()))
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

	for _, f := range sorters {
		// When keys are equal, we will continue to the next sort level to break the tie.
		if less, equal := f(a, b); !equal {
			return less
		}
	}

	return false
}

func lessString(a, b string) (less bool, equal bool) {
	switch {
	case a < b:
		return true, false
	case b < a:
		return false, false
	default:
		// Not done, resume to next comparison key
		return false, true
	}
}

// Compares k8s resource kinds, applies semantic ordering (order we want to apply in)
func lessKind(a, b string) (less bool, equal bool) {
	if a == b {
		// Tell caller we will need to continue to next sort level for a tie-breaker
		return false, true
	}

	aRank, aOk := kindOrderMap[a]
	bRank, bOk := kindOrderMap[b]

	switch {
	// Known order kinds come before unknown order kinds
	case aOk && !bOk:
		return true, false
	case bOk && !aOk:
		return false, false
	case aOk && bOk:
		return aRank < bRank, false
	}

	// Fallback to lexical comparison
	return a < b, false
}

// Order we should apply resources to a cluster.
// Borrowed from https://github.com/helm/helm/blob/7cad59091a9451b2aa4f95aa882ea27e6b195f98/pkg/tiller/kind_sorter.go
var kindOrder = []string{
	"Namespace",
	"ResourceQuota",
	"LimitRange",
	"PodSecurityPolicy",
	"Secret",
	"ConfigMap",
	"StorageClass",
	"PersistentVolume",
	"PersistentVolumeClaim",
	"ServiceAccount",
	"CustomResourceDefinition",
	"ClusterRole",
	"ClusterRoleBinding",
	"Role",
	"RoleBinding",
	"Service",
	"DaemonSet",
	"Pod",
	"ReplicationController",
	"ReplicaSet",
	"Deployment",
	"StatefulSet",
	"Job",
	"CronJob",
	"Ingress",
	"APIService",
}

var kindOrderMap map[string]int

func init() {
	// Build lookup table for sorting
	kindOrderMap = make(map[string]int)
	for i, kind := range kindOrder {
		kindOrderMap[kind] = i
	}
}

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

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// sortResources sources resources by moving extensions to the end of the slice. The order of all
// the other resources is preserved.
func sortResources(resources []*metav1.APIResourceList) {
	sort.SliceStable(resources, func(i, j int) bool {
		left := resources[i]
		leftGV, _ := schema.ParseGroupVersion(left.GroupVersion)
		// not checking error because it should be impossible to fail to parse data coming from the
		// apiserver
		if leftGV.Group == "extensions" {
			// always sort extensions at the bottom by saying left is "greater"
			return false
		}

		right := resources[j]
		rightGV, _ := schema.ParseGroupVersion(right.GroupVersion)
		// not checking error because it should be impossible to fail to parse data coming from the
		// apiserver
		if rightGV.Group == "extensions" {
			// always sort extensions at the bottom by saying left is "less"
			return true
		}

		return i < j
	})
}

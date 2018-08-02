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
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_UnstructuredSlice_Sort(t *testing.T) {
	objects := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1beta2",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "d3",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "s2",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "d1",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "d2",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "s1",
				},
			},
		},
	}

	expected := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "s1",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "s2",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "d1",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1beta2",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "d3",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "d2",
				},
			},
		},
	}

	for i := 0; i < 10; i++ {
		UnstructuredSlice(objects).Sort()
	}

	require.Equal(t, expected, objects)
}

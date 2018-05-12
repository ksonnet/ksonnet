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

func Test_tagManaged(t *testing.T) {
	obj := &unstructured.Unstructured{
		Object: genObject(),
	}

	err := tagManaged(obj)
	require.NoError(t, err)

	metadata, ok := obj.Object["metadata"].(map[string]interface{})
	require.True(t, ok)

	annotations, ok := metadata["annotations"].(map[string]interface{})
	require.True(t, ok)

	managed, ok := annotations["ksonnet.io/managed"].(string)
	require.True(t, ok)

	expected := "[\"$['metadata']['annotations']['ksonnet.io/dummy']\",\"$['metadata']['name']\",\"$['spec']['replicas']\",\"$['spec']['template']['metadata']['labels']['app']\",\"$['spec']['template']['spec']['containers'][?(@.name==\\\"guiroot\\\")]\"]"
	require.Equal(t, expected, managed)
}

func Test_objectPaths(t *testing.T) {
	got := objectPaths(genObject())

	expected := []objectPath{
		{
			path: []string{"apiVersion"},
		},
		{
			path: []string{"kind"},
		},
		{
			path: []string{"metadata", "annotations", "ksonnet.io/dummy"},
		},
		{
			path: []string{"metadata", "name"},
		},
		{
			path: []string{"spec", "replicas"},
		},
		{
			path: []string{"spec", "template", "metadata", "labels", "app"},
		},
		{
			path: []string{"spec", "template", "spec", "containers"},
			childNames: []string{
				"guiroot",
			},
		},
	}

	require.Equal(t, expected, got)
}

func Test_convertToJSONPath(t *testing.T) {
	cases := []struct {
		name     string
		path     []string
		expected string
	}{
		{
			name:     "simple",
			path:     []string{"foo", "bar"},
			expected: "$['foo']['bar']",
		},
		{
			name:     "with selector",
			path:     []string{"foo", "bar", `?(@.name=='name')`},
			expected: "$['foo']['bar'][?(@.name=='name')]",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := convertToJSONPath(tc.path)
			require.Equal(t, tc.expected, got)
		})
	}
}

func genObject() map[string]interface{} {
	return map[string]interface{}{
		"apiVersion": "apps/v1beta1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name": "guiroot",
			"annotations": map[string]interface{}{
				"ksonnet.io/dummy": "dummy",
			},
		},
		"spec": map[string]interface{}{
			"replicas": 1,
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "guiroot",
					},
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"image": "gcr.io/heptio-images/ks-guestbook-demo:0.1",
							"name":  "guiroot",
							"ports": []interface{}{
								map[string]interface{}{
									"containerPort": 80,
								},
							},
						},
					},
				},
			},
		},
	}
}

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

	expected := "{\"pristine\":\"H4sIAAAAAAAA/1yPzWrsMAyF9/cxzjrzk93F6z5AV92UWSiJSE1iSdhKYQh+9+IMlNCV8RH6vqMdZPGDc4kqCCCzcvvuB3bq0WGJMiHgjW3VZ2JxdEjsNJETwg4SUSePKqV9l6Ii7Neot2lL6YmA11s7CCVGwLzFrOotKcZj28psaxypIPQdnJOt5NwGZ9NKA6+HhMzOnBNoVHGKwrkgfO6IieZDOebW6IvNo16OtNyWcpk3Lj6oLpeJk4b7tV38p2YH0+wv3i/+XbMj/L/XR33UWuu/HwAAAP//AQAA///Dx6kERQEAAA==\"}"
	require.Equal(t, expected, managed)
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

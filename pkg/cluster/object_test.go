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

func TestSetMetadataLabel(t *testing.T) {
	cases := []struct {
		name     string
		key      string
		value    string
		object   map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:  "with existing label",
			key:   "key",
			value: "value",
			object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"existing": "existing",
					},
				},
			},
			expected: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"existing": "existing",
						"key":      "value",
					},
				},
			},
		},
		{
			name:  "without existing label",
			key:   "key",
			value: "value",
			object: map[string]interface{}{
				"metadata": map[string]interface{}{},
			},
			expected: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"key": "value",
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj := &unstructured.Unstructured{
				Object: tc.object,
			}

			SetMetaDataLabel(obj, tc.key, tc.value)

			require.Equal(t, tc.expected, obj.Object)
		})
	}
}

func TestSetMetadataAnnotation(t *testing.T) {
	cases := []struct {
		name     string
		key      string
		value    string
		object   map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:  "with existing annotation",
			key:   "key",
			value: "value",
			object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"existing": "existing",
					},
				},
			},
			expected: map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"existing": "existing",
						"key":      "value",
					},
				},
			},
		},
		{
			name:  "without existing annotation",
			key:   "key",
			value: "value",
			object: map[string]interface{}{
				"metadata": map[string]interface{}{},
			},
			expected: map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"key": "value",
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj := &unstructured.Unstructured{
				Object: tc.object,
			}

			SetMetaDataAnnotation(obj, tc.key, tc.value)

			require.Equal(t, tc.expected, obj.Object)
		})
	}
}

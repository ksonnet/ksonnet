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

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/kubernetes/pkg/kubectl/resource"
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

func TestManagedObjects(t *testing.T) {
	obj1 := &unstructured.Unstructured{}
	obj1.SetUID(types.UID("uid"))

	obj2 := &unstructured.Unstructured{}
	obj2.SetUID(types.UID("uid"))

	obj3 := &unstructured.Unstructured{}
	obj3.SetUID(types.UID("uid2"))

	cases := []struct {
		name         string
		resourceInfo ResourceInfo
		expected     []*unstructured.Unstructured
		isErr        bool
	}{
		{
			name: "no error",
			resourceInfo: func() *fakeResourceInfo {
				return &fakeResourceInfo{
					infos: []*resource.Info{
						{Object: obj1},
						{Object: obj2},
						{Object: obj3},
					},
				}
			}(),
			expected: []*unstructured.Unstructured{obj1, obj3},
		},

		{
			name: "resource info error",
			resourceInfo: func() *fakeResourceInfo {
				return &fakeResourceInfo{
					err: errors.New("error"),
				}
			}(),
			isErr: true,
		},

		{
			name: "infos error",
			resourceInfo: func() *fakeResourceInfo {
				return &fakeResourceInfo{
					infosErr: errors.New("error"),
				}
			}(),
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			objects, err := ManagedObjects(tc.resourceInfo)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, objects)
		})
	}
}

type fakeResourceInfo struct {
	err      error
	infos    []*resource.Info
	infosErr error
}

func (fri *fakeResourceInfo) Err() error {
	return fri.err
}

func (fri *fakeResourceInfo) Infos() ([]*resource.Info, error) {
	return fri.infos, fri.infosErr
}

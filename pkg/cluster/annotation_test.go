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
)

func Test_defaultAnnoationApplier_SetOriginalConfiguration(t *testing.T) {
	cases := []struct {
		name  string
		obj   *unstructured.Unstructured
		setup func(m *defaultAnnotationApplier)
		isErr bool
	}{
		{
			name: "tag object",
			obj: &unstructured.Unstructured{
				Object: genObject(),
			},
		},
		{
			name:  "nil object",
			isErr: true,
		},
		{
			name: "nil encoder",
			obj: &unstructured.Unstructured{
				Object: genObject(),
			},
			setup: func(m *defaultAnnotationApplier) {
				m.codec = nil
			},
			isErr: true,
		},
		{
			name: "encode failure",
			obj: &unstructured.Unstructured{
				Object: genObject(),
			},
			setup: func(m *defaultAnnotationApplier) {
				fm := &fakeAnnotationCodec{
					encodeError: errors.New("failure"),
				}
				m.codec = fm
			},
			isErr: true,
		},
		{
			name: "marshal failure",
			obj: &unstructured.Unstructured{
				Object: genObject(),
			},
			setup: func(m *defaultAnnotationApplier) {
				fm := &fakeAnnotationCodec{
					marshalError: errors.New("failure"),
				}
				m.codec = fm
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newDefaultAnnotationApplier()
			if tc.setup != nil {
				tc.setup(m)
			}

			err := m.SetOriginalConfiguration(tc.obj)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			metadata, ok := tc.obj.Object["metadata"].(map[string]interface{})
			require.True(t, ok)

			annotations, ok := metadata["annotations"].(map[string]interface{})
			require.True(t, ok)

			managed, ok := annotations["ksonnet.io/managed"].(string)
			require.True(t, ok)

			expected := "{\"pristine\":\"H4sIAAAAAAAA/1yPzWrsMAyF9/cxzjrzk93F6z5AV92UWSiJSE1iSdhKYQh+9+IMlNCV8RH6vqMdZPGDc4kqCCCzcvvuB3bq0WGJMiHgjW3VZ2JxdEjsNJETwg4SUSePKqV9l6Ii7Neot2lL6YmA11s7CCVGwLzFrOotKcZj28psaxypIPQdnJOt5NwGZ9NKA6+HhMzOnBNoVHGKwrkgfO6IieZDOebW6IvNo16OtNyWcpk3Lj6oLpeJk4b7tV38p2YH0+wv3i/+XbMj/L/XR33UWuu/HwAAAP//AQAA///Dx6kERQEAAA==\"}"
			require.Equal(t, expected, managed)
		})
	}
}

type fakeAnnotationCodec struct {
	encodeError error

	marshalBytes []byte
	marshalError error

	decodeObject map[string]interface{}
	decodeError  error
}

func (m *fakeAnnotationCodec) Encode(in map[string]interface{}) error {
	return m.encodeError
}

func (m *fakeAnnotationCodec) Marshal() ([]byte, error) {
	return m.marshalBytes, m.marshalError
}

func (m *fakeAnnotationCodec) Decode() (map[string]interface{}, error) {
	return m.decodeObject, m.decodeError
}

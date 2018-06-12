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
	"bytes"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestShow(t *testing.T) {
	dummyObjects := func() ([]*unstructured.Unstructured, error) {
		return []*unstructured.Unstructured{
			{Object: map[string]interface{}{"kind": "a"}},
			{Object: map[string]interface{}{"kind": "b"}},
		}, nil
	}

	errObjects := func() ([]*unstructured.Unstructured, error) {
		return nil, errors.New("fail")
	}

	cases := []struct {
		name        string
		format      string
		expected    string
		findObjects func() ([]*unstructured.Unstructured, error)
		isErr       bool
	}{
		{
			name:        "show yaml",
			format:      "yaml",
			expected:    "---\nkind: a\n---\nkind: b\n",
			findObjects: dummyObjects,
		},
		{
			name:        "show json",
			format:      "json",
			expected:    "{\n  \"apiVersion\": \"v1\",\n  \"items\": [\n    {\n      \"kind\": \"a\"\n    },\n    {\n      \"kind\": \"b\"\n    }\n  ],\n  \"kind\": \"List\"\n}\n",
			findObjects: dummyObjects,
		},
		{
			name:        "unknown format",
			format:      "xml",
			findObjects: dummyObjects,
			isErr:       true,
		},
		{
			name:        "unable to find objects",
			findObjects: errObjects,
			isErr:       true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			test.WithApp(t, "/", func(appMock *mocks.App, fs afero.Fs) {
				var buf bytes.Buffer

				config := ShowConfig{
					App:     appMock,
					EnvName: "default",
					Out:     &buf,
					Format:  tc.format,
				}

				fn := func(a app.App, envName string, componentNames []string) ([]*unstructured.Unstructured, error) {
					assert.Equal(t, "default", envName)
					return tc.findObjects()
				}

				findOpt := func(s *Show) {
					s.findObjectsFn = fn
				}

				err := RunShow(config, findOpt)
				if tc.isErr {
					require.Error(t, err)
					return
				}

				require.NoError(t, err)
				require.Equal(t, tc.expected, buf.String())
			})
		})
	}
}

func Test_sortByKind(t *testing.T) {
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

	for i := 0; i < 10; i++ {
		sortByKind(objects)
	}

	expected := []*unstructured.Unstructured{
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
	}

	require.Equal(t, expected, objects)
}

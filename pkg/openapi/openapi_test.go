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

package openapi

import (
	"testing"

	"github.com/go-openapi/spec"
	"github.com/go-openapi/strfmt"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestValidateAgainstSchema(t *testing.T) {
	test.WithApp(t, "/", func(a *mocks.App, fs afero.Fs) {
		obj := &unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
			},
		}

		stubbedDefinitionName := func(*unstructured.Unstructured) (string, error) {
			return "name", nil
		}

		schema := &spec.Schema{}
		stubbedLoadSchema := func(app.App, string, string) (*spec.Schema, error) {
			return schema, nil
		}

		stubbedValidate := func(s *spec.Schema, data interface{}, formats strfmt.Registry) error {
			require.Equal(t, schema, s)
			require.Equal(t, obj.Object, data)
			require.Equal(t, strfmt.Default, formats)

			return nil
		}

		v := validateAgainstSchema{
			definitionName: stubbedDefinitionName,
			loadSchema:     stubbedLoadSchema,
			validate:       stubbedValidate,
		}

		errs := v.run(a, obj, "default")
		require.Nil(t, errs)
	})
}

func Test_definitionName(t *testing.T) {
	cases := []struct {
		name         string
		kind         string
		apiVersion   string
		expectedName string
		isErr        bool
		expectedErr  error
	}{
		{
			name:         "object not in core",
			kind:         "Deployment",
			apiVersion:   "apps/v1",
			expectedName: "io.k8s.api.apps.v1.Deployment",
		},
		{
			name:         "object in core",
			kind:         "Service",
			apiVersion:   "v1",
			expectedName: "io.k8s.api.core.v1.Service",
		},
		{
			name:        "crd",
			kind:        "mycrd",
			apiVersion:  "mycrd.ksonnet.io/v1",
			isErr:       true,
			expectedErr: errUnsupportedDefinition,
		},
		{
			name:       "missing kind",
			apiVersion: "v1",
			isErr:      true,
		},
		{
			name:  "missing apiVersion",
			kind:  "Service",
			isErr: true,
		},
		{
			name:       "invalid apiVersion",
			kind:       "Service",
			apiVersion: "v1/v1/v2",
			isErr:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj := &unstructured.Unstructured{
				Object: make(map[string]interface{}),
			}

			if tc.kind != "" {
				obj.Object["kind"] = tc.kind
			}

			if tc.apiVersion != "" {
				obj.Object["apiVersion"] = tc.apiVersion
			}

			name, err := definitionName(obj)
			if tc.isErr {
				if tc.expectedErr != nil {
					require.Equal(t, tc.expectedErr, err)
					return
				}

				require.Error(t, err)
				return
			}

			require.Equal(t, tc.expectedName, name)
		})
	}
}

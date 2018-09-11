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

package actions

import (
	"bytes"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/registry"
	"github.com/stretchr/testify/require"
)

func TestRegistryList(t *testing.T) {
	cases := []struct {
		name       string
		outputType string
		outputFile string
		isErr      bool
	}{
		{
			name:       "output table",
			outputType: "table",
			outputFile: "registry/list/output.txt",
		},
		{
			name:       "output table",
			outputType: "json",
			outputFile: "registry/list/output.json",
		},
		{
			name:       "invalid output",
			outputType: "invalid",
			isErr:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withApp(t, func(appMock *amocks.App) {
				in := map[string]interface{}{
					OptionApp:           appMock,
					OptionOutput:        tc.outputType,
					OptionTLSSkipVerify: false,
				}

				a, err := NewRegistryList(in)
				require.NoError(t, err)

				var buf bytes.Buffer
				a.out = &buf

				a.registryListFn = func(app.App) ([]registry.Registry, error) {
					registries := []registry.Registry{
						mockRegistry("override"),
						mockRegistry("incubator"),
					}
					return registries, nil
				}
				a.registryIsOverrideFn = func(name string) bool {
					if name == "override" {
						return true
					}
					return false
				}

				err = a.Run()
				if tc.isErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				assertOutput(t, tc.outputFile, buf.String())
			})
		})
	}
}

func TestRegistryList_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := NewRegistryList(in)
	require.Error(t, err)
}

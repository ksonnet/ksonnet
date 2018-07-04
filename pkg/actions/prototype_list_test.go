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

	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	registrymocks "github.com/ksonnet/ksonnet/pkg/registry/mocks"
	"github.com/stretchr/testify/require"
)

func TestPrototypeList(t *testing.T) {
	cases := []struct {
		name       string
		outputType string
		outputFile string
		isErr      bool
	}{
		{
			name:       "table output",
			outputType: "table",
			outputFile: "prototype/list/output.txt",
		},
		{
			name:       "json output",
			outputType: "json",
			outputFile: "prototype/list/output.json",
		},
		{
			name:       "invalid output type",
			outputType: "invalid",
			isErr:      true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withApp(t, func(appMock *amocks.App) {
				prototypes := prototype.Prototypes{}

				manager := &registrymocks.PackageManager{}
				manager.On("Prototypes").Return(prototypes, nil)

				in := map[string]interface{}{
					OptionApp:    appMock,
					OptionOutput: tc.outputType,
				}

				a, err := NewPrototypeList(in)
				require.NoError(t, err)

				a.packageManager = manager

				var buf bytes.Buffer
				a.out = &buf

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

func TestPrototypeList_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := NewPrototypeList(in)
	require.Error(t, err)
}

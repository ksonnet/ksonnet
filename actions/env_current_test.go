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
	"github.com/stretchr/testify/require"
)

func TestEnvCurrent(t *testing.T) {
	cases := []struct {
		name        string
		envName     string
		currentName string
		unset       bool
		output      string
		isErr       bool
	}{
		{
			name: "show current environment with no current environment",
		},
		{
			name:        "show current environment with current environment set",
			currentName: "default",
			output:      "default",
		},
		{
			name:    "set current",
			envName: "default",
		},
		{
			name:  "unset current",
			unset: true,
		},
		{
			name:    "error if set and unset together",
			unset:   true,
			envName: "default",
			isErr:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withApp(t, func(appMock *amocks.App) {
				appMock.On("CurrentEnvironment").Return(tc.currentName)
				appMock.On("SetCurrentEnvironment", tc.envName).Return(nil)

				in := map[string]interface{}{
					OptionApp:     appMock,
					OptionEnvName: tc.envName,
					OptionUnset:   tc.unset,
				}

				a, err := newEnvCurrent(in)
				require.NoError(t, err)

				var buf bytes.Buffer
				a.out = &buf

				err = a.run()
				if tc.isErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)
			})
		})
	}
}

func TestEnvCurrent_invalid_input(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		in := map[string]interface{}{
			OptionApp: "invalid",
		}

		_, err := newEnvCurrent(in)
		require.Error(t, err)
	})
}

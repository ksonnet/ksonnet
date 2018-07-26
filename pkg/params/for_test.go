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

package params

import (
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildEnvParamsForModule(t *testing.T) {
	cases := []struct {
		name            string
		input           string
		componentParams string
		moduleName      string
		output          string
		isErr           bool
	}{
		{
			name:            "in a nested module",
			input:           "input.jsonnet",
			componentParams: "component_params.jsonnet",
			moduleName:      "app.project-1",
			output:          "module.jsonnet",
		},
		{
			name:            "in root module",
			input:           "input.jsonnet",
			componentParams: "component_params.jsonnet",
			moduleName:      "/",
			output:          "root.jsonnet",
		},
		{
			name:            "no match",
			input:           "input.jsonnet",
			componentParams: "component_params.jsonnet",
			moduleName:      "no-match",
			output:          "no_match.jsonnet",
		},
		{
			name:            "empty input, still output components field",
			input:           "input_empty.jsonnet",
			componentParams: "component_empty.jsonnet",
			moduleName:      "no-match",
			output:          "no_match.jsonnet",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			paramsStr := test.ReadTestData(t, filepath.Join("build_env_params", tc.input))
			componentParams := test.ReadTestData(t, filepath.Join("build_env_params", tc.componentParams))

			got, err := BuildEnvParamsForModule(tc.moduleName, paramsStr, componentParams, ".")
			require.NoError(t, err)

			expected := test.ReadTestData(t, filepath.Join("build_env_params", tc.output))
			assert.Equal(t, expected, got)
		})
	}

}

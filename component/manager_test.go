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

package component

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_defaultManager_Component(t *testing.T) {
	app, fs := appMock("/")

	stageFile(t, fs, "params-mixed.libsonnet", "/components/params.libsonnet")
	stageFile(t, fs, "deployment.yaml", "/components/deployment.yaml")
	stageFile(t, fs, "params-mixed.libsonnet", "/components/nested/params.libsonnet")
	stageFile(t, fs, "deployment.yaml", "/components/nested/deployment.yaml")

	dm := defaultManager{}

	c, err := dm.Component(app, "", "deployment")
	require.NoError(t, err)

	expected := "deployment"
	require.Equal(t, expected, c.Name(false))
}

func Test_default_manager_ResolvePath(t *testing.T) {
	app, fs := appMock("/")

	stageFile(t, fs, "params-mixed.libsonnet", "/components/params.libsonnet")
	stageFile(t, fs, "deployment.yaml", "/components/deployment.yaml")
	stageFile(t, fs, "params-mixed.libsonnet", "/components/nested/params.libsonnet")
	stageFile(t, fs, "deployment.yaml", "/components/nested/deployment.yaml")

	dm := &defaultManager{}

	cases := []struct {
		name   string
		cName  string
		module string
		isErr  bool
	}{
		{
			name:   "/",
			module: "/",
		},
		{
			name:   "deployment",
			module: "/",
			cName:  "deployment",
		},
		{
			name:   "/deployment",
			module: "/",
			cName:  "deployment",
		},
		{
			name:   "/nested/deployment",
			module: "/nested",
			cName:  "deployment",
		},
		{
			name:   "nested/deployment",
			module: "/nested",
			cName:  "deployment",
		},
		{
			name:  "nested/invalid",
			isErr: true,
		},
		{
			name:  "invalid",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ns, c, err := dm.ResolvePath(app, tc.name)

			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			if tc.cName == "" {
				assert.Nil(t, c)
			} else {
				require.NotNil(t, c)
				assert.Equal(t, tc.cName, c.Name(false))
			}

			assert.Equal(t, tc.module, ns.Name())
		})
	}
}

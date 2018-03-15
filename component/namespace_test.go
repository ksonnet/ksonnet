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

func TestNamespace_Components(t *testing.T) {
	app, fs := appMock("/app")

	stageFile(t, fs, "certificate-crd.yaml", "/app/components/ns1/certificate-crd.yaml")
	stageFile(t, fs, "params-with-entry.libsonnet", "/app/components/ns1/params.libsonnet")
	stageFile(t, fs, "params-no-entry.libsonnet", "/app/components/params.libsonnet")

	cases := []struct {
		name   string
		nsName string
		count  int
	}{
		{
			name:   "no components",
			nsName: "/",
		},
		{
			name:   "with components",
			nsName: "ns1",
			count:  1,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {

			ns, err := GetNamespace(app, tc.nsName)
			require.NoError(t, err)

			assert.Equal(t, tc.nsName, ns.Name())
			components, err := ns.Components()
			require.NoError(t, err)

			assert.Len(t, components, tc.count)
		})
	}

}

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

package registry

import (
	"testing"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/metadata/app/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_defaultManager_Registries(t *testing.T) {
	dm := &defaultManager{}

	specs := app.RegistryRefSpecs{
		"incubator": &app.RegistryRefSpec{
			Protocol: "github",
			URI:      "github.com/foo/bar",
		},
	}

	appMock := &mocks.App{}
	appMock.On("Registries").Return(specs)

	registries, err := dm.Registries(appMock)
	require.NoError(t, err)

	require.Len(t, registries, 1)

	r := registries[0]
	assert.Equal(t, "incubator", r.Name())
	assert.Equal(t, "github", r.Protocol())
	assert.Equal(t, "github.com/foo/bar", r.URI())
}

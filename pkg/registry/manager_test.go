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
	"github.com/stretchr/testify/require"
)

func OffTest_defaultManager_Registries(t *testing.T) {
	dm := &defaultManager{}

	specs := app.RegistryRefSpecs{
		"incubator": &app.RegistryRefSpec{
			Protocol: ProtocolGitHub,
			URI:      "github.com/ksonnet/parts/tree/master/incubator",
		},
	}

	appMock := &mocks.App{}
	appMock.On("Registries").Return(specs)

	registries, err := dm.List(appMock)
	require.NoError(t, err)

	require.Len(t, registries, 1)
}

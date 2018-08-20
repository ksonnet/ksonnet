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
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPkgRemove(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		libName := "incubator/apache"

		in := map[string]interface{}{
			OptionApp:     appMock,
			OptionPkgName: libName,
		}

		a, err := NewPkgRemove(in)
		require.NoError(t, err)

		var updaterCalled bool
		fakeUpdater := func(name string, env string, spec *app.LibraryConfig) (*app.LibraryConfig, error) {
			updaterCalled = true
			assert.Equal(t, a.envName, env, "unexpected environment name")
			assert.Nil(t, spec)
			return nil, nil
		}

		a.libUpdateFn = fakeUpdater

		err = a.Run()
		require.NoError(t, err)
		assert.True(t, updaterCalled, "library reference updater not called")
	})
}

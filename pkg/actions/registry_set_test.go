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
	"github.com/ksonnet/ksonnet/pkg/registry"
	rmocks "github.com/ksonnet/ksonnet/pkg/registry/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRegistryUpdate_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := NewRegistrySet(in)
	require.Error(t, err)
}

// Generate spec->registry.Setter locators with customized mock registries
func mockLocator(name string, oldURI string, newURI string) locateFn {
	return func(app.App, *app.RegistryConfig) (registry.Setter, error) {
		return mockSetter(name, oldURI, newURI), nil
	}
}

func mockSetter(name string, oldURI string, newURI string) registry.Setter {
	i := 0
	returnValues := []*app.RegistryConfig{
		&app.RegistryConfig{
			Name: name,
			URI:  oldURI,
		},
		&app.RegistryConfig{
			Name: name,
			URI:  newURI,
		},
	}

	u := new(rmocks.Setter)
	u.On("SetURI", newURI).Return(nil)
	u.On("MakeRegistryConfig").Return(
		func() *app.RegistryConfig {
			if i > len(returnValues)-1 {
				return nil
			}

			ret := returnValues[i]
			i++
			return ret
		},
	)
	u.On("FetchRegistrySpec").Return(nil, nil)
	return u
}

// Test lookup of registry configuration by name
func TestRegistrySet_registryConfig(t *testing.T) {
	a := new(amocks.App)
	registries := app.RegistryConfigs{
		"incubator": &app.RegistryConfig{
			Name:     "incubator",
			Protocol: string(registry.ProtocolGitHub),
			URI:      "github.com/ksonnet/parts/tree/master/incubator",
		},
	}
	a.On("Registries").Return(registries, nil)

	tests := []struct {
		caseName  string
		name      string
		app       app.App
		expected  *app.RegistryConfig
		expectErr bool
	}{
		{
			caseName:  "specific registry",
			name:      "incubator",
			app:       a,
			expected:  registries["incubator"],
			expectErr: false,
		},
		{
			caseName:  "unknown registry",
			name:      "unknown",
			app:       a,
			expected:  nil,
			expectErr: true,
		},
		{
			caseName:  "empty name",
			name:      "",
			app:       a,
			expected:  nil,
			expectErr: true,
		},
		{
			caseName:  "nil app",
			name:      "unknown",
			app:       nil,
			expected:  nil,
			expectErr: true,
		},
	}

	for _, tc := range tests {
		result, err := registryConfig(tc.app, tc.name)
		if tc.expectErr {
			require.Errorf(t, err, "test: %v", tc.name)
		} else {
			require.NoErrorf(t, err, "test: %v", tc.name)
		}

		assert.Equal(t, tc.expected, result)
	}
}

func TestRegistryUpdate_doSetURI(t *testing.T) {
	// Helpers
	makeApp := func(expectedName string, expectedURI string) *amocks.App {
		a := new(amocks.App)
		a.On("UpdateRegistry",
			mock.MatchedBy(func(spec *app.RegistryConfig) bool {
				if !assert.NotNil(t, spec) {
					return false
				}
				if !assert.Equal(t, expectedName, spec.Name) {
					return false
				}

				if !assert.Equal(t, expectedURI, spec.URI) {
					return false
				}

				return true
			}),
		).Return(nil).Once()
		return a
	}

	makeSpec := func(name string, uri string) *app.RegistryConfig {
		return &app.RegistryConfig{
			Name: name,
			URI:  uri,
		}
	}

	tests := []struct {
		name         string
		oldURI       string
		newURI       string
		shouldUpdate bool
		expectErr    bool
	}{
		{
			name:         "normal update",
			oldURI:       "github.com/ksonnet/parts/tree/master/incubator",
			newURI:       "github.com/ksonnet/parts/tree/experimental/incubator",
			shouldUpdate: true,
			expectErr:    false,
		},
		{
			name:         "no change, shouldn't update",
			oldURI:       "github.com/ksonnet/parts/tree/master/incubator",
			newURI:       "github.com/ksonnet/parts/tree/master/incubator",
			shouldUpdate: false,
			expectErr:    false,
		},
		{
			name:         "empty uri returns error",
			oldURI:       "github.com/ksonnet/parts/tree/master/incubator",
			newURI:       "",
			shouldUpdate: false,
			expectErr:    true,
		},
	}

	for _, tc := range tests {
		a := makeApp(tc.name, tc.newURI)
		l := mockLocator(tc.name, tc.oldURI, tc.newURI)
		currentCfg := makeSpec(tc.name, tc.oldURI)

		err := doSetURI(a, l, currentCfg, tc.newURI)
		if tc.expectErr {
			require.Errorf(t, err, "test: %v", tc.name)
		} else {
			require.NoErrorf(t, err, "test: %v", tc.name)
		}

		// Assert app.UpdateRegistry gets called with the new, updated uri,
		// if it was valid and a change was made.
		if tc.shouldUpdate {
			//newCfg := makeSpec(tc.currentCfg.Name, tc.newURI)
			//tc.app.AssertCalled(t, "UpdateRegistry", newCfg)
			a.AssertExpectations(t)
		} else {
			a.AssertNumberOfCalls(t, "UpdateRegistry", 0)
		}
	}
}

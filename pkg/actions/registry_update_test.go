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
	"sort"
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
	_, err := NewRegistryUpdate(in)
	require.Error(t, err)
}

// Generate spec->registry.Updater locators with customized mock registries
func mockRegistryLocator(oldVersion string, newVersion string) LocateFn {
	return func(app.App, *app.RegistryRefSpec) (registry.Updater, error) {
		u := new(rmocks.Updater)
		u.On("Update", oldVersion).Return(func(v string) (string, error) {
			return newVersion, nil
		})
		return u, nil
	}
}

func mockUpdater(oldVersion string, newVersion string) registry.Updater {
	u := new(rmocks.Updater)
	u.On("Update", oldVersion).Return(newVersion, nil)
	return u
}

// Test that a set of registries to update can be resolved, one specific or all if unspecified.
func TestRegistryUpdate_resolveUpdateSet(t *testing.T) {
	a := new(amocks.App)
	a.On("Registries").Return(
		app.RegistryRefSpecs{
			"custom":    nil,
			"helm":      nil,
			"incubator": nil,
		},
		nil,
	)
	ru := &RegistryUpdate{
		app: a,
	}

	tests := []struct {
		caseName  string
		name      string
		expected  []string
		expectErr bool
	}{
		{
			caseName:  "all registries",
			name:      "",
			expected:  []string{"custom", "helm", "incubator"},
			expectErr: false,
		},
		{
			caseName:  "specific registry",
			name:      "incubator",
			expected:  []string{"incubator"},
			expectErr: false,
		},
		{
			caseName:  "unknown registry",
			name:      "unknown",
			expected:  []string{},
			expectErr: true,
		},
	}

	for _, tc := range tests {
		result, err := ru.resolveUpdateSet(tc.name)
		if tc.expectErr {
			require.Errorf(t, err, "test: %v", tc.name)
		} else {
			require.NoErrorf(t, err, "test: %v", tc.name)
		}

		// Don't make assertions about return value if error was returned
		if err != nil {
			continue
		}

		sort.Strings(result)
		assert.Equal(t, tc.expected, result)
	}
}

func TestRegistryUpdate_doUpdateRegistry(t *testing.T) {
	// Helpers
	makeApp := func(newVersion string) *amocks.App {
		a := new(amocks.App)
		a.On("UpdateRegistry",
			mock.MatchedBy(func(spec *app.RegistryRefSpec) bool {
				if spec == nil {
					t.Errorf("spec is nil")
					return false
				}
				if spec.GitVersion == nil {
					t.Errorf("spec.GitVersion is nil")
					return false
				}
				if spec.GitVersion.CommitSHA != newVersion {
					t.Errorf("unexpected version argument: expected %v, got %v", newVersion, spec.GitVersion.CommitSHA)
					return false
				}
				return true
			}),
		).Return(nil).Once()
		return a
	}

	makeSpec := func(version string) *app.RegistryRefSpec {
		return &app.RegistryRefSpec{
			GitVersion: &app.GitVersionSpec{
				CommitSHA: version,
			},
		}
	}

	tests := []struct {
		name             string
		app              *amocks.App
		updater          registry.Updater
		rs               *app.RegistryRefSpec
		requestedVersion string
		expected         string
		shouldUpdate     bool
		expectErr        bool
	}{
		{
			name:             "normal update",
			app:              makeApp("newVersion"),
			updater:          mockUpdater("", "newVersion"),
			rs:               makeSpec("currentVersion"),
			requestedVersion: "",
			expected:         "newVersion",
			shouldUpdate:     true,
			expectErr:        false,
		},
		{
			name:             "no change, shouldn't update",
			app:              makeApp("XXXX"),
			updater:          mockUpdater("", "currentVersion"),
			rs:               makeSpec("currentVersion"),
			requestedVersion: "",
			expected:         "currentVersion",
			shouldUpdate:     false,
			expectErr:        false,
		},
		{
			name:             "doesn't support targeted version yet",
			app:              makeApp("newVersion"),
			updater:          mockUpdater("", "newVersion"),
			rs:               makeSpec("currentVersion"),
			requestedVersion: "someVersion",
			expected:         "someVersion",
			shouldUpdate:     false,
			expectErr:        true,
		},
		// {
		// 	name:             "nil app returns error",
		// 	app:              nil,
		// 	updater:          mockUpdater("", "newVersion"),
		// 	rs:               makeSpec("currentVersion"),
		// 	requestedVersion: "",
		// 	expected:         "",
		// 	shouldUpdate:     false,
		// 	expectErr:        true,
		// },
		{
			name:             "no registrySpec returns error",
			app:              makeApp("newVersion"),
			updater:          mockUpdater("", "newVersion"),
			rs:               nil,
			requestedVersion: "",
			expected:         "",
			shouldUpdate:     false,
			expectErr:        true,
		},
		{
			name:             "missing rs.GitVersion still updates app",
			app:              makeApp("newVersion"),
			updater:          mockUpdater("", "newVersion"),
			rs:               &app.RegistryRefSpec{},
			requestedVersion: "",
			expected:         "newVersion",
			shouldUpdate:     true,
			expectErr:        false,
		},
	}

	for _, tc := range tests {
		result, err := doUpdateRegistry(tc.app, tc.updater, tc.rs, tc.requestedVersion)
		if tc.expectErr {
			require.Errorf(t, err, "test: %v", tc.name)
		} else {
			require.NoErrorf(t, err, "test: %v", tc.name)
		}

		// Don't make assertions about return value if error was returned
		if err != nil {
			continue
		}

		assert.Equal(t, tc.expected, result)

		// Assert app.UpdateRegistry gets called with the new, updated version,
		// when there is a newer version found.
		if tc.shouldUpdate {
			newRS := makeSpec(tc.expected)
			tc.app.AssertCalled(t, "UpdateRegistry", newRS)
		} else {
			tc.app.AssertNumberOfCalls(t, "UpdateRegistry", 0)
		}
	}
}

// func TestRegistryUpdate_doUpdate(t *testing.T) {
// 	a := new(amocks.App)
// 	spec := &app.RegistryRefSpec{
// 		GitVersion: &app.GitVersionSpec{
// 			CommitSHA: "oldversion"
// 		}
// 	}
// 	a.On("Registries").Return(
// 		app.RegistryRefSpecs{
// 			"incubator": spec,
// 		},
// 		nil,
// 	)
// 	ru := &RegistryUpdate{
// 		app:      a,
// 		locateFn: mockRegistryLocator("oldversion", "newversion"),
// 	}

// }

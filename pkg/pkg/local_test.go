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

package pkg

import (
	"fmt"
	"path/filepath"
	"testing"

	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func withLocalPackage(t *testing.T, name string, version string, useLegacyPath bool, fn func(a *amocks.App, fs afero.Fs)) {
	test.WithApp(t, "/app", func(a *amocks.App, fs afero.Fs) {
		a.On("VendorPath").Return("/app/vendor")

		var dstName = name
		if version != "" && !useLegacyPath {
			dstName = fmt.Sprintf("%s@%s", name, version)
		}
		srcPath := filepath.Join("incubator", name)
		dstPath := filepath.Join("/app", "vendor", "incubator", dstName)
		test.StageDir(t, fs, srcPath, dstPath)

		fn(a, fs)
	})
}

func TestLocal_New(t *testing.T) {
	tests := []struct {
		caseName   string
		registry   string
		name       string
		version    string
		srcPath    string
		targetPath string
		expectErr  bool
	}{
		{
			caseName:   "package versioned, vendor path versioned",
			registry:   "incubator",
			name:       "apache",
			version:    "1.2.3",
			srcPath:    "incubator/apache",
			targetPath: "/app/vendor/incubator/apache@1.2.3",
			expectErr:  false,
		},
		{
			caseName:   "package versioned, vendor path unversioned",
			registry:   "incubator",
			name:       "apache",
			version:    "1.2.3",
			srcPath:    "incubator/apache",
			targetPath: "/app/vendor/incubator/apache",
			expectErr:  false,
		},
		{
			caseName:   "package unversioned, vendor path unversioned",
			registry:   "incubator",
			name:       "apache",
			version:    "",
			srcPath:    "incubator/apache",
			targetPath: "/app/vendor/incubator/apache",
			expectErr:  false,
		},
		{
			caseName:   "package versioned, vendor path has wrong version",
			registry:   "incubator",
			name:       "apache",
			version:    "1.2.3",
			srcPath:    "incubator/apache",
			targetPath: "/app/vendor/incubator/apache@4.5.6",
			expectErr:  true,
		},
	}

	for _, tc := range tests {
		test.WithApp(t, "/app", func(a *amocks.App, fs afero.Fs) {
			a.On("VendorPath").Return("/app/vendor")

			test.StageDir(t, fs, tc.srcPath, tc.targetPath)
			_, err := NewLocal(a, tc.name, tc.registry, tc.version, nil)
			if tc.expectErr {
				require.Error(t, err, tc.caseName)
			} else {
				require.NoError(t, err, tc.caseName)
			}
		})
	}
}

func TestLocal_Name(t *testing.T) {
	withLocalPackage(t, "apache", "", false, func(a *amocks.App, fs afero.Fs) {
		l, err := NewLocal(a, "apache", "incubator", "", nil)
		require.NoError(t, err)

		require.Equal(t, "apache", l.Name())
	})
}

func TestLocal_RegistryName(t *testing.T) {
	withLocalPackage(t, "apache", "", false, func(a *amocks.App, fs afero.Fs) {
		l, err := NewLocal(a, "apache", "incubator", "", nil)
		require.NoError(t, err)

		require.Equal(t, "incubator", l.RegistryName())
	})
}

func TestLocal_IsInstalled(t *testing.T) {
	withLocalPackage(t, "apache", "", false, func(a *amocks.App, fs afero.Fs) {
		ic := &fakeInstallChecker{
			isInstalled: true,
		}

		l, err := NewLocal(a, "apache", "incubator", "", ic)
		require.NoError(t, err)

		i, err := l.IsInstalled()
		assert.NoError(t, err)
		assert.True(t, i)
	})
}

func TestLocal_Description(t *testing.T) {
	withLocalPackage(t, "apache", "", false, func(a *amocks.App, fs afero.Fs) {
		l, err := NewLocal(a, "apache", "incubator", "", nil)
		require.NoError(t, err)

		require.Equal(t, "part description", l.Description())
	})
}

func TestLocal_Prototypes(t *testing.T) {
	tests := []struct {
		caseName      string
		name          string
		version       string
		useLegacyPath bool
		expected      []string
	}{
		{
			caseName: "versioned path",
			name:     "apache",
			version:  "1.2.3",
			expected: []string{
				"io.ksonnet.pkg.apache-simple",
			},
		},
		{
			caseName:      "legacy path",
			name:          "apache",
			version:       "1.2.3",
			useLegacyPath: true,
			expected: []string{
				"io.ksonnet.pkg.apache-simple",
			},
		},
		{
			caseName: "unversioned",
			name:     "apache",
			version:  "",
			expected: []string{
				"io.ksonnet.pkg.apache-simple",
			},
		},
	}

	for _, tc := range tests {
		withLocalPackage(t, tc.name, tc.version, tc.useLegacyPath, func(a *amocks.App, fs afero.Fs) {
			l, err := NewLocal(a, tc.name, "incubator", tc.version, nil)
			require.NoErrorf(t, err, "[%v] NewLocal", tc.caseName)

			prototypes, err := l.Prototypes()
			require.NoErrorf(t, err, "[%v] Prototypes", tc.caseName)

			assert.Truef(t, len(prototypes) >= len(tc.expected), "[%v] length of prototypes was %d", tc.caseName, len(prototypes))

			protoNames := make([]string, 0, len(prototypes))
			for _, proto := range prototypes {
				protoNames = append(protoNames, proto.Name)
			}

			assert.Subsetf(t, protoNames, tc.expected, "[%v] comparing prototype names", tc.caseName)
		})
	}
}

func TestLocal_Path(t *testing.T) {
	tests := []struct {
		caseName      string
		registry      string
		name          string
		version       string
		useLegacyPath bool
		expected      string
	}{
		{
			caseName: "versioned package",
			registry: "incubator",
			name:     "apache",
			version:  "1.2.3",
			expected: filepath.FromSlash("/app/vendor/incubator/apache@1.2.3"),
		},
		{
			caseName: "unversioned package",
			registry: "incubator",
			name:     "apache",
			version:  "",
			expected: filepath.FromSlash("/app/vendor/incubator/apache"),
		},
		{
			caseName:      "versioned package, legacy path",
			registry:      "incubator",
			name:          "apache",
			version:       "1.2.3",
			useLegacyPath: true,
			expected:      filepath.FromSlash("/app/vendor/incubator/apache@1.2.3"),
		},
	}

	for _, tc := range tests {
		withLocalPackage(t, tc.name, tc.version, tc.useLegacyPath, func(a *amocks.App, fs afero.Fs) {
			l, err := NewLocal(a, tc.name, tc.registry, tc.version, nil)
			require.NoError(t, err, tc.caseName)

			actual := l.Path()
			require.Equal(t, tc.expected, actual, tc.caseName)
		})
	}
}

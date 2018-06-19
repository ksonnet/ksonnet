// Copyright 2017 The ksonnet authors
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
package lib

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/util/kslib"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	swaggerLocation  = "/blankSwagger.json"
	blankSwaggerData = `{
  "swagger": "2.0",
  "info": {
   "title": "Kubernetes",
   "version": "v1.7.0"
  },
  "paths": {
  },
  "definitions": {
  }
}`
)

func TestGenerateLibData(t *testing.T) {
	cases := []struct {
		name        string
		basePath    string
		generator   KsLibGenerator
		swaggerData []byte
	}{
		{
			name:     "use version path",
			basePath: "v1.7.0",
			generator: &fakeKsLibGenerator{
				ksonnetLib: &kslib.KsonnetLib{},
			},
			swaggerData: []byte(blankSwaggerData),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			afero.WriteFile(fs, swaggerLocation, tc.swaggerData, os.ModePerm)

			specFlag := fmt.Sprintf("file:%s", swaggerLocation)
			libPath := "lib"

			libManager, err := NewManager(specFlag, fs, libPath)
			require.NoError(t, err)

			libManager.generator = tc.generator

			err = libManager.GenerateLibData()
			require.NoError(t, err)

			// Verify contents of lib.
			genPath := filepath.Join(libPath, KsonnetLibHome, tc.basePath)

			checkKsLib(t, fs, genPath)
		})
	}
}

func checkKsLib(t *testing.T, fs afero.Fs, path string) {
	files := []string{"swagger.json", "k.libsonnet", "k8s.libsonnet"}
	for _, f := range files {
		p := filepath.Join(path, f)
		exists, err := afero.Exists(fs, p)
		assert.NoError(t, err, p)
		assert.True(t, exists, "%q did not exist", p)
	}
}

func TestManager_GetLibPath(t *testing.T) {
	cases := []struct {
		name     string
		initFs   func(*testing.T, string, string) afero.Fs
		expected string
	}{
		{
			name: "with ksonnet-lib",
			initFs: func(t *testing.T, version, libPath string) afero.Fs {
				fs := afero.NewMemMapFs()
				klPath := filepath.Join(libPath, KsonnetLibHome, version)
				err := fs.MkdirAll(klPath, 0755)
				require.NoError(t, err)

				return fs
			},
			expected: filepath.FromSlash("lib/ksonnet-lib/v1.10.3"),
		},
		{
			name: "without ksonnet-lib",
			initFs: func(t *testing.T, version, libPath string) afero.Fs {
				fs := afero.NewMemMapFs()
				klPath := filepath.Join(libPath, version)
				err := fs.MkdirAll(klPath, 0755)
				require.NoError(t, err)

				return fs
			},
			expected: filepath.FromSlash("lib/v1.10.3"),
		},
		{
			name: "doesn't already exist",
			initFs: func(t *testing.T, version, libPath string) afero.Fs {
				fs := afero.NewMemMapFs()
				return fs
			},
			expected: filepath.FromSlash("lib/ksonnet-lib/v1.10.3"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			specFlag := "version:v1.10.3"
			libPath := "lib"

			fs := tc.initFs(t, "v1.10.3", "lib")

			libManager, err := NewManager(specFlag, fs, libPath)
			require.NoError(t, err)

			got, err := libManager.GetLibPath()
			require.NoError(t, err)

			require.Equal(t, tc.expected, got)
		})
	}
}

type fakeKsLibGenerator struct {
	ksonnetLib *kslib.KsonnetLib
	err        error
}

var _ (KsLibGenerator) = (*fakeKsLibGenerator)(nil)

func (g *fakeKsLibGenerator) Generate() (*kslib.KsonnetLib, error) {
	return g.ksonnetLib, g.err
}

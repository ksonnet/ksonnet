// Copyright 2017 The kubecfg authors
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
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ksonnet/ksonnet/metadata/app"
)

var (
	existingPaths = []string{
		"/app.yaml",
		"/components/a.jsonnet",
		"/components/b.jsonnet",
		"/components/other",
		"/components/params.libsonnet",
		"/components/nested/a.jsonnet",
		"/components/nested/params.libsonnet",
		"/components/nested/very/deeply/c.jsonnet",
		"/components/nested/very/deeply/params.libsonnet",
		"/components/shallow/c.jsonnet",
		"/components/shallow/params.libsonnet",
	}

	invalidPaths = []string{
		"/app.yaml",
		"/components/a.jsonnet",
		"/components/a.txt",
	}
)

type stubAppSpecer struct {
	appSpec *app.Spec
	err     error
}

var _ AppSpecer = (*stubAppSpecer)(nil)

func newStubAppSpecer(appSpec *app.Spec) *stubAppSpecer {
	return &stubAppSpecer{appSpec: appSpec}
}

func (s *stubAppSpecer) AppSpec() (*app.Spec, error) {
	return s.appSpec, s.err
}

func makePaths(t *testing.T, fs afero.Fs, paths []string) {
	for _, path := range paths {
		dir := filepath.Dir(path)
		err := fs.MkdirAll(dir, 0755)
		require.NoError(t, err)

		_, err = fs.Create(path)
		require.NoError(t, err)
	}
}

func TestPath(t *testing.T) {

	cases := []struct {
		name     string
		paths    []string
		in       string
		expected string
		isErr    bool
	}{
		{
			name:     "in root namespace",
			paths:    existingPaths,
			in:       "a",
			expected: "/components/a.jsonnet",
		},
		{
			name:     "in nested namespace",
			paths:    existingPaths,
			in:       "nested/a",
			expected: "/components/nested/a.jsonnet",
		},
		{
			name:  "not found",
			paths: existingPaths,
			in:    "z",
			isErr: true,
		},
		{
			name:  "invalid path",
			paths: invalidPaths,
			in:    "a",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			makePaths(t, fs, tc.paths)

			path, err := Path(fs, "/", tc.in)
			if tc.isErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				assert.Equal(t, tc.expected, path)
			}
		})
	}
}

func TestExtractNamedspacedComponent(t *testing.T) {
	cases := []struct {
		name      string
		path      string
		nsPath    string
		component string
	}{
		{
			name:      "component in root namespace",
			path:      "my-deployment",
			nsPath:    "",
			component: "my-deployment",
		},
		{
			name:      "component in root namespace",
			path:      "nested/my-deployment",
			nsPath:    "nested",
			component: "my-deployment",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			root := "/"
			ns, component := ExtractNamespacedComponent(fs, root, tc.path)
			assert.Equal(t, tc.nsPath, ns.Path)
			assert.Equal(t, component, component)
		})
	}
}

func TestNamespace_ParamsPath(t *testing.T) {
	cases := []struct {
		name     string
		nsName   string
		expected string
	}{
		{
			name:     "root namespace",
			expected: "/components/params.libsonnet",
		},
		{
			name:     "nested namespace",
			nsName:   "nested",
			expected: "/components/nested/params.libsonnet",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ns := Namespace{Path: tc.nsName, root: "/"}
			assert.Equal(t, tc.expected, ns.ParamsPath())
		})
	}
}

func TestNamespace_ComponentPaths(t *testing.T) {
	cases := []struct {
		name     string
		nsPath   string
		expected []string
	}{
		{
			name: "root namespace",
			expected: []string{
				"/components/a.jsonnet",
				"/components/b.jsonnet",
			},
		},
		{
			name:   "nested namespace",
			nsPath: "nested",
			expected: []string{
				"/components/nested/a.jsonnet",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			makePaths(t, fs, existingPaths)

			ns := Namespace{Path: tc.nsPath, fs: fs, root: "/"}

			paths, err := ns.ComponentPaths()
			require.NoError(t, err)

			require.Equal(t, tc.expected, paths)
		})
	}
}

func TestNamespace_Components(t *testing.T) {
	cases := []struct {
		name     string
		nsPath   string
		expected []string
	}{
		{
			name: "root namespace",
			expected: []string{
				"a",
				"b",
			},
		},
		{
			name:   "nested namespace",
			nsPath: "nested",
			expected: []string{
				"nested/a",
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			makePaths(t, fs, existingPaths)

			ns := Namespace{Path: tc.nsPath, fs: fs, root: "/"}

			paths, err := ns.Components()
			require.NoError(t, err)

			require.Equal(t, tc.expected, paths)
		})
	}
}

func TestNamespaces(t *testing.T) {
	fs := afero.NewMemMapFs()

	makePaths(t, fs, existingPaths)

	namespaces, err := Namespaces(fs, "/")
	require.NoError(t, err)

	expected := []Namespace{
		{Path: "", fs: fs, root: "/"},
		{Path: "nested", fs: fs, root: "/"},
		{Path: "nested/very/deeply", fs: fs, root: "/"},
		{Path: "shallow", fs: fs, root: "/"},
	}

	assert.Equal(t, expected, namespaces)
}

func TestMakePathsByNameSpace(t *testing.T) {
	fs := afero.NewMemMapFs()
	makePaths(t, fs, existingPaths)

	cases := []struct {
		name     string
		targets  []string
		expected map[Namespace][]string
		isErr    bool
	}{
		{
			name: "no target paths",
			expected: map[Namespace][]string{
				Namespace{fs: fs, root: "/"}: []string{
					"/components/a.jsonnet",
					"/components/b.jsonnet",
				},
				Namespace{fs: fs, root: "/", Path: "nested"}: []string{
					"/components/nested/a.jsonnet",
				},
				Namespace{fs: fs, root: "/", Path: "nested/very/deeply"}: []string{
					"/components/nested/very/deeply/c.jsonnet",
				},
				Namespace{fs: fs, root: "/", Path: "shallow"}: []string{
					"/components/shallow/c.jsonnet",
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			envSpec := &app.EnvironmentSpec{
				Targets: tc.targets,
			}

			appSpec := &app.Spec{
				Environments: app.EnvironmentSpecs{"default": envSpec},
			}
			appSpecer := newStubAppSpecer(appSpec)

			root := "/"
			env := "default"

			paths, err := MakePathsByNamespace(fs, appSpecer, root, env)
			if tc.isErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, paths)
			}

		})
	}
}

func TestMakePaths(t *testing.T) {

	cases := []struct {
		name     string
		targets  []string
		expected []string
		isErr    bool
	}{
		{
			name: "no target paths",
			expected: []string{
				"/components/a.jsonnet",
				"/components/b.jsonnet",
				"/components/nested/a.jsonnet",
				"/components/nested/very/deeply/c.jsonnet",
				"/components/shallow/c.jsonnet",
			},
		},
		{
			name: "jsonnet target path file",
			targets: []string{
				"a.jsonnet",
			},
			expected: []string{"/components/a.jsonnet"},
		},
		{
			name: "jsonnet target path dir",
			targets: []string{
				"nested",
			},
			expected: []string{
				"/components/nested/a.jsonnet",
				"/components/nested/very/deeply/c.jsonnet",
			},
		},
		{
			name: "jsonnet target path dir and files",
			targets: []string{
				"shallow/c.jsonnet",
				"nested",
			},
			expected: []string{
				"/components/nested/a.jsonnet",
				"/components/nested/very/deeply/c.jsonnet",
				"/components/shallow/c.jsonnet",
			},
		},
		{
			name:    "target points to missing path",
			targets: []string{"missing"},
			isErr:   true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			makePaths(t, fs, existingPaths)

			envSpec := &app.EnvironmentSpec{
				Targets: tc.targets,
			}

			appSpec := &app.Spec{
				Environments: app.EnvironmentSpecs{"default": envSpec},
			}
			appSpecer := newStubAppSpecer(appSpec)

			root := "/"
			env := "default"

			paths, err := MakePaths(fs, appSpecer, root, env)
			if tc.isErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, paths)
			}
		})
	}
}

func TestMakePaths_invalid_appSpecer(t *testing.T) {
	fs := afero.NewMemMapFs()
	_, err := MakePaths(fs, nil, "/", "default")
	require.Error(t, err)
}

func TestMakePaths_invalid_fs(t *testing.T) {
	appSpecer := newStubAppSpecer(nil)
	_, err := MakePaths(nil, appSpecer, "/", "default")
	require.Error(t, err)
}

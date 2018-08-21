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
	"os"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/pkg"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

type removeAllFn func(path string) error
type fakeRemover struct {
	removeAllFn removeAllFn
}

func (f fakeRemover) RemoveAll(path string) error {
	return f.removeAllFn(path)
}

type fakeVendorPathResolver struct {
	pathFn        func(d pkg.Descriptor) (string, error)
	isInstalledFn func(d pkg.Descriptor) (bool, error)
}

func (f fakeVendorPathResolver) VendorPath(d pkg.Descriptor) (string, error) {
	return f.pathFn(d)
}

func (f fakeVendorPathResolver) IsInstalled(d pkg.Descriptor) (bool, error) {
	return f.isInstalledFn(d)
}

func makeVendorPathResolver(t *testing.T, desc pkg.Descriptor, isInstalled bool, path string) fakeVendorPathResolver {
	return fakeVendorPathResolver{
		pathFn: func(d pkg.Descriptor) (string, error) {
			assert.Equal(t, desc, d)
			return path, nil
		},
		isInstalledFn: func(d pkg.Descriptor) (bool, error) {
			assert.Equal(t, desc, d)
			return isInstalled, nil
		},
	}
}

func TestGarbageCollector_RemoveOrphans(t *testing.T) {
	tests := []struct {
		name             string
		d                pkg.Descriptor
		pkgPath          string
		isInstalled      bool
		expectCallRemove bool
		wantErr          bool
	}{
		{
			name: "should remove",
			d: pkg.Descriptor{
				Name:    "nginx",
				Version: "1.2.3",
			},
			pkgPath:          "/app/vendor/nginx@1.2.3",
			isInstalled:      false,
			expectCallRemove: true,
		},
		{
			name: "should not remove",
			d: pkg.Descriptor{
				Name:    "nginx",
				Version: "1.2.3",
			},
			isInstalled:      false,
			expectCallRemove: false,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var called bool
			remover := fakeRemover{
				removeAllFn: func(path string) error {
					assert.Equal(t, tc.pkgPath, path)
					called = true
					return nil
				},
			}
			finderChecker := makeVendorPathResolver(t, tc.d, tc.isInstalled, tc.pkgPath)
			gc := GarbageCollector{
				pkgManager: finderChecker,
				fs:         remover,
			}
			if err := gc.RemoveOrphans(tc.d); (err != nil) != tc.wantErr {
				t.Errorf("GarbageCollector.RemoveOrphans() error = %v, wantErr %v", err, tc.wantErr)
			}
			assert.Equal(t, tc.expectCallRemove, called)
		})
	}
}

func Test_removeEmptyParents(t *testing.T) {
	tests := []struct {
		name        string
		stagePaths  []string
		root        string
		path        string
		expectPaths []string
		expectErr   bool
	}{
		{
			name: "parent not empty",
			root: "/app/vendor",
			path: "/app/vendor/incubator/nginx@1.2.3",
			stagePaths: []string{
				"/app/vendor/incubator/mysql@4.5.6",
			},
			expectPaths: []string{
				"/app/vendor/incubator/mysql@4.5.6",
			},
		},
		{
			name: "parent empty",
			root: "/app/vendor",
			path: "/app/vendor/incubator/nginx@1.2.3",
			stagePaths: []string{
				"/app/vendor/incubator",
			},
			expectPaths: []string{
				"/app/vendor",
			},
		},
		{
			name: "grandparent not empty",
			root: "/app/vendor",
			path: "/app/vendor/helm-stable/mysql/helm/0.9.3/mysql",
			stagePaths: []string{
				"/app/vendor/helm-stable/mysql/helm/0.9.3",
				"/app/vendor/helm-stable/nginx/helm/1.2.3",
			},
			expectPaths: []string{
				"/app/vendor/helm-stable/nginx/helm/1.2.3",
			},
		},
		{
			name: "grandparent empty",
			root: "/app/vendor",
			path: "/app/vendor/helm-stable/mysql/helm/0.9.3/mysql",
			stagePaths: []string{
				"/app/vendor/helm-stable/mysql/helm/0.9.3",
			},
			expectPaths: []string{
				"/app/vendor",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := afero.NewMemMapFs()
			for _, path := range tc.stagePaths {
				fs.MkdirAll(path, os.FileMode(0755))
			}

			err := removeEmptyParents(fs, tc.path, tc.root)
			if (err != nil) != tc.expectErr {
				t.Errorf("GarbageCollector.removeEmptyParents() error = %v, wantErr %v", err, tc.expectErr)
				return
			}
			if err != nil {
				return
			}

			test.AssertExpectedPaths(t, fs, tc.root, tc.expectPaths)
		})
	}
}

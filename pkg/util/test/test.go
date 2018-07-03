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

package test

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/pkg/errors"
	godiff "github.com/shazow/go-diff"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// ReadTestData reads a file from `testdata` and returns it as a string.
func ReadTestData(t *testing.T, name string) string {
	path := filepath.Join("testdata", name)
	data, err := ioutil.ReadFile(path)
	require.NoError(t, err)

	return string(data)
}

// StageFile stages a file on on the provided filesystem from
// testdata.
func StageFile(t *testing.T, fs afero.Fs, src, dest string) {
	in := filepath.Join("testdata", src)

	b, err := ioutil.ReadFile(in)
	require.NoError(t, err)

	dir := filepath.Dir(dest)
	err = fs.MkdirAll(dir, 0755)
	require.NoError(t, err)

	err = afero.WriteFile(fs, dest, b, 0644)
	require.NoError(t, err)
}

// StageDir stages a directory on the provided filesystem from
// testdata.
func StageDir(t *testing.T, fs afero.Fs, src, dest string) {
	root, err := filepath.Abs(filepath.Join("testdata", src))
	require.NoError(t, err, "finding absolute path")

	err = filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "walk received an error")
		}

		cur := filepath.Join(dest, strings.TrimPrefix(path, root))
		if fi.IsDir() {
			return fs.MkdirAll(cur, 0755)
		}

		copyFile(fs, path, cur)
		return nil
	})

	require.NoError(t, err)
}

func copyFile(fs afero.Fs, src, dest string) error {
	from, err := os.Open(src)
	if err != nil {
		return errors.Wrapf(err, "opening %q for copy", src)
	}
	defer from.Close()

	to, err := fs.OpenFile(dest, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return errors.Wrapf(err, "opening %q for write", dest)
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	return err
}

// WithApp runs an enclosure with a mocked app and fs.
func WithApp(t *testing.T, root string, fn func(*mocks.App, afero.Fs)) {
	fs := afero.NewMemMapFs()

	WithAppFs(t, root, fs, fn)
}

// WithAppFs runs an enclosure with a mocked app and fs. Allow supplying the fs.
func WithAppFs(t *testing.T, root string, fs afero.Fs, fn func(*mocks.App, afero.Fs)) {
	a := &mocks.App{}
	a.On("Fs").Return(fs)
	a.On("Root").Return(root)
	a.On("LibPath", mock.AnythingOfType("string")).Return(filepath.Join(root, "lib", "v1.8.7"), nil)

	fn(a, fs)
}

// AssertOutput asserts the output matches the actual contents
func AssertOutput(t *testing.T, filename, actual string) {
	path := filepath.Join("testdata", filepath.FromSlash(filename))
	f, err := os.Open(path)
	require.NoError(t, err)
	defer f.Close()

	rActual := strings.NewReader(actual)

	var buf bytes.Buffer
	err = godiff.DefaultDiffer().Diff(&buf, f, rActual)
	require.NoError(t, err)
	require.Empty(t, buf.String())
}

// dumpFs logs the contents of an afero.Fs virtual filesystem interface.
func DumpFs(t *testing.T, fs afero.Fs) {
	if fs == nil {
		return
	}

	err := afero.Walk(fs, "/", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			t.Logf("%v", err)
			return err
		}

		t.Logf("%s", path)

		return nil
	})
	if err != nil {
		t.Logf("%v", err)
	}
}

// AssertDirectoriesMatch compares the contents of two directories (recursively) and asserts they are equivelent.
func AssertDirectoriesMatch(t *testing.T, fs afero.Fs, rootA string, rootB string) {
	// Map a path from root a to root b
	// e.g. rootA == /foo, rootB == /bar, mapAToB(/foo/something/nice) -> /bar/something/nice
	// Returns error if path is not rooted under rootA.
	type mapFunc func(string) (string, error)
	mapper := func(s, rootA, rootB string) (string, error) {
		a := filepath.Clean(rootA)
		b := filepath.Clean(rootB)
		p := filepath.Clean(s)

		if !strings.HasPrefix(p, a) { // TODO compare actual path segments
			return "", errors.Errorf("%v not rooted under %v", s, rootA)
		}

		unrooted := strings.TrimPrefix(p, a)
		return filepath.Join(b, unrooted), nil
	}
	mapAToB := func(s string) (string, error) {
		return mapper(s, rootA, rootB)
	}
	mapBToA := func(s string) (string, error) {
		return mapper(s, rootB, rootA)
	}
	walker := func(mapper mapFunc, diffSet map[string]struct{}, compareFiles bool, path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.Wrap(err, "bad root path")
		}
		pathB, err := mapper(path)
		if err != nil {
			return errors.Wrapf(err, "mapping %v", path)
		}

		ok, err := afero.Exists(fs, pathB)
		if err != nil {
			return errors.Wrapf(err, "checking existence: %v", pathB)
		}
		if !ok {
			diffSet[path] = struct{}{}
			return nil
		}

		// For contents, only compare files
		if info.IsDir() {
			return nil
		}

		// Exists in both trees, compare file contents
		fileA, err := fs.Open(path)
		defer func(f io.Closer) {
			if f != nil {
				f.Close()
			}
		}(fileA)
		if err != nil {
			return errors.Wrapf(err, "opening file %v", path)
		}
		fileB, err := fs.Open(pathB)
		defer func(f io.Closer) {
			if f != nil {
				f.Close()
			}
		}(fileB)
		if err != nil {
			return errors.Wrapf(err, "opening file %v", pathB)
		}

		var buf bytes.Buffer
		err = godiff.DefaultDiffer().Diff(&buf, fileA, fileB)
		if err != nil {
			return errors.Wrapf(err, "comparing %v and %v", path, pathB)
		}

		// Assert files matched
		assert.Empty(t, buf.String())
		return nil
	} // end walker

	diffSet := make(map[string]struct{})

	err := afero.Walk(fs, rootA, func(path string, info os.FileInfo, err error) error {
		return walker(mapAToB, diffSet, true, path, info, err)
	})
	assert.NoError(t, err, "walking %v", rootA)

	err = afero.Walk(fs, rootB, func(path string, info os.FileInfo, err error) error {
		return walker(mapBToA, diffSet, false, path, info, err)
	})
	assert.NoError(t, err, "walking %v", rootB)

	assert.Empty(t, diffSet, "file set symmetric_difference")
}

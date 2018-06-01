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

package archive

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_UnTgz(t *testing.T) {
	f, err := os.Open(filepath.ToSlash("testdata/mysql-0.6.0.tgz"))
	require.NoError(t, err)

	paths := make(map[string]bool)

	handler := func(tf *File) error {
		paths[tf.Name] = true
		return nil
	}

	expected := map[string]bool{
		"mysql/Chart.yaml":                                   true,
		"mysql/values.yaml":                                  true,
		"mysql/templates/NOTES.txt":                          true,
		"mysql/templates/_helpers.tpl":                       true,
		"mysql/templates/configurationFiles-configmap.yaml":  true,
		"mysql/templates/deployment.yaml":                    true,
		"mysql/templates/initializationFiles-configmap.yaml": true,
		"mysql/templates/pvc.yaml":                           true,
		"mysql/templates/secrets.yaml":                       true,
		"mysql/templates/svc.yaml":                           true,
		"mysql/templates/tests/test-configmap.yaml":          true,
		"mysql/templates/tests/test.yaml":                    true,
		"mysql/.helmignore":                                  true,
		"mysql/README.md":                                    true,
	}

	tgz := &Tgz{}
	err = tgz.Unarchive(f, handler)

	require.NoError(t, err)
	require.Equal(t, expected, paths)
}

func Test_UnTgz_invalid_reader(t *testing.T) {
	handler := func(tf *File) error {
		return nil
	}

	r := strings.NewReader("not gzip")

	tgz := &Tgz{}
	err := tgz.Unarchive(r, handler)
	require.Error(t, err)
}

func Test_UnTgz_nil_reader(t *testing.T) {
	handler := func(tf *File) error {
		return nil
	}

	tgz := &Tgz{}
	err := tgz.Unarchive(nil, handler)
	require.Error(t, err)
}

func Test_UnTgz_not_invalid_tar(t *testing.T) {
	handler := func(tf *File) error {
		return nil
	}

	f, err := os.Open(filepath.ToSlash("testdata/foo.txt.gz"))
	require.NoError(t, err)

	tgz := &Tgz{}
	err = tgz.Unarchive(f, handler)
	require.Error(t, err)
}

func Test_UnTgz_handler_failed(t *testing.T) {
	handler := func(tf *File) error {
		return errors.New("fail")
	}

	f, err := os.Open(filepath.ToSlash("testdata/mysql-0.6.0.tgz"))
	require.NoError(t, err)

	tgz := &Tgz{}
	err = tgz.Unarchive(f, handler)
	require.Error(t, err)
}

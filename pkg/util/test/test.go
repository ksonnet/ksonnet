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
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

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

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

package params

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/stretchr/testify/require"
)

func Test_PatchJSON(t *testing.T) {
	jsonObject, err := ioutil.ReadFile(filepath.Join("testdata", "rbac-1.json"))
	require.NoError(t, err)

	patch, err := ioutil.ReadFile(filepath.Join("testdata", "patch.json"))
	require.NoError(t, err)

	got, err := PatchJSON(string(jsonObject), string(patch), "rbac-1")
	require.NoError(t, err)

	test.AssertOutput(t, "rbac-1-patched.json", got)
}

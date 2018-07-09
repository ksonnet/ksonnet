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

package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSort(t *testing.T) {
	versionNames := []string{
		"6.5.1",
		"0.1.3",
		"11.3",
	}

	var versions []Version

	for _, s := range versionNames {
		v, err := Make(s)
		require.NoError(t, err)

		versions = append(versions, v)
	}

	Sort(versions)

	var got []string
	for _, v := range versions {
		got = append(got, v.String())
	}

	expected := []string{"0.1.3", "6.5.1", "11.3"}

	assert.Equal(t, expected, got)
}

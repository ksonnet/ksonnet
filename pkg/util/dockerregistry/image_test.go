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

package dockerregistry

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImageName_String(t *testing.T) {
	cases := []struct {
		name     string
		expected string
	}{
		{
			name:     "foo/bar",
			expected: "foo/bar:latest",
		},
		{
			name:     "foo/bar@sha256:abcde",
			expected: "foo/bar@sha256:abcde",
		},
		{
			name:     "myregistry:5000/foo/bar",
			expected: "myregistry:5000/foo/bar:latest",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in, err := ParseImageName(tc.name)
			require.NoError(t, err)

			got := in.String()

			require.Equal(t, tc.expected, got)
		})
	}
}

func TestImageName_RegistryRepoName(t *testing.T) {
	cases := []struct {
		name     string
		repoName string
		expected string
	}{
		{
			name:     "with repo name",
			repoName: "foo",
			expected: "foo/bar",
		},
		{
			name:     "without repo name",
			expected: "library/bar",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := ImageName{
				Repository: tc.repoName,
				Name:       "bar",
			}

			got := in.RegistryRepoName()

			require.Equal(t, tc.expected, got)
		})
	}
}

func TestImageName_RegistryURL(t *testing.T) {
	cases := []struct {
		name        string
		registryURL string
		expected    string
	}{
		{
			name:        "with repo name",
			registryURL: "myregistry:5000",
			expected:    "https://myregistry:5000",
		},
		{
			name:     "without repo name",
			expected: "https://registry-1.docker.io",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			in := ImageName{
				Registry: tc.registryURL,
			}

			got := in.RegistryURL()

			require.Equal(t, tc.expected, got)
		})
	}
}

func TestParseImageName(t *testing.T) {
	cases := []struct {
		name     string
		expected string
		isErr    bool
	}{
		{
			name:     "foo",
			expected: "foo:latest",
		},
		{
			name:     "foo:latest",
			expected: "foo:latest",
		},
		{
			name:     "foo/bar",
			expected: "foo/bar:latest",
		},
		{
			name:     "foo/bar@sha256:abcded",
			expected: "foo/bar@sha256:abcded",
		},
		{
			name:     "foo/bar/baz",
			expected: "foo/bar/baz:latest",
		},
		{
			name:  "foo/bar/baz/qux",
			isErr: true,
		},
		{
			name:  "foo/bar::latest",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseImageName(tc.name)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tc.expected, got.String())
		})
	}
}

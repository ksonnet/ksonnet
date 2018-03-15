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

package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_GetAPISpec(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/swagger.json")
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(b))
	}))
	defer ts.Close()

	cases := []struct {
		name      string
		serverURL string
		expected  string
	}{
		{
			name:      "invalid server URL",
			serverURL: "http://+++",
			expected:  defaultVersion,
		},
		{
			name:      "with a server",
			serverURL: ts.URL,
			expected:  "version:v1.9.3",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := Config{}
			got := c.GetAPISpec(tc.serverURL)
			require.Equal(t, tc.expected, got)
		})
	}

}

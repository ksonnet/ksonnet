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
	"crypto/tls"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_defaultDigester_ManifestV2Digest(t *testing.T) {
	cases := []struct {
		name       string
		handler    http.HandlerFunc
		expectedFn func(*url.URL) string
		isErr      bool
	}{
		{
			name: "in general",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Docker-Content-Digest", "sha256:abcde")
			},
			expectedFn: func(u *url.URL) string {
				return fmt.Sprintf("%s%s@%s", u.Host, "/foo/bar", "sha256:abcde")
			},
		},
		{
			name: "image not found",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			},
			isErr: true,
		},
		{
			name: "unknown error",
			handler: func(w http.ResponseWriter, r *http.Request) {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ts := httptest.NewTLSServer(tc.handler)
			defer ts.Close()

			d := NewDefaultDigester()
			d.clientFactory = func() *http.Client {
				tr := &http.Transport{
					TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
				}

				c := &http.Client{
					Timeout:   1 * time.Second,
					Transport: tr,
				}

				return c
			}

			u, err := url.Parse(ts.URL)
			require.NoError(t, err)

			u.Path = "/foo/bar:latest"

			imageName := fmt.Sprintf("%s%s", u.Host, u.Path)

			digest, err := d.ManifestV2Digest(imageName)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tc.expectedFn(u), digest)
		})
	}
}

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

package github

import (
	"errors"
	"net/http"
	"net/url"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_defaultGitHub_ValidateURL(t *testing.T) {
	cases := []struct {
		name     string
		url      string
		c        httpClient
		urlParse func(string) (*url.URL, error)
		isErr    bool
	}{
		{
			name: "url exists",
			url:  "https://github.com/ksonnet/parts",
			c: &fakeHTTPClient{
				headResp: &http.Response{
					Status:     http.StatusText(http.StatusOK),
					StatusCode: http.StatusOK,
				},
			},
		},
		{
			name: "hostname and path",
			url:  "github.com/ksonnet/parts",
			c: &fakeHTTPClient{
				headResp: &http.Response{
					Status:     http.StatusText(http.StatusOK),
					StatusCode: http.StatusOK,
				},
			},
		},
		{
			name: "client failure",
			c: &fakeHTTPClient{
				headErr: errors.New("failed"),
			},
			isErr: true,
		},
		{
			name: "invalid status code",
			c: &fakeHTTPClient{
				headResp: &http.Response{
					Status:     http.StatusText(http.StatusNotFound),
					StatusCode: http.StatusNotFound,
				},
			},
			isErr: true,
		},
		{
			name:     "url parse failed",
			urlParse: func(string) (*url.URL, error) { return nil, errors.New("fail") },
			isErr:    true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.urlParse == nil {
				tc.urlParse = url.Parse
			}

			dg := defaultGitHub{
				httpClient: tc.c,
				urlParse:   tc.urlParse,
			}

			err := dg.ValidateURL(tc.url)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}

type fakeHTTPClient struct {
	headResp *http.Response
	headErr  error
}

func (c *fakeHTTPClient) Head(string) (*http.Response, error) {
	return c.headResp, c.headErr
}

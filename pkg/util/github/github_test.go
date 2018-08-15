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
	"context"
	"errors"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeTransport struct {
	resp *http.Response
	err  error
}

func (f *fakeTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f.resp, f.err
}

func fakeHTTPClient(resp *http.Response, err error) *http.Client {
	c := &http.Client{
		Transport: &fakeTransport{
			resp: resp,
			err:  err,
		},
	}
	return c
}
func Test_defaultGitHub_ValidateURL(t *testing.T) {
	cases := []struct {
		name     string
		url      string
		c        *http.Client
		urlParse func(string) (*url.URL, error)
		isErr    bool
	}{
		{
			name: "url exists",
			url:  "https://github.com/ksonnet/parts",
			c: fakeHTTPClient(
				&http.Response{
					Status:     http.StatusText(http.StatusOK),
					StatusCode: http.StatusOK,
				}, nil,
			),
		},
		{
			name: "hostname and path",
			url:  "github.com/ksonnet/parts",
			c: fakeHTTPClient(
				&http.Response{
					Status:     http.StatusText(http.StatusOK),
					StatusCode: http.StatusOK,
				}, nil,
			),
		},
		{
			name: "client failure",
			c: fakeHTTPClient(
				nil, errors.New("failed"),
			),
			isErr: true,
		},
		{
			name: "invalid status code",
			c: fakeHTTPClient(
				&http.Response{
					Status:     http.StatusText(http.StatusNotFound),
					StatusCode: http.StatusNotFound,
				}, nil,
			),
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

type mockTransport struct {
	roundTrip func(req *http.Request) (*http.Response, error)
}

func (f *mockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return f.roundTrip(req)
}

// Ensure httpClient propogates into vendor GitHub client
func Test_defaultGitHub_client(t *testing.T) {
	var called bool
	transport := &mockTransport{
		roundTrip: func(req *http.Request) (*http.Response, error) {
			called = true
			return nil, errors.New("N/A")
		},
	}
	httpClient := &http.Client{
		Transport: transport,
	}
	wrapper := NewGitHub(httpClient)
	dgh, ok := wrapper.(*defaultGitHub)
	require.Truef(t, ok, "unexpected type: %T", wrapper)

	os.Setenv("GITHUB_TOKEN", "")
	github := dgh.client()
	ctx := context.Background()
	_, _, _ = github.Repositories.GetCommitSHA1(ctx, "ksonnet", "ksonnet", "master", "")
	assert.True(t, called, "custom http client not called")
	called = false

	// Test with GITHUB_TOKEN
	os.Setenv("GITHUB_TOKEN", "foobar")
	github = dgh.client()
	_, _, _ = github.Repositories.GetCommitSHA1(ctx, "ksonnet", "ksonnet", "master", "")
	assert.True(t, called, "custom http client not called (with GITHUB_TOKEN)")
}

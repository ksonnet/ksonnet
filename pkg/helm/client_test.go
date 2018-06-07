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

package helm

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_normalizeHelmURI(t *testing.T) {
	cases := []struct {
		name     string
		uri      string
		expected string
		isErr    bool
	}{
		{
			name:     "host with index.yaml",
			uri:      "http://host/index.yaml",
			expected: "http://host/index.yaml",
		},
		{
			name:     "host with index.yaml in nested path",
			uri:      "http://host/nested/index.yaml",
			expected: "http://host/nested/index.yaml",
		},
		{
			name:     "host with no index.yaml",
			uri:      "http://host",
			expected: "http://host/index.yaml",
		},
		{
			name:     "host with path and no index.yaml",
			uri:      "http://host/nested",
			expected: "http://host/nested/index.yaml",
		},
		{
			name:  "invalid URL",
			uri:   "ht tp://host",
			isErr: true,
		},
		{
			name:  "not a URL",
			uri:   "invalid",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := normalizeHelmURI(tc.uri)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tc.expected, got)
		})
	}
}

func TestRepository_Latest(t *testing.T) {
	r := Repository{
		Charts: map[string][]RepositoryChart{
			"app-a": []RepositoryChart{
				{
					Name:    "app-a",
					Version: "0.2.0",
				},
				{
					Name:    "app-a",
					Version: "0.3.0",
				},
			},
		},
	}

	got := r.Latest()

	expected := []RepositoryChart{{Name: "app-a", Version: "0.3.0"}}

	require.Equal(t, expected, got)
}

func Test_httpGetter_Get(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "response")
	}))

	defer ts.Close()

	g := newHTTPGetter()

	r, err := g.Get(ts.URL)
	require.NoError(t, err)

	b, err := ioutil.ReadAll(r.Body)
	require.NoError(t, err)
	defer r.Body.Close()

	assert.Equal(t, http.StatusOK, r.StatusCode)
	assert.Equal(t, "response", string(b))
}

func genCharts() RepositoryCharts {
	return RepositoryCharts{
		{
			Name:    "app-a",
			Version: "0.2.0",
		},
		{
			Name:    "app-a",
			Version: "0.3.0",
		},
	}
}

func TestRepositoryCharts_Sort_Len(t *testing.T) {
	assert.Equal(t, 2, genCharts().Len())
}

func TestRepositoryCharts_Sort_Swap(t *testing.T) {
	charts := genCharts()

	charts.Swap(0, 1)
	expected := RepositoryCharts{
		{
			Name:    "app-a",
			Version: "0.3.0",
		},
		{
			Name:    "app-a",
			Version: "0.2.0",
		},
	}

	assert.Equal(t, expected, charts)
}

func TestRepositoryCharts_Sort_Less(t *testing.T) {
	charts := RepositoryCharts{
		{
			Name:    "app-a",
			Version: "0.3.0",
		},
		{
			Name:    "app-a",
			Version: "0.2.0",
		},
		{
			Name: "app-a",
		},
	}

	cases := []struct {
		name     string
		i        int
		j        int
		expected bool
	}{
		{
			name:     "i version is greater than j",
			i:        0,
			j:        1,
			expected: true,
		},
		{
			name:     "i version is less than j",
			i:        1,
			j:        0,
			expected: false,
		},
		{
			name:     "i version is equal to j",
			i:        1,
			j:        1,
			expected: false,
		},
		{
			name:     "i version is invalid",
			i:        2,
			j:        1,
			expected: false,
		},
		{
			name:     "j version is invalid",
			i:        1,
			j:        2,
			expected: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := charts.Less(tc.i, tc.j)
			require.Equal(t, tc.expected, got)
		})
	}
}

func Test_repositoryClient_Repository(t *testing.T) {
	cases := []struct {
		name    string
		httpGet func(*testing.T) Getter
		isErr   bool
	}{
		{
			name: "entries were found and are valid",
			httpGet: func(t *testing.T) Getter {
				f, err := os.Open(filepath.ToSlash("testdata/index.yaml"))
				require.NoError(t, err)

				r := &http.Response{
					Body:       f,
					Status:     http.StatusText(http.StatusOK),
					StatusCode: http.StatusOK,
				}

				return &fakeGetter{getResponse: r}
			},
		},
		{
			name: "get failed",
			httpGet: func(t *testing.T) Getter {
				return &fakeGetter{getErr: errors.New("failed")}
			},
			isErr: true,
		},
		{
			name: "server returned invalid status",
			httpGet: func(t *testing.T) Getter {
				r := &http.Response{
					Body:       ioutil.NopCloser(strings.NewReader("")),
					Status:     http.StatusText(http.StatusNotFound),
					StatusCode: http.StatusNotFound,
				}

				return &fakeGetter{getResponse: r}
			},
			isErr: true,
		},
		{
			name: "repository served unexpected file",
			httpGet: func(t *testing.T) Getter {
				r := &http.Response{
					Body:       ioutil.NopCloser(strings.NewReader("<invalid>")),
					Status:     http.StatusText(http.StatusOK),
					StatusCode: http.StatusOK,
				}

				return &fakeGetter{getResponse: r}
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hrc, err := NewHTTPClient("http://example.com", tc.httpGet(t))
			require.NoError(t, err)

			entries, err := hrc.Repository()
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Len(t, entries.Charts, 1)
		})
	}
}

func Test_repositoryClient_Chart(t *testing.T) {

	getChartOK := func(t *testing.T) Getter {
		f, err := os.Open(filepath.ToSlash("testdata/index.yaml"))
		require.NoError(t, err)

		r := &http.Response{
			Body:       f,
			Status:     http.StatusText(http.StatusOK),
			StatusCode: http.StatusOK,
		}

		return &fakeGetter{getResponse: r}
	}

	getChartNotFound := func(t *testing.T) Getter {
		r := &http.Response{
			Body:       ioutil.NopCloser(strings.NewReader("")),
			Status:     http.StatusText(http.StatusNotFound),
			StatusCode: http.StatusNotFound,
		}

		return &fakeGetter{getResponse: r}
	}

	cases := []struct {
		name         string
		getterFn     func(*testing.T) Getter
		chartName    string
		chartVersion string
		expected     *RepositoryChart
		isErr        bool
	}{
		{
			name:         "chart found",
			getterFn:     getChartOK,
			chartName:    "argo-ci",
			chartVersion: "0.1.1",
			expected: &RepositoryChart{
				Description: "A Helm chart for Kubernetes",
				Name:        "argo-ci",
				URLs:        []string{"charts/argo-ci-0.1.1.tgz"},
				Version:     "0.1.1",
			},
		},
		{
			name:      "chart found with no version specified",
			getterFn:  getChartOK,
			chartName: "argo-ci",
			expected: &RepositoryChart{
				Description: "A Helm chart for Kubernetes",
				Name:        "argo-ci",
				URLs:        []string{"charts/argo-ci-0.1.1.tgz"},
				Version:     "0.1.1",
			},
		},
		{
			name:      "chart not found",
			getterFn:  getChartOK,
			chartName: "not-found",
			isErr:     true,
		},
		{
			name:         "chart version not found",
			getterFn:     getChartOK,
			chartName:    "argo-ci",
			chartVersion: "0.1.2",
			isErr:        true,
		},
		{
			name:         "chart not found with version specified",
			getterFn:     getChartOK,
			chartName:    "not-found",
			chartVersion: "0.1.2",
			isErr:        true,
		},
		{
			name:         "unable to query repository",
			getterFn:     getChartNotFound,
			chartName:    "argo-ci",
			chartVersion: "0.1.1",
			isErr:        true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			require.NotNil(t, tc.getterFn, "handler is nil")

			hrc, err := NewHTTPClient("http://example.com", tc.getterFn(t))
			require.NoError(t, err)

			chart, err := hrc.Chart(tc.chartName, tc.chartVersion)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tc.expected, chart)
		})
	}
}

func Test_repositoryClient_Fetch(t *testing.T) {
	g := &fakeGetter{
		getterFn: func(s string) (*http.Response, error) {
			switch s {
			case "http://example.com/foo/bar.tgz":
				r := &http.Response{
					Body:       ioutil.NopCloser(strings.NewReader("contents")),
					StatusCode: http.StatusOK,
					Status:     http.StatusText(http.StatusOK),
				}

				return r, nil
			case "http://example.com/fail":
				return nil, errors.New("fail")
			default:
				r := &http.Response{
					StatusCode: http.StatusNotFound,
					Status:     http.StatusText(http.StatusNotFound),
				}

				return r, nil
			}
		},
	}

	cases := []struct {
		name     string
		uri      string
		expected string
		isErr    bool
	}{
		{
			name:     "successful get with path",
			uri:      "foo/bar.tgz",
			expected: "contents",
		},
		{
			name:     "successful get with url",
			uri:      "http://example.com/foo/bar.tgz",
			expected: "contents",
		},
		{
			name:  "non 200 response",
			uri:   "missing",
			isErr: true,
		},
		{
			name:  "getter failed",
			uri:   "http://example.com/fail",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			hrc, err := NewHTTPClient("http://example.com", g)
			require.NoError(t, err)

			r, err := hrc.Fetch(tc.uri)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			b, err := ioutil.ReadAll(r)
			require.NoError(t, err)

			require.Equal(t, tc.expected, string(b))
		})
	}
}

type fakeGetter struct {
	getterFn    func(string) (*http.Response, error)
	getResponse *http.Response
	getErr      error
}

func (g *fakeGetter) Get(s string) (*http.Response, error) {
	if g.getterFn != nil {
		return g.getterFn(s)
	}

	return g.getResponse, g.getErr
}

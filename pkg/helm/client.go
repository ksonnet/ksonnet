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
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"sort"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/ghodss/yaml"
	"github.com/pkg/errors"
)

var (
	validHelmSchemes = map[string]bool{
		"http":  true,
		"https": true,
	}
)

// RepositoryClient is a client for retrieving Helm repository contents. The structure
// is loosely documented at https://github.com/kubernetes/helm/blob/master/docs/chart_repository.md.
type RepositoryClient interface {
	// Repository returns the contents of a Helm repository.
	Repository() (*Repository, error)
	// Chart returns a Chart with a given name and version. If the version is blank, it returns
	// the latest version.
	Chart(name, version string) (*RepositoryChart, error)
	// Fetch fetches URIs for a chart.
	Fetch(uri string) (io.ReadCloser, error)
}

// Repository is metadata describing the contents of a Helm repository.
type Repository struct {
	Charts map[string][]RepositoryChart `json:"entries,omitempty"`
}

// Latest returns the latest version of charts in a Helm repository.
func (hr *Repository) Latest() []RepositoryChart {
	var out []RepositoryChart

	for _, chart := range hr.Charts {
		if len(chart) > 0 {
			sort.Sort(RepositoryCharts(chart))
			out = append(out, chart[0])
		}
	}

	return out
}

// RepositoryChart is metadata describing a Helm Chart in a repository.
type RepositoryChart struct {
	Description string   `json:"description,omitempty"`
	Name        string   `json:"name,omitempty"`
	URLs        []string `json:"urls,omitempty"`
	Version     string   `json:"version,omitempty"`
}

// RepositoryCharts is a slice of RepositoryChart.
type RepositoryCharts []RepositoryChart

func (rc RepositoryCharts) Len() int {
	return len(rc)
}

func (rc RepositoryCharts) Swap(i, j int) { rc[i], rc[j] = rc[j], rc[i] }

func (rc RepositoryCharts) Less(i, j int) bool {
	v1, err := semver.Make(rc[i].Version)
	if err != nil {
		return false
	}
	v2, err := semver.Make(rc[j].Version)
	if err != nil {
		return false
	}

	return v1.Compare(v2) == 1
}

type Getter interface {
	Get(string) (*http.Response, error)
}

type httpGetter struct {
	client *http.Client
}

func newHTTPGetter() *httpGetter {
	c := &http.Client{
		Timeout: 30 * time.Second,
	}

	return &httpGetter{
		client: c,
	}
}

func (g *httpGetter) Get(s string) (*http.Response, error) {
	req, err := http.NewRequest(http.MethodGet, s, nil)
	if err != nil {
		return nil, err
	}

	return g.client.Do(req)
}

// HTTPClient is a HTTP Helm repository client.
type HTTPClient struct {
	url    *url.URL
	getter Getter
}

var _ RepositoryClient = (*HTTPClient)(nil)

// NewHTTPClient creates an instance of HTTPClient
func NewHTTPClient(urlStr string, hg Getter) (*HTTPClient, error) {
	normalized, err := normalizeHelmURI(urlStr)
	if err != nil {
		return nil, errors.Wrap(err, "normalizing Helm repository URL")
	}

	u, err := url.Parse(normalized)
	if err != nil {
		return nil, err
	}

	if hg == nil {
		hg = newHTTPGetter()
	}

	return &HTTPClient{
		url:    u,
		getter: hg,
	}, nil
}

// Repository returns the Helm repository's content.
func (hrc *HTTPClient) Repository() (*Repository, error) {
	r, err := hrc.Fetch(hrc.url.String())
	if r != nil {
		defer r.Close()
	}

	if err != nil {
		return nil, errors.Wrap(err, "retrieving repository index")
	}

	b, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, errors.Wrap(err, "reading repository index.yaml body")
	}

	var hre Repository

	if err := yaml.Unmarshal(b, &hre); err != nil {
		return nil, errors.Wrap(err, "unmarshalling repository index.yaml")
	}

	return &hre, nil
}

// Chart returns a chart from the repository. If version is blank, it returns the latest.
func (hrc *HTTPClient) Chart(name, version string) (*RepositoryChart, error) {
	repo, err := hrc.Repository()
	if err != nil {
		return nil, errors.Wrap(err, "retrieving repository")
	}

	if version == "" {
		for _, chart := range repo.Latest() {
			if name == chart.Name {
				return &chart, nil
			}
		}

		return nil, errors.Errorf("chart %q was not found", name)
	}

	charts, ok := repo.Charts[name]
	if !ok {
		return nil, errors.Errorf("chart %q was not found", name)
	}

	for _, chart := range charts {
		if version == chart.Version {
			return &chart, nil
		}
	}

	return nil, errors.Errorf("chart %q with version %q was not found", name, version)
}

// Fetch fetches URLs from a repository. If uri is a path, it will use the client URL as the base.
func (hrc *HTTPClient) Fetch(uri string) (io.ReadCloser, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	if u.Host == "" {
		*u = *hrc.url
		u.Path = uri
	}

	resp, err := hrc.getter.Get(u.String())
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.Errorf("unexpected HTTP status code %q when retrieving %q",
			resp.Status, u.String())
	}

	return resp.Body, nil
}

// normalizeHelmURI normalizes a Helm repository URI by returning the
// full URL to repository's index.yaml file.
func normalizeHelmURI(s string) (string, error) {
	u, err := url.Parse(s)
	if err != nil {
		return "", errors.Wrapf(err, "parsing Helm URI %q", s)
	}

	if _, ok := validHelmSchemes[u.Scheme]; !ok {
		return "", errors.Errorf("%q is an invalid scheme for Helm repository", u.Scheme)
	}

	if strings.HasSuffix(u.Path, "index.yaml") {
		return s, nil
	}

	u.Path = path.Join(u.Path, "index.yaml")
	return u.String(), nil
}

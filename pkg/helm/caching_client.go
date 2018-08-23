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

type nameVersion struct {
	name    string
	version string
}

// CachingClient is a caching wrapper over a Helm HTTP client
type CachingClient struct {
	RepositoryClient
	cache map[nameVersion]*RepositoryChart // not thread-safe
}

func NewCachingClient(c RepositoryClient) *CachingClient {
	return &CachingClient{
		RepositoryClient: c,
		cache:            make(map[nameVersion]*RepositoryChart),
	}
}

// Chart returns a Chart with a given name and version. If the version is blank, it returns
// the latest version.
func (c *CachingClient) Chart(name, version string) (*RepositoryChart, error) {
	tuple := nameVersion{name, version}
	if chart, ok := c.cache[tuple]; ok {
		return chart, nil
	}

	chart, err := c.RepositoryClient.Chart(name, version)
	if err != nil {
		return nil, err
	}

	if c.cache == nil {
		c.cache = make(map[nameVersion]*RepositoryChart)
	}
	c.cache[tuple] = chart

	return chart, nil
}

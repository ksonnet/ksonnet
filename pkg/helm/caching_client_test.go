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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeRepositoryClient struct {
	entries    *Repository
	entriesErr error

	chart    *RepositoryChart
	chartErr error

	fetchReader io.ReadCloser
	fetchErr    error
}

func (hrc *fakeRepositoryClient) Repository() (*Repository, error) {
	return hrc.entries, hrc.entriesErr
}

func (hrc *fakeRepositoryClient) Chart(string, string) (*RepositoryChart, error) {
	if hrc.chart == nil {
		return nil, hrc.chartErr
	}
	chartCopy := *hrc.chart
	return &chartCopy, hrc.chartErr
}

func (hrc *fakeRepositoryClient) Fetch(s string) (io.ReadCloser, error) {
	return hrc.fetchReader, hrc.fetchErr
}

func TestCachingClient_Chart(t *testing.T) {
	bareClient := &fakeRepositoryClient{
		chart: &RepositoryChart{
			Name:        "app-a",
			Version:     "0.1.0",
			Description: "description",
			URLs:        []string{"http://example.com/archive"},
		},
	}
	cc := &CachingClient{
		RepositoryClient: bareClient,
	}

	// Show that bare client returns different chart each time
	var name = "app-a"
	var version = "0.1.0"
	var prev *RepositoryChart
	for i := 0; i < 2; i++ {
		chart, err := bareClient.Chart(name, version)
		require.NoError(t, err)
		assert.False(t, prev == chart)
		prev = chart
	}

	// Show that caching client does return same instance when name/version is the same
	var err error
	prev, err = cc.Chart(name, version)
	require.NoError(t, err)
	for i := 0; i < 2; i++ {
		chart, err := cc.Chart(name, version)
		require.NoError(t, err)
		assert.True(t, prev == chart)
	}

	// Show that it changes when name or version change
	name = "app-b"
	chart, err := cc.Chart(name, version)
	require.NoError(t, err)
	assert.False(t, prev == chart)
	prev = chart

	chart, err = cc.Chart(name, version)
	require.NoError(t, err)
	assert.True(t, prev == chart)

	version = "0.2.0"
	chart, err = cc.Chart(name, version)
	require.NoError(t, err)
	assert.False(t, prev == chart)
}

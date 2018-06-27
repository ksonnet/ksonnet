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
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_RegistryClient_ManifestDigest(t *testing.T) {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{
		Timeout:   1 * time.Second,
		Transport: tr,
	}

	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		acceptMIMEType := r.Header.Get("Accept")
		assert.Equal(t, mimeTypeDockerManifest, acceptMIMEType)
		assert.Equal(t, http.MethodHead, r.Method)

		w.Header().Set("Docker-Content-Digest", "sha256:abcde")
	}))

	defer ts.Close()

	c := NewRegistryClient(client, ts.URL)

	digest, err := c.ManifestDigest("foo/bar", "latest")
	require.NoError(t, err)

	require.Equal(t, "sha256:abcde", digest)
}

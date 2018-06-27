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
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// Resolver is able to resolve docker image names into more specific forms
type Resolver interface {
	Resolve(image *ImageName) error
}

type registryResolver struct {
	Client *http.Client
	cache  map[string]string
}

// newRegistryResolver returns a resolver that looks up a docker
// registry to resolve digests
func newRegistryResolver(httpClient *http.Client) Resolver {
	return &registryResolver{
		Client: httpClient,
		cache:  make(map[string]string),
	}
}

func (r *registryResolver) Resolve(n *ImageName) error {
	if n.Digest != "" {
		// Already has explicit digest
		return nil
	}

	if digest, ok := r.cache[n.String()]; ok {
		n.Digest = digest
		return nil
	}

	c := NewRegistryClient(r.Client, n.RegistryURL())
	digest, err := c.ManifestDigest(n.RegistryRepoName(), n.Tag)
	if err != nil {
		switch err.(type) {
		case *imageNotFoundError:
			return err
		default:
			return errors.Errorf("unable to fetch digest for %s: %v", n, err)
		}
	}

	r.cache[n.String()] = digest
	n.Digest = digest
	return nil
}

// ResolverClient is client resolves data from manifests.
type ResolverClient interface {
	ManifestV2Digest(image string) (string, error)
}

// DefaultResolverClient resolves digests for a docker image.
type DefaultResolverClient struct {
	clientFactory func() *http.Client
}

var _ ResolverClient = (*DefaultResolverClient)(nil)

// NewDefaultDigester creates an instance of DefaultDigester.
func NewDefaultDigester() *DefaultResolverClient {
	return &DefaultResolverClient{
		clientFactory: func() *http.Client {
			return &http.Client{
				Transport: NewAuthTransport(http.DefaultTransport),
				Timeout:   15 * time.Second,
			}
		},
	}
}

// ManifestV2Digest returns the 'Docker-Content-Digest' field of the the
// header when making a request to the v2 manifest.
func (d *DefaultResolverClient) ManifestV2Digest(image string) (string, error) {
	n, err := ParseImageName(image)
	if err != nil {
		return "", errors.Wrap(err, "parsing image name")
	}

	client := d.clientFactory()

	resolver := newRegistryResolver(client)
	if err = resolver.Resolve(&n); err != nil {
		return "", errors.Wrap(err, "resolving image")
	}

	return n.String(), nil
}

// ResolveImage loads the digest reference for a docker image.
func ResolveImage(image string) (string, error) {
	d := NewDefaultDigester()
	return d.ManifestV2Digest(image)
}

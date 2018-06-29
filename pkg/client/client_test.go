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
	"testing"

	swagger "github.com/emicklei/go-restful-swagger12"
	"github.com/googleapis/gnostic/OpenAPIv2"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
)

func TestConfig_GetAPISpec(t *testing.T) {
	cases := []struct {
		name      string
		version   string
		disc      discovery.DiscoveryInterface
		createErr error
	}{
		{
			name:    "in general",
			version: "version:v1.9.3",
			disc:    &fakeDiscovery{withVersion: "v1.9.3"},
		},
		{
			name:    "with dev version",
			version: "version:v1.9.3",
			disc:    &fakeDiscovery{withVersion: "v1.9.3-fadecafe"},
		},
		{
			name:    "with dev version",
			version: "version:v1.9.3",
			disc:    &fakeDiscovery{withVersion: "v1.9.3+facade"},
		},
		{
			name:      "unable to create discovery client",
			version:   "version:v1.8.0",
			createErr: errors.New("failed"),
		},
		{
			name:    "retrieve open api schema error",
			version: "version:v1.8.0",
			disc:    &fakeDiscovery{withError: true},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := Config{
				Config: &clientConfig{},
				discoveryClient: func() (discovery.DiscoveryInterface, error) {
					return tc.disc, tc.createErr
				},
			}
			got := c.GetAPISpec()
			require.Equal(t, tc.version, got)
		})
	}

}

type clientConfig struct {
}

var _ clientcmd.ClientConfig = (*clientConfig)(nil)

func (c *clientConfig) RawConfig() (clientcmdapi.Config, error) {
	return clientcmdapi.Config{}, errors.Errorf("not implemented")
}

func (c *clientConfig) ClientConfig() (*restclient.Config, error) {
	rc := &restclient.Config{}

	return rc, nil
}

func (c *clientConfig) Namespace() (string, bool, error) {
	return "", false, errors.Errorf("not implemented")
}

func (c *clientConfig) ConfigAccess() clientcmd.ConfigAccess {
	var ca clientcmd.ConfigAccess

	return ca
}

type fakeDiscovery struct {
	withError   bool
	withVersion string
}

var _ discovery.DiscoveryInterface = (*fakeDiscovery)(nil)

func (c *fakeDiscovery) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeDiscovery) ServerResources() ([]*metav1.APIResourceList, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeDiscovery) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return nil, nil
}

func (c *fakeDiscovery) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	return nil, nil
}

func (c *fakeDiscovery) ServerGroups() (*metav1.APIGroupList, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeDiscovery) ServerVersion() (*version.Info, error) {
	if c.withError {
		return nil, errors.New("server version error")
	}

	return &version.Info{
		GitVersion: c.withVersion,
	}, nil
}

func (c *fakeDiscovery) OpenAPISchema() (*openapi_v2.Document, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeDiscovery) SwaggerSchema(version schema.GroupVersion) (*swagger.ApiDeclaration, error) {
	return nil, errors.New("not implemented")
}

func (c *fakeDiscovery) RESTClient() restclient.Interface {
	return nil
}

// Copyright 2017 The kubecfg authors
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

package utils

import (
	"sync"

	"github.com/emicklei/go-restful/swagger"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/version"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

type memcachedDiscoveryClient struct {
	cl              discovery.DiscoveryInterface
	lock            sync.RWMutex
	servergroups    *metav1.APIGroupList
	serverresources map[string]*metav1.APIResourceList
}

// NewMemcachedDiscoveryClient creates a new DiscoveryClient that
// caches results in memory
func NewMemcachedDiscoveryClient(cl discovery.DiscoveryInterface) discovery.CachedDiscoveryInterface {
	c := &memcachedDiscoveryClient{cl: cl}
	c.Invalidate()
	return c
}

func (c *memcachedDiscoveryClient) Fresh() bool {
	return true
}

func (c *memcachedDiscoveryClient) Invalidate() {
	c.lock.Lock()
	defer c.lock.Unlock()

	c.servergroups = nil
	c.serverresources = make(map[string]*metav1.APIResourceList)
}

func (c *memcachedDiscoveryClient) RESTClient() rest.Interface {
	return c.cl.RESTClient()
}

func (c *memcachedDiscoveryClient) ServerGroups() (*metav1.APIGroupList, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	var err error
	if c.servergroups != nil {
		return c.servergroups, nil
	}
	c.servergroups, err = c.cl.ServerGroups()
	return c.servergroups, err
}

func (c *memcachedDiscoveryClient) ServerResourcesForGroupVersion(groupVersion string) (*metav1.APIResourceList, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	var err error
	if v := c.serverresources[groupVersion]; v != nil {
		return v, nil
	}
	c.serverresources[groupVersion], err = c.cl.ServerResourcesForGroupVersion(groupVersion)
	return c.serverresources[groupVersion], err
}

func (c *memcachedDiscoveryClient) ServerResources() ([]*metav1.APIResourceList, error) {
	return c.cl.ServerResources()
}

func (c *memcachedDiscoveryClient) ServerPreferredResources() ([]*metav1.APIResourceList, error) {
	return c.cl.ServerPreferredResources()
}

func (c *memcachedDiscoveryClient) ServerPreferredNamespacedResources() ([]*metav1.APIResourceList, error) {
	return c.cl.ServerPreferredNamespacedResources()
}

func (c *memcachedDiscoveryClient) ServerVersion() (*version.Info, error) {
	return c.cl.ServerVersion()
}

func (c *memcachedDiscoveryClient) SwaggerSchema(version schema.GroupVersion) (*swagger.ApiDeclaration, error) {
	return c.cl.SwaggerSchema(version)
}

var _ discovery.CachedDiscoveryInterface = &memcachedDiscoveryClient{}

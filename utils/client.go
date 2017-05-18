package utils

import (
	"sync"

	"github.com/emicklei/go-restful/swagger"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/pkg/api/unversioned"
	"k8s.io/client-go/pkg/version"
	"k8s.io/client-go/rest"
)

type memcachedDiscoveryClient struct {
	cl                        discovery.DiscoveryInterface
	lock                      sync.RWMutex
	servergroups              *unversioned.APIGroupList
	serverresources           map[string]*unversioned.APIResourceList
	serverresourcesIsComplete bool
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
	c.serverresources = make(map[string]*unversioned.APIResourceList)
	c.serverresourcesIsComplete = false
}

func (c *memcachedDiscoveryClient) RESTClient() rest.Interface {
	return c.cl.RESTClient()
}

func (c *memcachedDiscoveryClient) ServerGroups() (*unversioned.APIGroupList, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	var err error
	if c.servergroups != nil {
		return c.servergroups, nil
	}
	c.servergroups, err = c.cl.ServerGroups()
	return c.servergroups, err
}

func (c *memcachedDiscoveryClient) ServerResourcesForGroupVersion(groupVersion string) (*unversioned.APIResourceList, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	var err error
	if v := c.serverresources[groupVersion]; v != nil {
		return v, nil
	}
	if c.serverresourcesIsComplete {
		return &unversioned.APIResourceList{}, nil
	}
	c.serverresources[groupVersion], err = c.cl.ServerResourcesForGroupVersion(groupVersion)
	return c.serverresources[groupVersion], err
}

func (c *memcachedDiscoveryClient) ServerResources() (map[string]*unversioned.APIResourceList, error) {
	c.lock.Lock()
	defer c.lock.Unlock()

	var err error
	if c.serverresourcesIsComplete {
		return c.serverresources, nil
	}
	c.serverresources, err = c.cl.ServerResources()
	if err == nil {
		c.serverresourcesIsComplete = true
	}
	return c.serverresources, err
}

func (c *memcachedDiscoveryClient) ServerPreferredResources() ([]unversioned.GroupVersionResource, error) {
	return c.cl.ServerPreferredResources()
}

func (c *memcachedDiscoveryClient) ServerPreferredNamespacedResources() ([]unversioned.GroupVersionResource, error) {
	return c.cl.ServerPreferredNamespacedResources()
}

func (c *memcachedDiscoveryClient) ServerVersion() (*version.Info, error) {
	return c.cl.ServerVersion()
}

func (c *memcachedDiscoveryClient) SwaggerSchema(version unversioned.GroupVersion) (*swagger.ApiDeclaration, error) {
	return c.cl.SwaggerSchema(version)
}

var _ discovery.CachedDiscoveryInterface = &memcachedDiscoveryClient{}

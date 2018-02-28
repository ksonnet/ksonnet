// Copyright 2017 The ksonnet authors
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
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path"
	"reflect"
	"time"

	"github.com/ksonnet/ksonnet/metadata"
	str "github.com/ksonnet/ksonnet/strings"
	"github.com/ksonnet/ksonnet/utils"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/tools/clientcmd"
)

// Config is a wrapper around client-go's ClientConfig
type Config struct {
	Overrides    *clientcmd.ConfigOverrides
	LoadingRules *clientcmd.ClientConfigLoadingRules

	Config clientcmd.ClientConfig
}

// NewClientConfig initializes a new client.Config with the provided loading rules and overrides.
func NewClientConfig(overrides clientcmd.ConfigOverrides, loadingRules clientcmd.ClientConfigLoadingRules) *Config {
	config := clientcmd.NewInteractiveDeferredLoadingClientConfig(&loadingRules, &overrides, os.Stdin)
	return &Config{Overrides: &overrides, LoadingRules: &loadingRules, Config: config}
}

// NewDefaultClientConfig initializes a new ClientConfig with default loading rules and no overrides.
func NewDefaultClientConfig() *Config {
	overrides := clientcmd.ConfigOverrides{}
	loadingRules := *clientcmd.NewDefaultClientConfigLoadingRules()
	loadingRules.DefaultClientConfig = &clientcmd.DefaultClientConfig
	config := clientcmd.NewInteractiveDeferredLoadingClientConfig(&loadingRules, &overrides, os.Stdin)

	return &Config{Overrides: &overrides, LoadingRules: &loadingRules, Config: config}
}

// InitClient initializes a new ClientConfig given the specified environment
// spec and returns the ClientPool, DiscoveryInterface, and namespace.
func InitClient(env string) (dynamic.ClientPool, discovery.DiscoveryInterface, string, error) {
	clientConfig := NewDefaultClientConfig()
	return clientConfig.RestClient(&env)
}

// GetAPISpec reads the kubernetes API version from this client's swagger.json.
// We anticipate the swagger.json to be located at <server>/swagger.json.
// If no swagger is found, or we are unable to authenticate to the server, we
// will default to version:v1.7.0.
func (c *Config) GetAPISpec(server string) string {
	const (
		defaultVersion = "version:v1.7.0"
	)

	type Info struct {
		Version string `json:"version"`
	}

	type Spec struct {
		Info Info `json:"info"`
	}

	u, err := url.Parse(server)
	u.Path = path.Join(u.Path, "swagger.json")
	url := u.String()

	client := http.Client{
		Timeout: time.Second * 2,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log.Debugf("Failed to create request at %s\n%s", url, err.Error())
		return defaultVersion
	}

	res, err := client.Do(req)
	if err != nil {
		log.Debugf("Failed to open swagger at %s\n%s", url, err.Error())
		return defaultVersion
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Debugf("Failed to read swagger at %s\n%s", url, err.Error())
		return defaultVersion
	}

	spec := Spec{}
	jsonErr := json.Unmarshal(body, &spec)
	if jsonErr != nil {
		log.Debugf("Failed to parse swagger at %s\n%s", url, err.Error())
		return defaultVersion
	}

	return fmt.Sprintf("version:%s", spec.Info.Version)
}

// Namespace returns the namespace for the provided ClientConfig.
func (c *Config) Namespace() (string, error) {
	ns, _, err := c.Config.Namespace()
	return ns, err
}

// RestClient returns the ClientPool, DiscoveryInterface, and Namespace based on the environment spec.
func (c *Config) RestClient(envName *string) (dynamic.ClientPool, discovery.DiscoveryInterface, string, error) {
	if envName != nil {
		err := c.overrideCluster(*envName)
		if err != nil {
			return nil, nil, "", err
		}
	}

	conf, err := c.Config.ClientConfig()
	if err != nil {
		return nil, nil, "", err
	}

	disco, err := discovery.NewDiscoveryClientForConfig(conf)
	if err != nil {
		return nil, nil, "", err
	}

	discoCache := utils.NewMemcachedDiscoveryClient(disco)
	mapper := discovery.NewDeferredDiscoveryRESTMapper(discoCache, dynamic.VersionInterfaces)
	pathresolver := dynamic.LegacyAPIPathResolverFunc

	pool := dynamic.NewClientPool(conf, mapper, pathresolver)

	ns, err := c.Namespace()
	if err != nil {
		return nil, nil, "", err
	}

	return pool, discoCache, ns, nil
}

// BindClientGoFlags binds client-go flags to the specified command. This way
// any overrides to client-go flags will automatically update the client config.
func (c *Config) BindClientGoFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringVar(&c.LoadingRules.ExplicitPath, "kubeconfig", "", "Path to a kubeconfig file. Alternative to env var $KUBECONFIG.")
	clientcmd.BindOverrideFlags(c.Overrides, cmd.PersistentFlags(), clientcmd.RecommendedConfigOverrideFlags(""))
}

// ResolveContext returns the server and namespace of the cluster at the
// provided context. If the context string is empty, the "default" context is
// used.
func (c *Config) ResolveContext(context string) (server, namespace string, err error) {
	rawConfig, err := c.Config.RawConfig()
	if err != nil {
		return "", "", err
	}

	// use the default context where context is empty
	if context == "" {
		if rawConfig.CurrentContext == "" && len(rawConfig.Clusters) == 0 {
			// User likely does not have a kubeconfig file.
			return "", "", fmt.Errorf("No current context found. Make sure a kubeconfig file is present")
		}
		// Note: "" is a valid rawConfig.CurrentContext
		context = rawConfig.CurrentContext
	}

	ctx := rawConfig.Contexts[context]
	if ctx == nil {
		return "", "", fmt.Errorf("context '%s' does not exist in the kubeconfig file", context)
	}

	log.Infof("Using context '%s' from the kubeconfig file specified at the environment variable $KUBECONFIG", context)
	cluster, exists := rawConfig.Clusters[ctx.Cluster]
	if !exists {
		return "", "", fmt.Errorf("No cluster with name '%s' exists", ctx.Cluster)
	}

	return cluster.Server, ctx.Namespace, nil
}

// overrideCluster ensures that the server specified in the environment is
// associated in the user's kubeconfig file during deployment to a ksonnet
// environment. We will error out if it is not.
//
// If the environment server the user is attempting to deploy to is not the current
// kubeconfig context, we must manually override the client-go --cluster flag
// to ensure we are deploying to the correct cluster.
func (c *Config) overrideCluster(envName string) error {
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}

	metadataManager, err := metadata.Find(cwd)
	if err != nil {
		return err
	}

	rawConfig, err := c.Config.RawConfig()
	if err != nil {
		return err
	}

	var servers = make(map[string]string)
	for name, cluster := range rawConfig.Clusters {
		server, err := str.NormalizeURL(cluster.Server)
		if err != nil {
			return err
		}

		servers[server] = name
	}

	//
	// check to ensure that the environment we are trying to deploy to is
	// created, and that the server is located in kubeconfig.
	//

	log.Debugf("Validating deployment at '%s' with server '%v'", envName, reflect.ValueOf(servers).MapKeys())
	env, err := metadataManager.GetEnvironment(envName)
	if err != nil {
		return err
	}

	server, err := str.NormalizeURL(env.Destination.Server)
	if err != nil {
		return err
	}

	if _, ok := servers[server]; ok {
		clusterName := servers[server]
		if c.Overrides.Context.Cluster == "" {
			log.Debugf("Overwriting --cluster flag with '%s'", clusterName)
			c.Overrides.Context.Cluster = clusterName
		}
		if c.Overrides.Context.Namespace == "" {
			log.Debugf("Overwriting --namespace flag with '%s'", env.Destination.Namespace)
			c.Overrides.Context.Namespace = env.Destination.Namespace
		}
		return nil
	}

	return fmt.Errorf("Attempting to deploy to environment '%s' at '%s', but cannot locate a server at that address", envName, env.Destination.Server)
}

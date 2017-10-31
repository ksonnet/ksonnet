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

package metadata

import (
	"os"
	"regexp"
	"strings"

	"github.com/ksonnet/ksonnet/metadata/app"
	param "github.com/ksonnet/ksonnet/metadata/params"
	"github.com/ksonnet/ksonnet/metadata/registry"
	"github.com/ksonnet/ksonnet/prototype"
	"github.com/spf13/afero"
)

var appFS afero.Fs
var defaultFolderPermissions = os.FileMode(0755)
var defaultFilePermissions = os.FileMode(0644)

// AbsPath is an advisory type that represents an absolute path. It is advisory
// in that it is not forced to be absolute, but rather, meant to indicate
// intent, and make code easier to read.
type AbsPath string

// AbsPaths is a slice of `AbsPath`.
type AbsPaths []string

// Manager abstracts over a ksonnet application's metadata, allowing users to do
// things like: create and delete environments; search for prototypes; vendor
// libraries; and other non-core-application tasks.
type Manager interface {
	Root() AbsPath
	LibPaths(envName string) (libPath, envLibPath, envComponentPath, envParamsPath AbsPath)

	// Components API.
	ComponentPaths() (AbsPaths, error)
	CreateComponent(name string, text string, params param.Params, templateType prototype.TemplateType) error

	// Params API.
	SetComponentParams(component string, params param.Params) error
	GetComponentParams(name string) (param.Params, error)
	GetAllComponentParams() (map[string]param.Params, error)
	// GetEnvironmentParams will take the name of an environment and return a
	// mapping of parameters of the form:
	// componentName => {param key => param val}
	// i.e.: "nginx" => {"replicas" => 1, "name": "nginx"}
	GetEnvironmentParams(name string) (map[string]param.Params, error)
	SetEnvironmentParams(env, component string, params param.Params) error

	// Environment API.
	CreateEnvironment(name, uri, namespace string, spec ClusterSpec) error
	DeleteEnvironment(name string) error
	GetEnvironments() ([]*Environment, error)
	GetEnvironment(name string) (*Environment, error)
	SetEnvironment(name string, desired *Environment) error

	// Spec API.
	AppSpec() (*app.Spec, error)

	// Registry API.
	AddRegistry(name, protocol, uri string) (*registry.Spec, error)
}

// Find will recursively search the current directory and its parents for a
// `.ksonnet` folder, which marks the application root. Returns error if there
// is no application root.
func Find(path AbsPath) (Manager, error) {
	return findManager(path, afero.NewOsFs())
}

// Init will retrieve a cluster API specification, generate a
// capabilities-compliant version of ksonnet-lib, and then generate the
// directory tree for an application.
func Init(name string, rootPath AbsPath, spec ClusterSpec, serverURI, namespace *string) (Manager, error) {
	return initManager(name, rootPath, spec, serverURI, namespace, appFS)
}

// ClusterSpec represents the API supported by some cluster. There are several
// ways to specify a cluster, including: querying the API server, reading an
// OpenAPI spec in some file, or consulting the OpenAPI spec released in a
// specific version of Kubernetes.
type ClusterSpec interface {
	data() ([]byte, error)
	resource() string // For testing parsing logic.
}

// ParseClusterSpec will parse a cluster spec flag and output a well-formed
// ClusterSpec object. For example, if the flag is `--version:v1.7.1`, then we
// will output a ClusterSpec representing the cluster specification associated
// with the `v1.7.1` build of Kubernetes.
func ParseClusterSpec(specFlag string) (ClusterSpec, error) {
	return parseClusterSpec(specFlag, appFS)
}

// isValidName returns true if a name (e.g., for an environment) is valid.
// Broadly, this means it does not contain punctuation, whitespace, leading or
// trailing slashes.
func isValidName(name string) bool {
	// No unicode whitespace is allowed. `Fields` doesn't handle trailing or
	// leading whitespace.
	fields := strings.Fields(name)
	if len(fields) > 1 || len(strings.TrimSpace(name)) != len(name) {
		return false
	}

	hasPunctuation := regexp.MustCompile(`[\\,;':!()?"{}\[\]*&%@$]+`).MatchString
	hasTrailingSlashes := regexp.MustCompile(`/+$`).MatchString
	hasLeadingSlashes := regexp.MustCompile(`^/+`).MatchString
	return len(name) != 0 && !hasPunctuation(name) && !hasTrailingSlashes(name) && !hasLeadingSlashes(name)
}

func init() {
	appFS = afero.NewOsFs()
}

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
	"github.com/ksonnet/ksonnet/metadata/parts"
	"github.com/ksonnet/ksonnet/metadata/registry"
	"github.com/ksonnet/ksonnet/prototype"
	"github.com/spf13/afero"
)

var appFS afero.Fs
var defaultFolderPermissions = os.FileMode(0755)
var defaultFilePermissions = os.FileMode(0644)

// Manager abstracts over a ksonnet application's metadata, allowing users to do
// things like: create and delete environments; search for prototypes; vendor
// libraries; and other non-core-application tasks.
type Manager interface {
	Root() string
	LibPaths() (envPath, vendorPath string)
	EnvPaths(env string) (libPath, mainPath, paramsPath string, err error)

	// Components API.
	ComponentPaths() ([]string, error)
	GetAllComponents() ([]string, error)
	CreateComponent(name string, text string, params param.Params, templateType prototype.TemplateType) error
	DeleteComponent(name string) error

	// Params API.
	SetComponentParams(component string, params param.Params) error
	GetComponentParams(name string) (param.Params, error)
	GetAllComponentParams(cwd string) (map[string]param.Params, error)
	// GetEnvironmentParams will take the name of an environment and return a
	// mapping of parameters of the form:
	// componentName => {param key => param val}
	// i.e.: "nginx" => {"replicas" => 1, "name": "nginx"}
	GetEnvironmentParams(name string) (map[string]param.Params, error)
	SetEnvironmentParams(env, component string, params param.Params) error

	// Environment API.
	CreateEnvironment(name, uri, namespace, spec string) error
	DeleteEnvironment(name string) error
	GetEnvironments() (app.EnvironmentSpecs, error)
	GetEnvironment(name string) (*app.EnvironmentSpec, error)
	SetEnvironment(name, desiredName string) error
	// ErrorOnSpecFile is a temporary API to inform < 0.9.0 ks users of environment directory changes.
	ErrorOnSpecFile() error

	// Spec API.
	AppSpec() (*app.Spec, error)
	WriteAppSpec(*app.Spec) error

	// Dependency/registry API.
	AddRegistry(name, protocol, uri, version string) (*registry.Spec, error)
	GetRegistry(name string) (*registry.Spec, string, error)
	GetPackage(registryName, libID string) (*parts.Spec, error)
	CacheDependency(registryName, libID, libName, libVersion string) (*parts.Spec, error)
	GetDependency(libName string) (*parts.Spec, error)
	GetAllPrototypes() (prototype.SpecificationSchemas, error)
}

// Find will recursively search the current directory and its parents for a
// `.ksonnet` folder, which marks the application root. Returns error if there
// is no application root.
func Find(path string) (Manager, error) {
	return findManager(path, afero.NewOsFs())
}

// Init will generate the directory tree for a ksonnet project.
func Init(name, rootPath string, k8sSpecFlag, serverURI, namespace *string) (Manager, error) {
	// Generate `incubator` registry. We do this before before creating
	// directory tree, in case the network call fails.
	const (
		defaultIncubatorRegName = "incubator"
		defaultIncubatorURI     = "github.com/ksonnet/parts/tree/master/" + defaultIncubatorRegName
	)

	gh, err := makeGitHubRegistryManager(&app.RegistryRefSpec{
		Name:     "incubator",
		Protocol: "github",
		URI:      defaultIncubatorURI,
	})
	if err != nil {
		return nil, err
	}

	return initManager(name, rootPath, k8sSpecFlag, serverURI, namespace, gh, appFS)
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

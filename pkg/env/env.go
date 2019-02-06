// Copyright 2018 The kubecfg authors
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

package env

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/ksonnet/ksonnet/pkg/helm"
	utilio "github.com/ksonnet/ksonnet/pkg/util/io"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/params"
	"github.com/ksonnet/ksonnet/pkg/registry"
	"github.com/ksonnet/ksonnet/pkg/util/jsonnet"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const (
	// primary environment files.
	envFileName     = "main.jsonnet"
	paramsFileName  = "params.libsonnet"
	globalsFileName = "globals.libsonnet"

	// envRootName is the name for the environment root.
	envRootName = "environments"
)

var (
	componentJPaths  = make([]string, 0)
	componentExtVars = make(map[string]string)
	componentTlaVars = make(map[string]string)
)

// AddJPaths adds paths to JPath for a component evaluation.
func AddJPaths(paths ...string) {
	componentJPaths = append(componentJPaths, paths...)
}

// AddExtVar adds an ext var to a component evaluation.
func AddExtVar(key, value string) {
	componentExtVars[key] = value
}

// AddExtVarFile adds an ext var from a file to component evaluation.
func AddExtVarFile(fs afero.Fs, key, filePath string) error {
	data, err := afero.ReadFile(fs, filePath)
	if err != nil {
		return err
	}

	componentExtVars[key] = string(data)
	return nil
}

// AddTlaVar adds a tla var to a component evaluation.
func AddTlaVar(key, value string) {
	componentTlaVars[key] = value
}

// AddTlaVarFile adds a tla var from a file to component evaluation.
func AddTlaVarFile(fs afero.Fs, key, filePath string) error {
	data, err := afero.ReadFile(fs, filePath)
	if err != nil {
		return err
	}

	componentTlaVars[key] = string(data)
	return nil
}

// MainFile returns the contents of the environment's main source.
func MainFile(a app.App, envName string) (string, error) {
	path, err := Path(a, envName, envFileName)
	if err != nil {
		return "", err
	}

	snippet, err := afero.ReadFile(a.Fs(), path)
	if err != nil {
		return "", err
	}

	return string(snippet), nil
}

// Evaluate evaluates an environment.
func Evaluate(a app.App, envName, components, paramsStr string, opts ...jsonnet.VMOpt) (string, error) {

	snippet, err := MainFile(a, envName)
	if err != nil {
		return "", err
	}

	evaluated, err := evaluateMain(a, envName, snippet, components, paramsStr, opts...)
	if err != nil {
		return "", err
	}

	return upgradeArray(evaluated)
}

func evaluateMain(a app.App, envName, snippet, components, paramsStr string, opts ...jsonnet.VMOpt) (string, error) {
	libPath, err := a.LibPath(envName)
	if err != nil {
		return "", err
	}

	appEnv, err := a.Environment(envName)
	if err != nil {
		return "", err
	}

	vm := jsonnet.NewVM(opts...)

	vm.AddJPath(componentJPaths...)
	vm.AddJPath(
		filepath.Join(a.Root(), envRootName),
		filepath.Join(a.Root(), envRootName, appEnv.Path),
		filepath.Join(a.Root(), "vendor"),
		filepath.Join(a.Root(), "lib"),
		libPath,
	)

	helmRenderer := helm.NewRenderer(a, envName)
	vm.AddFunctions(helmRenderer.JsonnetNativeFunc()...)

	// Re-vendor versioned packages, such that import paths will remain path-agnostic.
	// TODO Where should packagemanager come from?
	pm := registry.NewPackageManager(a)
	revendoredPath, cleanup, err := revendorPackages(a, pm, appEnv)
	if err != nil {
		return "", errors.Wrapf(err, "revendoring packages for environment: %v", envName)
	}
	defer cleanup()
	vm.AddJPath(revendoredPath) // TODO does precedence matter?
	// end re-vendor

	if len(appEnv.Targets) == 0 {
		vm.AddJPath(filepath.Join(a.Root(), "components"))
	} else {
		for _, moduleName := range appEnv.Targets {
			path := filepath.Join(append([]string{a.Root(), "components"}, moduleName)...)
			vm.AddJPath(path)
		}
	}

	envCode, err := params.JsonnetEnvObject(a, envName)
	if err != nil {
		return "", err
	}

	for k, v := range componentExtVars {
		vm.ExtVar(k, v)
	}

	for k, v := range componentTlaVars {
		vm.TLAVar(k, v)
	}

	vm.ExtCode("__ksonnet/environments", envCode)
	vm.ExtCode(ComponentsExtCodeKey, components)
	vm.ExtCode("__ksonnet/params", paramsStr)

	return vm.EvaluateSnippet(envFileName, snippet)
}

// upgradeArray wraps component lists in Kubernetes lists.
func upgradeArray(snippet string) (string, error) {
	vm := jsonnet.NewVM()

	vm.ExtCode("__src", snippet)
	return vm.EvaluateSnippet("upgradeArray", jsonnetUpgradeArray)
}

var jsonnetUpgradeArray = `
local __src = std.extVar("__src");
local components = std.objectFields(__src);

{
[x]:
  if std.type(__src[x]) == "array"
  then {apiVersion: "v1", kind: "List", items: __src[x]}
  else __src[x]
for x in components
}
`

func envRoot(a app.App, envName string) (string, error) {
	envSpec, err := a.Environment(envName)
	if err != nil {
		return "", err
	}

	return filepath.Join(a.Root(), envRootName, envSpec.Path), nil

}

// Path constructs a path to a file or directory in an environment.
func Path(a app.App, envName string, path ...string) (string, error) {
	base, err := envRoot(a, envName)
	if err != nil {
		return "", err
	}

	return filepath.Join(append([]string{base}, path...)...), nil
}

func environmentsCode(a app.App, envName string) (string, error) {
	envDetails, err := a.Environment(envName)
	if err != nil {
		return "", err
	}

	dest := map[string]string{
		"server":    envDetails.Destination.Server,
		"namespace": envDetails.Destination.Namespace,
	}

	marshalledDestination, err := json.Marshal(&dest)
	if err != nil {
		return "", err
	}

	return string(marshalledDestination), nil
}

// buildPackagePaths builds a set of version-specific package paths that
// should be made available when applying an environment.
// NOTE: we currently exclude unversioned packages, they can be picked
//       up in the legacy location under the vendor directory.
// Return map keys are qualified package names (<registry>/<package>).
func buildPackagePaths(pm registry.PackageManager, e *app.EnvironmentConfig) (map[string]string, error) {
	log := log.WithField("action", "env.buildPackagePaths")

	if pm == nil {
		return nil, errors.Errorf("nil package manager")
	}
	if e == nil {
		return nil, errors.Errorf("nil environment")
	}

	result := make(map[string]string)

	pkgList, err := pm.PackagesForEnv(e)
	if err != nil {
		return nil, err
	}

	for _, v := range pkgList {
		if v.Version() == "" {
			log.Debugf("skipping unversioned packaged: %s/%s", v.RegistryName(), v.Name())
			continue
		}
		k := fmt.Sprintf("%s/%s", v.RegistryName(), v.Name())
		result[k] = v.Path()
	}
	return result, nil
}

// Builds a vendor import path with the correct versions of referenced packages for
// the specified environment.
// The caller is responsible for calling the returned cleanup function to release
// and temporary resources.
func revendorPackages(a app.App, pm registry.PackageManager, e *app.EnvironmentConfig) (path string, cleanup func() error, err error) {
	log := log.WithField("action", "env.revendorPackages")

	noop := func() error { return nil }

	if a == nil {
		return "", noop, errors.Errorf("nil app")
	}
	if pm == nil {
		return "", nil, errors.Errorf("nil package manager")
	}
	if e == nil {
		return "", noop, errors.Errorf("nil environment")
	}
	fs := a.Fs()
	if fs == nil {
		return "", noop, errors.Errorf("nil filesystem interface")
	}

	// Enumerate packages
	pathByPkg, err := buildPackagePaths(pm, e)
	if err != nil {
		return "", noop, err
	}

	// Build our temporary space
	tmpDir, err := afero.TempDir(fs, "", "ksvendor")
	if err != nil {
		return "", noop, errors.Wrap(err, "creating temporary vendor path")
	}
	shouldCleanup := true // Used to decide whether we should cleanup in our defer or handoff responsibility to our callers
	internalCleanFunc := func() error {
		if !shouldCleanup {
			return nil
		}

		return fs.RemoveAll(tmpDir)
	}
	defer internalCleanFunc()

	// Copy each package to our temp directory destined for import,
	// removing version information from the path.
	// This allows our consumers to import the package with a version-agnostic import specifier.
	for k, srcPath := range pathByPkg {
		if srcPath == "" {
			log.Warnf("skipping package %v", k)
			continue
		}
		// Check for missing package.
		// There are two common causes:
		// 1. It is installed but in an unversioned path - the app hasn't been upgraded.
		// 2. It is actually missing - the vendor cache needs to be refreshed.
		// Currently we assume #1 and skip revendoring - it can be imported from the legacy path.
		ok, err := afero.Exists(fs, srcPath)
		if err != nil {
			return "", noop, err
		}
		if !ok {
			// TODO differentiate between above cases #1 and #2.
			log.Warnf("skipping missing path %v. Please run `ks upgrade`.", srcPath)
			continue
		}

		dstPath := filepath.Join(tmpDir, filepath.FromSlash(k))
		log.Debugf("preparing package %v->%v", srcPath, dstPath)
		if err := utilio.CopyRecursive(fs, dstPath, srcPath, app.DefaultFilePermissions, app.DefaultFolderPermissions); err != nil {
			return "", noop, errors.Wrapf(err, "copying package %v->%v", srcPath, dstPath)
		}
	}

	// Signal to our deferred cleanup function that our caller is now
	// the responsible party for cleaning up the temp directory.
	shouldCleanup = false
	callerCleanFunc := func() error {
		return fs.RemoveAll(tmpDir)
	}
	return tmpDir, callerCleanFunc, nil
}

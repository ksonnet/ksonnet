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

package component

import (
	"path"
	"path/filepath"
	"strings"

	"github.com/davecgh/go-spew/spew"
	"github.com/ksonnet/ksonnet/metadata/app"
	param "github.com/ksonnet/ksonnet/metadata/params"
	"github.com/ksonnet/ksonnet/prototype"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

var (
	// DefaultManager is the default manager for components.
	DefaultManager = &defaultManager{}
)

// Manager is an interface for interating with components.
type Manager interface {
	Components(ns Namespace) ([]Component, error)
	Component(ksApp app.App, nsName, componentName string) (Component, error)
	CreateComponent(ksApp app.App, name, text string, params param.Params, templateType prototype.TemplateType) (string, error)
	CreateNamespace(ksApp app.App, name string) error
	Namespace(ksApp app.App, nsName string) (Namespace, error)
	Namespaces(ksApp app.App, envName string) ([]Namespace, error)
	NSResolveParams(ns Namespace) (string, error)
	ResolvePath(ksApp app.App, path string) (Namespace, Component, error)
}

type defaultManager struct{}

var _ Manager = (*defaultManager)(nil)

func (dm *defaultManager) Namespaces(ksApp app.App, envName string) ([]Namespace, error) {
	return NamespacesFromEnv(ksApp, envName)
}

func (dm *defaultManager) Namespace(ksApp app.App, nsName string) (Namespace, error) {
	return GetNamespace(ksApp, nsName)
}

func (dm *defaultManager) NSResolveParams(ns Namespace) (string, error) {
	return ns.ResolvedParams()
}

func (dm *defaultManager) Components(ns Namespace) ([]Component, error) {
	return ns.Components()
}

func (dm *defaultManager) Component(ksApp app.App, nsName, componentName string) (Component, error) {
	return LocateComponent(ksApp, nsName, componentName)
}

func (dm *defaultManager) CreateComponent(ksApp app.App, name, text string, params param.Params, templateType prototype.TemplateType) (string, error) {
	return Create(ksApp, name, text, params, templateType)
}

func (dm *defaultManager) CreateNamespace(ksApp app.App, name string) error {
	parts := strings.Split(name, "/")
	dir := filepath.Join(append([]string{ksApp.Root(), "components"}, parts...)...)

	if err := ksApp.Fs().MkdirAll(dir, app.DefaultFolderPermissions); err != nil {
		return err
	}

	paramsDir := filepath.Join(dir, "params.libsonnet")
	return afero.WriteFile(ksApp.Fs(), paramsDir, GenParamsContent(), app.DefaultFilePermissions)
}

func (dm *defaultManager) ResolvePath(ksApp app.App, path string) (Namespace, Component, error) {
	isDir, err := dm.isComponentDir(ksApp, path)
	if err != nil {
		return nil, nil, errors.Wrap(err, "check for namespace directory")
	}

	if isDir {
		ns, err := dm.Namespace(ksApp, path)
		if err != nil {
			return nil, nil, err
		}

		return ns, nil, nil
	}

	nsName, cName, err := dm.checkComponent(ksApp, path)
	if err != nil {
		return nil, nil, err
	}

	spew.Dump(nsName)

	ns, err := dm.Namespace(ksApp, nsName)
	if err != nil {
		return nil, nil, err
	}

	c, err := dm.Component(ksApp, nsName, cName)
	if err != nil {
		return nil, nil, err
	}

	return ns, c, nil
}

func (dm *defaultManager) isComponentDir(ksApp app.App, path string) (bool, error) {
	parts := strings.Split(path, "/")
	dir := filepath.Join(append([]string{ksApp.Root(), componentsRoot}, parts...)...)
	dir = filepath.Clean(dir)

	return afero.DirExists(ksApp.Fs(), dir)
}

func (dm *defaultManager) checkComponent(ksApp app.App, name string) (string, string, error) {
	parts := strings.Split(name, "/")
	base := filepath.Join(append([]string{ksApp.Root(), componentsRoot}, parts...)...)
	base = filepath.Clean(base)

	exts := []string{".yaml", ".jsonnet", ".json"}
	for _, ext := range exts {
		exists, err := afero.Exists(ksApp.Fs(), base+ext)
		if err != nil {
			return "", "", errors.Wrap(err, "check for component")
		}

		if exists {
			dir, file := path.Split(base)
			nsName := strings.TrimPrefix(dir, path.Join(ksApp.Root(), componentsRoot))
			if len(nsName) > 0 {
				nsName = strings.TrimSuffix(nsName, "/")
			}

			return nsName, file, nil
		}
	}

	return "", "", errors.Errorf("%q is not a component or a namespace", name)
}

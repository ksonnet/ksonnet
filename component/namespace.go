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
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/pkg/params"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

func nsErrorMsg(format, nsName string) string {
	s := fmt.Sprintf("namespace %q", nsName)
	if nsName == "" {
		s = "root namespace"
	}

	return fmt.Sprintf(format, s)
}

// Namespace is a component namespace.
type Namespace struct {
	path string

	app app.App
}

// NewNamespace creates an instance of Namespace.
func NewNamespace(ksApp app.App, path string) Namespace {
	return Namespace{app: ksApp, path: path}
}

// ExtractNamespacedComponent extracts a namespace and a component from a path.
func ExtractNamespacedComponent(a app.App, path string) (Namespace, string) {
	nsPath, component := filepath.Split(path)
	ns := Namespace{path: nsPath, app: a}
	return ns, component
}

// Name returns the namespace name.
func (n *Namespace) Name() string {
	if n.path == "" {
		return "/"
	}
	return n.path
}

// GetNamespace gets a namespace by path.
func GetNamespace(a app.App, nsName string) (Namespace, error) {
	parts := strings.Split(nsName, "/")
	nsDir := filepath.Join(append([]string{a.Root(), componentsRoot}, parts...)...)

	exists, err := afero.Exists(a.Fs(), nsDir)
	if err != nil {
		return Namespace{}, err
	}

	if !exists {
		return Namespace{}, errors.New(nsErrorMsg("unable to find %s", nsName))
	}

	return Namespace{path: nsName, app: a}, nil
}

// ParamsPath generates the path to params.libsonnet for a namespace.
func (n *Namespace) ParamsPath() string {
	return filepath.Join(n.Dir(), paramsFile)
}

// SetParam sets params for a namespace.
func (n *Namespace) SetParam(path []string, value interface{}) error {
	paramsData, err := n.readParams()
	if err != nil {
		return err
	}

	updatedParams, err := params.Set(path, paramsData, "", value, "global")
	if err != nil {
		return err
	}

	if err = n.writeParams(updatedParams); err != nil {
		return err
	}

	return nil
}

func (n *Namespace) writeParams(src string) error {
	return afero.WriteFile(n.app.Fs(), n.ParamsPath(), []byte(src), 0644)
}

// Dir is the absolute directory for a namespace.
func (n *Namespace) Dir() string {
	parts := strings.Split(n.path, "/")
	path := []string{n.app.Root(), componentsRoot}
	if len(n.path) != 0 {
		path = append(path, parts...)
	}

	return filepath.Join(path...)
}

// NamespaceParameter is a namespaced paramater.
type NamespaceParameter struct {
	Component string
	Index     string
	Key       string
	Value     string
}

// ResolvedParams resolves paramaters for a namespace. It returns a JSON encoded
// string of component parameters.
func (n *Namespace) ResolvedParams() (string, error) {
	s, err := n.readParams()
	if err != nil {
		return "", err
	}

	return applyGlobals(s)
}

// Params returns the params for a namespace.
func (n *Namespace) Params(envName string) ([]NamespaceParameter, error) {
	components, err := n.Components()
	if err != nil {
		return nil, err
	}

	var nsps []NamespaceParameter
	for _, c := range components {
		params, err := c.Params(envName)
		if err != nil {
			return nil, err
		}

		for _, p := range params {
			nsps = append(nsps, p)
		}
	}

	return nsps, nil
}

func (n *Namespace) readParams() (string, error) {
	b, err := afero.ReadFile(n.app.Fs(), n.ParamsPath())
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// NamespacesFromEnv returns all namespaces given an environment.
func NamespacesFromEnv(a app.App, env string) ([]Namespace, error) {
	paths, err := MakePaths(a, env)
	if err != nil {
		return nil, err
	}

	prefix := a.Root() + "/components"

	seen := make(map[string]bool)
	var namespaces []Namespace
	for _, path := range paths {
		nsName := strings.TrimPrefix(path, prefix)
		if _, ok := seen[nsName]; !ok {
			seen[nsName] = true
			ns, err := GetNamespace(a, nsName)
			if err != nil {
				return nil, err
			}

			namespaces = append(namespaces, ns)
		}
	}

	return namespaces, nil
}

// Namespaces returns all component namespaces
func Namespaces(a app.App) ([]Namespace, error) {
	componentRoot := filepath.Join(a.Root(), componentsRoot)

	var namespaces []Namespace

	err := afero.Walk(a.Fs(), componentRoot, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			ok, err := isComponentDir(a.Fs(), path)
			if err != nil {
				return err
			}

			if ok {
				nsPath := strings.TrimPrefix(path, componentRoot)
				nsPath = strings.TrimPrefix(nsPath, string(filepath.Separator))
				ns := Namespace{path: nsPath, app: a}
				namespaces = append(namespaces, ns)
			}
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "walk component path")
	}

	sort.Slice(namespaces, func(i, j int) bool {
		return namespaces[i].Name() < namespaces[j].Name()
	})

	return namespaces, nil
}

// Components returns the components in a namespace.
func (n *Namespace) Components() ([]Component, error) {
	parts := strings.Split(n.path, "/")
	nsDir := filepath.Join(append([]string{n.app.Root(), componentsRoot}, parts...)...)

	fis, err := afero.ReadDir(n.app.Fs(), nsDir)
	if err != nil {
		return nil, err
	}

	var components []Component
	for _, fi := range fis {

		ext := filepath.Ext(fi.Name())
		path := filepath.Join(nsDir, fi.Name())

		switch ext {
		// TODO: these should be constants
		case ".yaml", ".json":
			component := NewYAML(n.app, n.Name(), path, n.ParamsPath())
			components = append(components, component)
		case ".jsonnet":
			component := NewJsonnet(n.app, n.Name(), path, n.ParamsPath())
			components = append(components, component)
		}
	}

	return components, nil
}

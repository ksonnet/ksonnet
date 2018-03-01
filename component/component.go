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

package component

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
)

const (
	// componentsDir is the name of the directory which houses components.
	componentsRoot = "components"
	// paramsFile is the params file for a component namespace.
	paramsFile = "params.libsonnet"
)

// Path returns returns the file system path for a component.
func Path(fs afero.Fs, root, name string) (string, error) {
	ns, localName := ExtractNamespacedComponent(fs, root, name)

	fis, err := afero.ReadDir(fs, ns.Dir())
	if err != nil {
		return "", err
	}

	var fileName string
	files := make(map[string]bool)

	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}

		base := strings.TrimSuffix(fi.Name(), filepath.Ext(fi.Name()))
		if _, ok := files[base]; ok {
			return "", errors.Errorf("Found multiple component files with component name %q", name)
		}
		files[base] = true

		if base == localName {
			fileName = fi.Name()
		}
	}

	if fileName == "" {
		return "", errors.Errorf("No component name %q found", name)
	}

	return filepath.Join(ns.Dir(), fileName), nil
}

// Namespace is a component namespace.
type Namespace struct {
	// Path is the path of the component namespace.
	Path string

	root string
	fs   afero.Fs
}

// NewNamespace creates an an instance of Namespace.
func NewNamespace(fs afero.Fs, root, name string) Namespace {
	return Namespace{
		Path: name,
		root: root,
		fs:   fs,
	}
}

// ExtractNamespacedComponent extracts a namespace and a component from a path.
func ExtractNamespacedComponent(fs afero.Fs, root, path string) (Namespace, string) {
	path, component := filepath.Split(path)
	path = strings.TrimSuffix(path, "/")
	ns := Namespace{Path: path, root: root, fs: fs}
	return ns, component
}

// ParamsPath generates the path to params.libsonnet for a namespace.
func (n *Namespace) ParamsPath() string {
	return filepath.Join(n.Dir(), paramsFile)
}

// ComponentPaths are the absolute paths to all the components in a namespace.
func (n *Namespace) ComponentPaths() ([]string, error) {
	dir := n.Dir()
	fis, err := afero.ReadDir(n.fs, dir)
	if err != nil {
		return nil, errors.Wrap(err, "read component dir")
	}

	var paths []string
	for _, fi := range fis {
		if fi.IsDir() {
			continue
		}

		if strings.HasSuffix(fi.Name(), ".jsonnet") {
			paths = append(paths, filepath.Join(dir, fi.Name()))
		}
	}

	sort.Strings(paths)

	return paths, nil
}

// Components returns the components in a namespace.
func (n *Namespace) Components() ([]string, error) {
	paths, err := n.ComponentPaths()
	if err != nil {
		return nil, err
	}

	dir := filepath.Join(n.root, componentsRoot) + "/"

	var names []string
	for _, path := range paths {
		name := strings.TrimPrefix(path, dir)
		name = strings.TrimSuffix(name, filepath.Ext(name))
		names = append(names, name)
	}

	return names, nil
}

// Dir is the absolute directory for a namespace.
func (n *Namespace) Dir() string {
	path := []string{n.root, componentsRoot}
	if n.Path != "" {
		path = append(path, strings.Split(n.Path, "/")...)
	}

	return filepath.Join(path...)
}

// Namespaces returns all component namespaces
func Namespaces(fs afero.Fs, root string) ([]Namespace, error) {
	componentRoot := filepath.Join(root, componentsRoot)

	var namespaces []Namespace

	err := afero.Walk(fs, componentRoot, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if fi.IsDir() {
			ok, err := isComponentDir(fs, path)
			if err != nil {
				return err
			}

			if ok {
				nsPath := strings.TrimPrefix(path, componentRoot)
				nsPath = strings.TrimPrefix(nsPath, "/")
				ns := Namespace{Path: nsPath, fs: fs, root: root}
				namespaces = append(namespaces, ns)
			}
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "walk component path")
	}

	sort.Slice(namespaces, func(i, j int) bool {
		return namespaces[i].Path < namespaces[j].Path
	})

	return namespaces, nil
}

func isComponentDir(fs afero.Fs, path string) (bool, error) {
	files, err := afero.ReadDir(fs, path)
	if err != nil {
		return false, errors.Wrapf(err, "read files in %s", path)
	}

	for _, file := range files {
		if file.Name() == paramsFile {
			return true, nil
		}
	}

	return false, nil
}

// MakePathsByNamespace creates a map of component paths categorized by namespace.
func MakePathsByNamespace(fs afero.Fs, ksApp app.App, root, env string) (map[Namespace][]string, error) {
	paths, err := MakePaths(fs, ksApp, root, env)
	if err != nil {
		return nil, err
	}

	m := make(map[Namespace][]string)

	for i := range paths {
		prefix := root + "/components/"
		if strings.HasSuffix(root, "/") {
			prefix = root + "components/"
		}
		path := strings.TrimPrefix(paths[i], prefix)
		ns, _ := ExtractNamespacedComponent(fs, root, path)
		if _, ok := m[ns]; !ok {
			m[ns] = make([]string, 0)
		}

		m[ns] = append(m[ns], paths[i])
	}

	return m, nil
}

// MakePaths creates a slice of component paths
func MakePaths(fs afero.Fs, ksApp app.App, root, env string) ([]string, error) {
	cpl, err := newComponentPathLocator(fs, ksApp, env)
	if err != nil {
		return nil, errors.Wrap(err, "create component path locator")
	}

	return cpl.Locate(root)
}

type componentPathLocator struct {
	fs      afero.Fs
	envSpec *app.EnvironmentSpec
}

func newComponentPathLocator(fs afero.Fs, ksApp app.App, env string) (*componentPathLocator, error) {
	if ksApp == nil {
		return nil, errors.New("app is nil")
	}

	if fs == nil {
		return nil, errors.New("fs is nil")
	}

	envSpec, err := ksApp.Environment(env)
	if err != nil {
		return nil, errors.Errorf("can't find %s environment", env)
	}

	return &componentPathLocator{
		fs:      fs,
		envSpec: envSpec,
	}, nil
}

func (cpl *componentPathLocator) Locate(root string) ([]string, error) {
	if len(cpl.envSpec.Targets) == 0 {
		return cpl.defaultPaths(root)
	}

	var paths []string

	for _, target := range cpl.envSpec.Targets {
		childPaths, err := cpl.expandPath(root, target)
		if err != nil {
			return nil, errors.Wrapf(err, "unable to expand %s", target)
		}
		paths = append(paths, childPaths...)
	}

	sort.Strings(paths)

	return paths, nil
}

// expandPath take a root and a target and returns all the jsonnet components in descendant paths.
func (cpl *componentPathLocator) expandPath(root, target string) ([]string, error) {
	path := filepath.Join(root, componentsRoot, target)
	fi, err := cpl.fs.Stat(path)
	if err != nil {
		return nil, err
	}

	var paths []string

	walkFn := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fi.IsDir() && isComponent(path) {
			paths = append(paths, path)
		}

		return nil
	}

	if fi.IsDir() {
		rootPath := filepath.Join(root, componentsRoot, fi.Name())
		if err := afero.Walk(cpl.fs, rootPath, walkFn); err != nil {
			return nil, errors.Wrapf(err, "search for components in %s", fi.Name())
		}
	} else if isComponent(fi.Name()) {
		paths = append(paths, path)
	}

	return paths, nil
}

func (cpl *componentPathLocator) defaultPaths(root string) ([]string, error) {
	var paths []string

	walkFn := func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !fi.IsDir() && isComponent(path) {
			paths = append(paths, path)
		}

		return nil
	}

	componentRoot := filepath.Join(root, componentsRoot)

	if err := afero.Walk(cpl.fs, componentRoot, walkFn); err != nil {
		return nil, errors.Wrap(err, "search for components")
	}

	return paths, nil
}

// isComponent reports if a file is a component. Components have a `jsonnet` extension.
func isComponent(path string) bool {
	return filepath.Ext(path) == ".jsonnet"
}

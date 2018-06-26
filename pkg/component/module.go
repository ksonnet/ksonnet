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
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/google/go-jsonnet/ast"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/astext"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/printer"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/params"
	"github.com/ksonnet/ksonnet/pkg/util/jsonnet"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

func moduleErrorMsg(format, module string) string {
	s := fmt.Sprintf("module %q", module)
	if module == "" {
		s = "root module"
	}

	return fmt.Sprintf(format, s)
}

// Module is a component module
type Module interface {
	// Components returns a slice of components in this module.
	Components() ([]Component, error)
	// DeleteParam deletes a parameter.
	DeleteParam(path []string) error
	// Dir returns the directory for the module.
	Dir() string
	// Name is the name of the module.
	Name() string
	// Params returns parameters defined in this module.
	Params(envName string) ([]ModuleParameter, error)
	// ParamsPath returns the path of the parameters for this module
	ParamsPath() string
	// paramsSource returns the source of the params for this module.
	ParamsSource() (io.ReadCloser, error)
	// Render renders the components in the module to a Jsonnet object.
	Render(envName string, componentNames ...string) (*astext.Object, map[string]string, error)
	// ResolvedParams evaluates the parameters for a module within an environment.
	ResolvedParams(envName string) (string, error)
	// SetParam sets a parameter for module.
	SetParam(path []string, value interface{}) error
}

// FilesystemModule is a component module that uses a filesystem for storage.
type FilesystemModule struct {
	path string

	app app.App
}

var _ Module = (*FilesystemModule)(nil)

// NewModule creates an instance of module.
func NewModule(ksApp app.App, path string) *FilesystemModule {
	return &FilesystemModule{app: ksApp, path: path}
}

// ExtractModuleComponent extracts a module and a component from a filesystem path.
func ExtractModuleComponent(a app.App, path string) (Module, string) {
	dir, file := filepath.Split(path)
	componentName := strings.TrimSuffix(file, filepath.Ext(file))

	componentRoot := filepath.Join(a.Root(), componentsRoot)
	moduleDir := strings.TrimPrefix(dir, componentRoot)

	m := &FilesystemModule{path: dirToModule(moduleDir), app: a}

	return m, componentName
}

// FromName returns a module and a component name given a component description.
// Component descriptions can be one of the following:
//   module.component
//   component
func FromName(name string) (string, string) {
	parts := strings.Split(name, ".")

	var moduleName, componentName string

	switch len(parts) {
	case 2:
		moduleName = parts[0]
		componentName = parts[1]
	case 1:
		componentName = parts[0]
	}

	return moduleName, componentName
}

// ModuleFromPath returns a module name from a file system path.
func ModuleFromPath(a app.App, path string) string {
	componentRoot := filepath.Join(a.Root(), componentsRoot)
	moduleName := strings.TrimPrefix(path, componentRoot)
	moduleName = strings.TrimPrefix(moduleName, string(filepath.Separator))

	return strings.Replace(moduleName, string(filepath.Separator), ".", -1)
}

// Name returns the module name.
func (m *FilesystemModule) Name() string {
	if m.path == "" {
		return "/"
	}
	return m.path
}

// GetModule gets a module by path.
func GetModule(a app.App, moduleName string) (Module, error) {
	if moduleName == "/" {
		moduleName = ""
	}

	if !isValidModuleName(moduleName) {
		return nil, errors.Errorf("%q is an invalid module name", moduleName)
	}

	parts := strings.Split(moduleName, ".")
	moduleDir := filepath.Join(append([]string{a.Root(), componentsRoot}, parts...)...)

	exists, err := afero.Exists(a.Fs(), moduleDir)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, errors.New(moduleErrorMsg("unable to find %s", moduleName))
	}

	return &FilesystemModule{path: moduleName, app: a}, nil
}

var (
	reValidModule = regexp.MustCompile(`^([_A-Za-z0-9]?[A-Za-z0-9\-_]+(\.[A-Za-z0-9\-_]+)*)?$`)
)

func isValidModuleName(name string) bool {
	if name == "." {
		return true
	}
	return reValidModule.MatchString(name)
}

// ParamsPath generates the path to params.libsonnet for a module.
func (m *FilesystemModule) ParamsPath() string {
	return filepath.Join(m.Dir(), paramsFile)
}

// SetParam sets params for a module.
func (m *FilesystemModule) SetParam(path []string, value interface{}) error {
	paramsData, err := m.readParams()
	if err != nil {
		return err
	}

	updated, err := params.SetInObject(path, paramsData, "", value, "global")
	if err != nil {
		return err
	}

	return m.writeParams(updated)
}

// DeleteParam deletes params for a module.
func (m *FilesystemModule) DeleteParam(path []string) error {
	paramsData, err := m.readParams()
	if err != nil {
		return err
	}

	updated, err := params.DeleteFromObject(path, paramsData, "", "global")
	if err != nil {
		return err
	}

	return m.writeParams(updated)
}

func (m *FilesystemModule) writeParams(src string) error {
	return afero.WriteFile(m.app.Fs(), m.ParamsPath(), []byte(src), 0644)
}

// Dir is the absolute directory for a module.
func (m *FilesystemModule) Dir() string {
	return filepath.Join(m.app.Root(), componentsRoot, moduleToDir(m.path))
}

// ParamsSource returns the source of params for a module as a reader.
func (m *FilesystemModule) ParamsSource() (io.ReadCloser, error) {
	path := filepath.Join(m.Dir(), paramsFile)
	f, err := m.app.Fs().Open(path)
	if err != nil {
		return nil, errors.Wrapf(err, "opening %q", path)
	}

	return f, nil
}

// ModuleParameter is a module parameter.
type ModuleParameter struct {
	Component string
	Key       string
	Value     string
}

// IsSameType returns true if the other ModuleParams is the same type. The types
// are the same if the component, index, and key match.
func (mp *ModuleParameter) IsSameType(other ModuleParameter) bool {
	return mp.Component == other.Component &&
		mp.Key == other.Key
}

// ResolvedParams resolves paramaters for a module. It returns a JSON encoded
// string of component parameters.
func (m *FilesystemModule) ResolvedParams(envName string) (string, error) {
	s, err := m.readParams()
	if err != nil {
		return "", err
	}

	envCode, err := params.JsonnetEnvObject(m.app, envName)
	if err != nil {
		return "", errors.Wrap(err, "building environment argument")
	}

	vm := jsonnet.NewVM()
	vm.AddJPath(
		filepath.Join(m.app.Root(), "vendor"),
		filepath.Join(m.app.Root(), "lib"),
	)

	vm.ExtCode("__ksonnet/environments", envCode)

	output, err := vm.EvaluateSnippet("params.libsonnet", s)
	if err != nil {
		return "", errors.Wrap(err, "evaluating params.libsonnet")
	}

	n, err := jsonnet.ParseNode("params.libsonnet", output)
	if err != nil {
		return "", errors.Wrap(err, "parsing parameters")
	}

	object, ok := n.(*astext.Object)
	if !ok {
		return "", errors.Errorf("params.libsonnet did not evaluate to an object (%T)", n)
	}

	var componentsObject *astext.Object

	for _, f := range object.Fields {
		id, err := jsonnet.FieldID(f)
		if err != nil {
			return "", err
		}

		if id == "components" {
			componentsObject = f.Expr2.(*astext.Object)
		}
	}

	if componentsObject == nil {
		return "", errors.New("could not find components object in params")
	}

	currentFields := make(map[string]bool)

	for _, f := range componentsObject.Fields {
		id, err := jsonnet.FieldID(f)
		if err != nil {
			return "", err
		}

		currentFields[id] = true
	}

	var buf bytes.Buffer
	if err := printer.Fprint(&buf, object); err != nil {
		return "", errors.Wrap(err, "could not update params")
	}

	return applyGlobals(buf.String())
}

// Params returns the params for a module.
func (m *FilesystemModule) Params(envName string) ([]ModuleParameter, error) {
	m.log().Debug("list module params")

	components, err := m.Components()
	if err != nil {
		return nil, err
	}

	var moduleParameters []ModuleParameter
	for _, c := range components {
		params, err := c.Params(envName)
		if err != nil {
			return nil, err
		}

		for _, p := range params {
			moduleParameters = append(moduleParameters, p)
		}
	}

	return moduleParameters, nil
}

func (m *FilesystemModule) readParams() (string, error) {
	b, err := afero.ReadFile(m.app.Fs(), m.ParamsPath())
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// ModulesFromEnv returns all modules given an environment.
func ModulesFromEnv(a app.App, env string) ([]Module, error) {
	paths, err := MakePaths(a, env)
	if err != nil {
		return nil, err
	}

	seen := make(map[string]bool)
	var modules []Module
	for _, path := range paths {
		module := ModuleFromPath(a, path)
		if _, ok := seen[module]; !ok {
			seen[module] = true
			m, err := GetModule(a, module)
			if err != nil {
				return nil, err
			}

			modules = append(modules, m)
		}
	}

	return modules, nil
}

// Modules returns all component modules
func Modules(a app.App) ([]Module, error) {
	componentRoot := filepath.Join(a.Root(), componentsRoot)

	var modules []Module

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
				modulePath := strings.TrimPrefix(path, componentRoot)
				modulePath = strings.TrimPrefix(modulePath, string(filepath.Separator))
				m := &FilesystemModule{path: modulePath, app: a}
				modules = append(modules, m)
			}
		}

		return nil
	})

	if err != nil {
		return nil, errors.Wrap(err, "walk component path")
	}

	sort.Slice(modules, func(i, j int) bool {
		return modules[i].Name() < modules[j].Name()
	})

	return modules, nil
}

// Components returns the components in a module.
func (m *FilesystemModule) Components() ([]Component, error) {
	parts := strings.Split(m.path, ".")
	moduleDir := filepath.Join(append([]string{m.app.Root(), componentsRoot}, parts...)...)

	fis, err := afero.ReadDir(m.app.Fs(), moduleDir)
	if err != nil {
		return nil, err
	}

	var components []Component
	for _, fi := range fis {

		ext := filepath.Ext(fi.Name())
		path := filepath.Join(moduleDir, fi.Name())

		switch ext {
		// TODO: these should be constants
		case ".yaml", ".json":
			component := NewYAML(m.app, m.Name(), path, m.ParamsPath())
			components = append(components, component)
		case ".jsonnet":
			component := NewJsonnet(m.app, m.Name(), path, m.ParamsPath())
			components = append(components, component)
		}
	}

	return components, nil
}

// Render converts components to JSON. If there are component names, only include
// those components.
func (m *FilesystemModule) Render(envName string, componentNames ...string) (*astext.Object, map[string]string, error) {
	components, err := m.Components()
	if err != nil {
		return nil, nil, err
	}

	doc := &astext.Object{
		Fields: astext.ObjectFields{},
	}

	componentMap := make(map[string]string)

	for _, c := range components {
		name, node, err := c.ToNode(envName)
		if err != nil {
			return nil, nil, err
		}

		f, err := astext.CreateField(name)
		if err != nil {
			return nil, nil, err
		}

		f.Hide = ast.ObjectFieldInherit

		f.Expr2 = node
		doc.Fields = append(doc.Fields, *f)

		componentMap[c.Name(true)] = c.Type()
	}

	return doc, componentMap, nil
}

func (m *FilesystemModule) log() *logrus.Entry {
	return logrus.WithField("module-name", m.Name())
}

// moduleToDir converts a module to a filesystem directory by replacing
// the . separator with a filesystem separator.
func moduleToDir(s string) string {
	return strings.Replace(s, ".", string(filepath.Separator), -1)
}

// dirToModule converts filesystem directory to a module by replacing
// filesystem separators with ".".
func dirToModule(s string) string {
	s = strings.Replace(s, string(filepath.Separator), ".", -1)
	return strings.TrimSuffix(s, ".")
}

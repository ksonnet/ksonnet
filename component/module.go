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
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/google/go-jsonnet/ast"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/astext"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/printer"
	"github.com/ksonnet/ksonnet/metadata/app"
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
	Components() ([]Component, error)
	DeleteParam(path []string) error
	Dir() string
	Name() string
	Params(envName string) ([]ModuleParameter, error)
	ParamsPath() string
	Render(envName string, componentNames ...string) (*astext.Object, map[string]string, error)
	ResolvedParams() (string, error)
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

// ExtractModuleComponent extracts a module and a component from a path.
func ExtractModuleComponent(a app.App, path string) (Module, string) {
	modulePath, component := filepath.Split(path)
	m := &FilesystemModule{path: modulePath, app: a}
	return m, component
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
	parts := strings.Split(moduleName, "/")
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
	parts := strings.Split(m.path, "/")
	path := []string{m.app.Root(), componentsRoot}
	if len(m.path) != 0 {
		path = append(path, parts...)
	}

	return filepath.Join(path...)
}

// ModuleParameter is a module parameter.
type ModuleParameter struct {
	Component string
	Index     string
	Key       string
	Value     string
}

// IsSameType returns true if the other ModuleParams is the same type. The types
// are the same if the component, index, and key match.
func (mp *ModuleParameter) IsSameType(other ModuleParameter) bool {
	return mp.Component == other.Component &&
		mp.Index == other.Index &&
		mp.Key == other.Key
}

// ResolvedParams resolves paramaters for a module. It returns a JSON encoded
// string of component parameters.
func (m *FilesystemModule) ResolvedParams() (string, error) {
	s, err := m.readParams()
	if err != nil {
		return "", err
	}

	object, err := jsonnet.Parse("params.libsonnet", s)
	if err != nil {
		return "", err
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

	components, err := m.Components()
	if err != nil {
		return "", err
	}

	for _, c := range components {
		summaries, err := c.Summarize()
		if err != nil {
			return "", err
		}

		for i := range summaries {
			summary := summaries[i]
			if summary.Type != "yaml" && summary.Type != "json" {
				continue
			}

			name := fmt.Sprintf("%s-%d", summary.ComponentName, summary.Index)
			if _, ok := currentFields[name]; !ok {
				field, err := astext.CreateField(name)
				if err != nil {
					return "", err
				}

				field.Hide = ast.ObjectFieldInherit
				field.Expr2 = &astext.Object{}
				componentsObject.Fields = append(componentsObject.Fields, *field)
			}
		}
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

	prefix := a.Root() + "/components"

	seen := make(map[string]bool)
	var modules []Module
	for _, path := range paths {
		module := strings.TrimPrefix(path, prefix)
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
	parts := strings.Split(m.path, "/")
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

type renderedModule struct {
	Components map[string]interface{} `json:"components"`
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
		m, err := c.ToMap(envName)
		if err != nil {
			return nil, nil, err
		}

		for k, v := range m {
			f, err := astext.CreateField(k)
			if err != nil {
				return nil, nil, err
			}

			f.Hide = ast.ObjectFieldInherit

			f.Expr2 = v
			doc.Fields = append(doc.Fields, *f)
		}

		componentMap[c.Name(false)] = c.Type()
	}

	return doc, componentMap, nil
}

func (m *FilesystemModule) log() *logrus.Entry {
	return logrus.WithField("module-name", m.Name())
}

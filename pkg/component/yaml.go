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
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ghodss/yaml"
	"github.com/google/go-jsonnet/ast"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/params"
	"github.com/ksonnet/ksonnet/pkg/schema"
	jsonnetutil "github.com/ksonnet/ksonnet/pkg/util/jsonnet"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const (
	// TypeYAML is a YAML component.
	TypeYAML = "yaml"

	paramsComponentRoot = "components"
)

// YAML represents a YAML component. Since JSON is a subset of YAML, it can handle JSON as well.
type YAML struct {
	app        app.App
	module     string
	source     string
	paramsPath string
}

var _ Component = (*YAML)(nil)

// NewYAML creates an instance of YAML.
func NewYAML(a app.App, module, source, paramsPath string) *YAML {
	return &YAML{
		app:        a,
		module:     module,
		source:     source,
		paramsPath: paramsPath,
	}
}

// Name is the component name.
func (y *YAML) Name(wantsNameSpaced bool) string {
	base := filepath.Base(y.source)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	if !wantsNameSpaced {
		return name
	}

	if y.module == "/" || y.module == "" {
		return name
	}

	return strings.Join([]string{y.module, name}, ".")
}

// Type always returns "yaml".
func (y *YAML) Type() string {
	return TypeYAML
}

// Remove removes the component.
func (y *YAML) Remove() error {
	m := NewModule(y.app, y.module)
	path := filepath.Join(m.Dir(), y.Name(false)+"."+y.Type())
	if err := y.app.Fs().Remove(path); err != nil {
		return errors.Wrapf(err, "removing %q", path)
	}

	return nil
}

// Params returns params for a component.
func (y *YAML) Params(envName string) ([]ModuleParameter, error) {
	y.log().WithField("env-name", envName).Debug("getting component params")

	ve, err := schema.ValueExtractorFactory()
	if err != nil {
		return nil, err
	}

	// find all the params for this component
	paramsData, err := y.readParams(envName)
	if err != nil {
		return nil, err
	}

	componentParams, err := params.ToMap(y.Name(false), paramsData, paramsComponentRoot)
	if err != nil {
		return nil, errors.Wrap(err, "could not find components")
	}

	ts, props, err := y.read()
	if err != nil {
		if err == schema.ErrEmptyYAML {
			return make([]ModuleParameter, 0), nil
		}
		return nil, err
	}

	valueMap, err := ve.Extract(ts.GVK(), props)
	if err != nil {
		return nil, err
	}

	return y.paramValues(y.Name(true), valueMap, componentParams, nil)
}

func isLeaf(path []string, key string, valueMap map[string]schema.Values) (string, bool) {
	childPath := strings.Join(append(path, key), ".")
	for _, v := range valueMap {
		if strings.Join(v.Lookup, ".") == childPath {
			return childPath, true
		}
	}

	return "", false
}

func (y *YAML) paramValues(componentName string, valueMap map[string]schema.Values, paramMap map[string]interface{}, path []string) ([]ModuleParameter, error) {
	y.log().WithFields(logrus.Fields{
		"prop-name": componentName,
		"value-map": fmt.Sprintf("%#v", valueMap),
		"param-map": fmt.Sprintf("%#v", paramMap),
		"path":      path,
	}).Debug("finding param values")
	var params []ModuleParameter

	for k, v := range paramMap {
		var s string
		switch t := v.(type) {
		default:
			if childPath, exists := isLeaf(path, k, valueMap); exists {
				s = fmt.Sprintf("%v", v)
				p := ModuleParameter{
					Component: componentName,
					Key:       childPath,
					Value:     s,
				}
				params = append(params, p)
			}
		case map[string]interface{}:
			if childPath, exists := isLeaf(path, k, valueMap); exists {
				b, err := json.Marshal(&v)
				if err != nil {
					return nil, err
				}
				s = string(b)
				p := ModuleParameter{
					Component: componentName,
					Key:       childPath,
					Value:     s,
				}
				params = append(params, p)
			} else {
				childPath := append(path, k)
				childParams, err := y.paramValues(componentName, valueMap, t, childPath)
				if err != nil {
					return nil, err
				}

				if len(childParams) == 0 {
					b, err := json.Marshal(&v)
					if err != nil {
						return nil, err
					}
					s = string(b)

					childParams = []ModuleParameter{
						{
							Component: componentName,
							Key:       strings.Join(childPath, "."),
							Value:     s,
						},
					}
				}

				params = append(params, childParams...)
			}
		case []interface{}:
			if childPath, exists := isLeaf(path, k, valueMap); exists {
				b, err := json.Marshal(&v)
				if err != nil {
					return nil, err
				}
				s = string(b)
				p := ModuleParameter{
					Component: componentName,
					Key:       childPath,
					Value:     s,
				}
				params = append(params, p)
			}
		}
	}

	if len(params) == 0 {
		y.log().Debug("there are no params")
	}
	return params, nil
}

// SetParam set parameter for a component.
func (y *YAML) SetParam(path []string, value interface{}) error {
	paramsData, err := y.readModuleParams()
	if err != nil {
		return err
	}

	updatedParams, err := params.SetInObject(path, paramsData, y.Name(false), value, paramsComponentRoot)
	if err != nil {
		return err
	}

	if err = y.writeParams(updatedParams); err != nil {
		return err
	}

	return nil
}

// DeleteParam deletes a param.
func (y *YAML) DeleteParam(path []string) error {
	paramsData, err := y.readModuleParams()
	if err != nil {
		return err
	}

	updatedParams, err := params.DeleteFromObject(path, paramsData, y.Name(false), paramsComponentRoot)
	if err != nil {
		return err
	}

	if err = y.writeParams(updatedParams); err != nil {
		return err
	}

	return nil
}

func (y *YAML) readParams(envName string) (string, error) {
	if envName == "" {
		return y.readModuleParams()
	}

	return envParams(y.app, y.module, envName)
}

func (y *YAML) readModuleParams() (string, error) {
	b, err := afero.ReadFile(y.app.Fs(), y.paramsPath)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (y *YAML) writeParams(src string) error {
	return afero.WriteFile(y.app.Fs(), y.paramsPath, []byte(src), 0644)
}

// Summarize generates a summary for a YAML component. For each manifest, it will
// return a slice of summaries of resources described.
func (y *YAML) Summarize() (Summary, error) {
	ts, props, err := y.read()
	if err != nil {
		if err == schema.ErrEmptyYAML {
			return Summary{}, nil
		}
		return Summary{}, err
	}

	name, err := props.Name()
	if err != nil {
		return Summary{}, err
	}

	return Summary{
		ComponentName: y.Name(true),
		Type:          y.ext(),
		APIVersion:    ts.APIVersion,
		Kind:          ts.RawKind,
		Name:          name,
	}, nil
}

// ToNode converts a YAML component to a Jsonnet node.
func (y *YAML) ToNode(envName string) (string, ast.Node, error) {
	key := y.Name(false)
	data, err := afero.ReadFile(y.app.Fs(), y.source)
	if err != nil {
		return "", nil, err
	}

	if len(data) == 0 {
		return "", nil, errors.New("object was empty")
	}

	data, err = yaml.YAMLToJSON(data)
	if err != nil {
		return "", nil, err
	}

	patchedData, err := y.applyParams(key, string(data))
	if err != nil {
		return "", nil, err
	}

	o, err := jsonnetutil.Parse(y.source, patchedData)
	if err != nil {
		return "", nil, err
	}

	return y.Name(true), o, nil
}

func (y *YAML) applyParams(componentName, data string) (string, error) {
	paramsData, err := afero.ReadFile(y.app.Fs(), y.paramsPath)
	if err != nil {
		return "", err
	}

	return params.PatchJSON(data, string(paramsData), componentName)
}

func (y *YAML) ext() string {
	return strings.TrimPrefix(filepath.Ext(y.source), ".")
}

func (y *YAML) log() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"component-name": y.Name(true),
		"component-type": "YAML",
	})
}

func (y *YAML) read() (*schema.TypeSpec, schema.Properties, error) {
	f, err := y.app.Fs().Open(y.source)
	if err != nil {
		return nil, nil, err
	}

	return schema.ImportYaml(f)
}

type paramPath struct {
	path  []string
	value interface{}
}

func mapToPaths(m map[string]interface{}, lookup map[string]bool, parent []string) []paramPath {
	paths := make([]paramPath, 0)

	for k, v := range m {
		cur := append(parent, k)

		switch t := v.(type) {
		default:
			pp := paramPath{path: cur, value: v}
			paths = append(paths, pp)

		case map[string]interface{}:
			children := mapToPaths(t, lookup, cur)

			route := strings.Join(cur, ".")
			if _, ok := lookup[route]; ok {
				pp := paramPath{path: cur, value: v}
				paths = append(paths, pp)
			} else {
				paths = append(paths, children...)
			}

		}
	}

	sort.Slice(paths, func(i, j int) bool {
		a := strings.Join(paths[i].path, ".")
		b := strings.Join(paths[j].path, ".")

		return a < b
	})

	return paths
}

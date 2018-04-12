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
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/pkg/params"
	"github.com/ksonnet/ksonnet/pkg/schema"
	jsonnetutil "github.com/ksonnet/ksonnet/pkg/util/jsonnet"
	"github.com/ksonnet/ksonnet/pkg/util/k8s"
	utilyaml "github.com/ksonnet/ksonnet/pkg/util/yaml"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	amyaml "k8s.io/apimachinery/pkg/util/yaml"
)

const (
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

	if y.module == "/" {
		return name
	}

	return path.Join(y.module, name)
}

// Params returns params for a component.
func (y *YAML) Params(envName string) ([]ModuleParameter, error) {
	y.log().WithField("env-name", envName).Debug("getting component params")

	ve, err := schema.ValueExtractorFactory()
	if err != nil {
		return nil, err
	}

	// find all the params for this component
	// keys will look like `component-id`
	paramsData, err := y.readParams(envName)
	if err != nil {
		return nil, err
	}

	componentParams, err := params.ToMap("", paramsData, paramsComponentRoot)
	if err != nil {
		return nil, errors.Wrap(err, "could not find components")
	}

	re, err := regexp.Compile(fmt.Sprintf(`^%s-(\d+)$`, y.Name(false)))
	if err != nil {
		return nil, err
	}

	readers, err := utilyaml.Decode(y.app.Fs(), y.source)
	if err != nil {
		return nil, err
	}

	var params []ModuleParameter
	for paramName, paramValue := range componentParams {
		y.log().WithField("prop-name", paramName).Debug("searching for props")
		matches := re.FindAllStringSubmatch(paramName, 1)
		if len(matches) > 0 {
			index := matches[0][1]
			i, err := strconv.Atoi(index)
			if err != nil {
				return nil, err
			}

			ts, props, err := schema.ImportYaml(readers[i])
			if err != nil {
				if err == schema.ErrEmptyYAML {
					continue
				}
				return nil, err
			}

			valueMap, err := ve.Extract(ts.GVK(), props)
			if err != nil {
				return nil, err
			}

			paramMap, ok := paramValue.(map[string]interface{})
			if !ok {
				return nil, errors.Errorf("component value for %q was not a map", paramName)
			}

			childParams, err := y.paramValues(y.Name(false), index, valueMap, paramMap, nil)
			if err != nil {
				return nil, err
			}

			params = append(params, childParams...)
		}
	}

	return params, nil
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

func (y *YAML) paramValues(componentName, index string, valueMap map[string]schema.Values, paramMap map[string]interface{}, path []string) ([]ModuleParameter, error) {
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
					Index:     index,
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
					Index:     index,
					Key:       childPath,
					Value:     s,
				}
				params = append(params, p)
			} else {
				childPath := append(path, k)
				childParams, err := y.paramValues(componentName, index, valueMap, t, childPath)
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
							Index:     index,
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
					Index:     index,
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

// Objects converts YAML to a slice of apimachinery Unstructured objects. Params for a YAML
// based component are keyed like, `name-id`, where `name` is the file name sans the extension,
// and the id is the position within the file (starting at 0). Params are named this way
// because a YAML file can contain more than one object.
func (y *YAML) Objects(paramsStr, envName string) ([]*unstructured.Unstructured, error) {
	if paramsStr == "" {
		dir := filepath.Dir(y.source)
		paramsFile := filepath.Join(dir, "params.libsonnet")

		b, err := afero.ReadFile(y.app.Fs(), paramsFile)
		if err != nil {
			return nil, err
		}

		paramsStr = string(b)
	}

	return y.raw(paramsStr)
}

// SetParam set parameter for a component.
func (y *YAML) SetParam(path []string, value interface{}, options ParamOptions) error {
	entry := fmt.Sprintf("%s-%d", y.Name(false), options.Index)
	// TODO: make this work with env params
	paramsData, err := y.readParams("")
	if err != nil {
		return err
	}

	updatedParams, err := params.SetInObject(path, paramsData, entry, value, paramsComponentRoot)
	if err != nil {
		return err
	}

	if err = y.writeParams(updatedParams); err != nil {
		return err
	}

	return nil
}

// DeleteParam deletes a param.
func (y *YAML) DeleteParam(path []string, options ParamOptions) error {
	entry := fmt.Sprintf("%s-%d", y.Name(false), options.Index)
	// TODO: make this work with env params
	paramsData, err := y.readParams("")
	if err != nil {
		return err
	}

	updatedParams, err := params.DeleteFromObject(path, paramsData, entry, paramsComponentRoot)
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
		return y.readNamespaceParams()
	}

	libPath, err := y.app.LibPath(envName)
	if err != nil {
		return "", err
	}

	ns, err := GetModule(y.app, y.module)
	if err != nil {
		return "", err
	}

	paramsStr, err := ns.ResolvedParams()
	if err != nil {
		return "", err
	}

	data, err := y.app.EnvironmentParams(envName)
	if err != nil {
		return "", err
	}

	envParams := upgradeParams(envName, data)

	vm := jsonnetutil.NewVM()
	vm.JPaths = []string{
		libPath,
		filepath.Join(y.app.Root(), "environments", envName),
		filepath.Join(y.app.Root(), "vendor"),
	}
	vm.ExtCode("__ksonnet/params", paramsStr)
	return vm.EvaluateSnippet("snippet", string(envParams))
}

func (y *YAML) readNamespaceParams() (string, error) {
	b, err := afero.ReadFile(y.app.Fs(), y.paramsPath)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (y *YAML) writeParams(src string) error {
	return afero.WriteFile(y.app.Fs(), y.paramsPath, []byte(src), 0644)
}

func (y *YAML) raw(paramsStr string) ([]*unstructured.Unstructured, error) {
	objects, err := y.readObject(paramsStr)
	if err != nil {
		return nil, errors.Wrap(err, "read object")
	}

	list, err := k8s.FlattenToV1(objects)
	if err != nil {
		return nil, errors.Wrap(err, "flatten objects")
	}

	return list, nil
}

func (y *YAML) hasParams() (bool, error) {
	dir := filepath.Dir(y.source)
	paramsFile := filepath.Join(dir, "params.libsonnet")

	exists, err := afero.Exists(y.app.Fs(), paramsFile)
	if err != nil || !exists {
		return false, nil
	}

	paramsObj, err := jsonnetutil.ImportFromFs(paramsFile, y.app.Fs())
	if err != nil {
		return false, errors.Wrap(err, "import params")
	}

	componentPath := []string{
		paramsComponentRoot,
		fmt.Sprintf("%s-0", y.Name(false)),
	}
	_, err = jsonnetutil.FindObject(paramsObj, componentPath)
	if err != nil {
		return false, nil
	}

	return true, nil
}

func (y *YAML) readObject(paramsStr string) ([]runtime.Object, error) {
	f, err := y.app.Fs().Open(y.source)
	if err != nil {
		return nil, err
	}

	base := strings.TrimSuffix(filepath.Base(y.source), filepath.Ext(y.source))

	decoder := amyaml.NewYAMLReader(bufio.NewReader(f))
	ret := []runtime.Object{}
	i := 0
	for {
		componentName := fmt.Sprintf("%s-%d", base, i)
		i++
		bytes, err := decoder.Read()
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}
		if len(bytes) == 0 {
			continue
		}
		jsondata, err := amyaml.ToJSON(bytes)
		if err != nil {
			return nil, err
		}

		patched, err := patchJSON(string(jsondata), paramsStr, componentName)
		if err != nil {
			return nil, err
		}

		jsondata = []byte(patched)

		obj, _, err := unstructured.UnstructuredJSONScheme.Decode(jsondata, nil, nil)
		if err != nil {
			return nil, err
		}
		ret = append(ret, obj)
	}
	return ret, nil
}

// Summarize generates a summary for a YAML component. For each manifest, it will
// return a slice of summaries of resources described.
func (y *YAML) Summarize() ([]Summary, error) {
	var summaries []Summary

	readers, err := utilyaml.Decode(y.app.Fs(), y.source)
	if err != nil {
		return nil, err
	}

	for i, r := range readers {
		ts, props, err := schema.ImportYaml(r)
		if err != nil {
			if err == schema.ErrEmptyYAML {
				continue
			}
			return nil, err
		}

		name, err := props.Name()
		if err != nil {
			return nil, err
		}

		summary := Summary{
			ComponentName: y.Name(false),
			IndexStr:      strconv.Itoa(i),
			Type:          y.ext(),
			APIVersion:    ts.APIVersion,
			Kind:          ts.RawKind,
			Name:          name,
		}
		summaries = append(summaries, summary)
	}

	return summaries, nil
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

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
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/google/go-jsonnet/ast"
	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/astext"
	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/pkg/params"
	"github.com/ksonnet/ksonnet/pkg/util/jsonnet"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

// Jsonnet is a component base on jsonnet.
type Jsonnet struct {
	app        app.App
	module     string
	source     string
	paramsPath string

	useJsonnetMemoryImporter bool
}

var _ Component = (*Jsonnet)(nil)

// NewJsonnet creates an instance of Jsonnet.
func NewJsonnet(a app.App, module, source, paramsPath string) *Jsonnet {
	return &Jsonnet{
		app:        a,
		module:     module,
		source:     source,
		paramsPath: paramsPath,
	}
}

// Name is the name of this component.
func (j *Jsonnet) Name(wantsNameSpaced bool) string {
	base := filepath.Base(j.source)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	if !wantsNameSpaced {
		return name
	}

	if j.module == "/" {
		return name
	}

	return path.Join(j.module, name)
}

// Type always returns "jsonnet".
func (j *Jsonnet) Type() string {
	return "jsonnet"
}

func jsonWalk(obj interface{}) ([]interface{}, error) {
	switch o := obj.(type) {
	case map[string]interface{}:
		if o["kind"] != nil && o["apiVersion"] != nil {
			return []interface{}{o}, nil
		}
		ret := []interface{}{}
		for _, v := range o {
			children, err := jsonWalk(v)
			if err != nil {
				return nil, err
			}
			ret = append(ret, children...)
		}
		return ret, nil
	case []interface{}:
		ret := make([]interface{}, 0, len(o))
		for _, v := range o {
			children, err := jsonWalk(v)
			if err != nil {
				return nil, err
			}
			ret = append(ret, children...)
		}
		return ret, nil
	default:
		return nil, fmt.Errorf("Unexpected object structure: %T", o)
	}
}

func (j *Jsonnet) evaluate(paramsStr, envName string) (string, error) {
	libPath, err := j.app.LibPath(envName)
	if err != nil {
		return "", err
	}

	vm := jsonnet.NewVM()
	if j.useJsonnetMemoryImporter {
		vm.Fs = j.app.Fs()
		vm.UseMemoryImporter = true
	}

	vm.JPaths = []string{
		libPath,
		filepath.Join(j.app.Root(), "vendor"),
	}
	vm.ExtCode("__ksonnet/params", paramsStr)

	envDetails, err := j.app.Environment(envName)
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
	vm.ExtCode("__ksonnet/environments", string(marshalledDestination))

	snippet, err := afero.ReadFile(j.app.Fs(), j.source)
	if err != nil {
		return "", err
	}

	log.WithFields(log.Fields{
		"component-name": j.Name(true),
	}).Debugf("convert component to jsonnet")
	return vm.EvaluateSnippet(j.source, string(snippet))
}

// SetParam set parameter for a component.
func (j *Jsonnet) SetParam(path []string, value interface{}) error {
	// TODO: make this work at the env level too
	paramsData, err := j.readParams("")
	if err != nil {
		return err
	}

	updatedParams, err := params.SetInObject(path, paramsData, j.Name(false), value, paramsComponentRoot)
	if err != nil {
		return err
	}

	if err = j.writeParams(updatedParams); err != nil {
		return err
	}

	return nil
}

// DeleteParam deletes a param.
func (j *Jsonnet) DeleteParam(path []string) error {
	// TODO: make this work at the env level too
	paramsData, err := j.readParams("")
	if err != nil {
		return err
	}

	updatedParams, err := params.DeleteFromObject(path, paramsData, j.Name(false), paramsComponentRoot)
	if err != nil {
		return err
	}

	if err = j.writeParams(updatedParams); err != nil {
		return err
	}

	return nil
}

// Params returns params for a component.
func (j *Jsonnet) Params(envName string) ([]ModuleParameter, error) {
	j.log().WithField("env-name", envName).Debug("getting component params")

	paramsData, err := j.readParams(envName)
	if err != nil {
		return nil, err
	}

	props, err := params.ToMap(j.Name(false), paramsData, paramsComponentRoot)
	if err != nil {
		return nil, errors.Wrap(err, "could not find components")
	}

	var params []ModuleParameter
	for k, v := range props {
		vStr, err := j.paramValue(v)
		if err != nil {
			return nil, err
		}
		np := ModuleParameter{
			Component: j.Name(false),
			Key:       k,
			Index:     "0",
			Value:     vStr,
		}

		params = append(params, np)
	}

	sort.Slice(params, func(i, j int) bool {
		return params[i].Key < params[j].Key
	})

	return params, nil
}

func (j *Jsonnet) paramValue(v interface{}) (string, error) {
	switch v.(type) {
	default:
		s := fmt.Sprintf("%v", v)
		return s, nil
	case string:
		s := fmt.Sprintf("%v", v)
		return strconv.Quote(s), nil
	case map[string]interface{}, []interface{}:
		b, err := json.Marshal(&v)
		if err != nil {
			return "", err
		}

		return string(b), nil
	}
}

// Summarize creates a summary for the component.
func (j *Jsonnet) Summarize() (Summary, error) {
	return Summary{
		ComponentName: j.Name(false),
		Type:          "jsonnet",
	}, nil
}

// ToMap converts a Jsonnet component to a map of jsonnet objects.
func (j *Jsonnet) ToMap(envName string) (map[string]ast.Node, error) {
	n, err := jsonnet.ImportNodeFromFs(j.source, j.app.Fs())
	if err != nil {
		return nil, err
	}

	return map[string]ast.Node{
		j.Name(false): n,
	}, nil
}

func stringValue(n ast.Node) (string, error) {
	ls, ok := n.(*ast.LiteralString)
	if !ok {
		return "", errors.New("node was not a LiteralString")
	}

	return ls.Value, nil
}

func extractName(object *astext.Object) (string, error) {
	m, err := jsonnet.ConvertObjectToMap(object)
	if err != nil {
		return "", err
	}

	var kind string
	var name string

	if o, ok := m["kind"]; ok {
		if s, ok := o.(string); ok {
			kind = s
		}
	}

	if o, ok := m["metadata"]; ok {
		if metadataField, ok := o.(map[string]interface{}); ok {
			if s, ok := metadataField["name"].(string); ok {
				name = s
			}
		}
	}

	if kind == "" {
		return "", errors.New("object did not have kind")
	}

	if name == "" {
		return "", errors.New("object did not have name")
	}

	return fmt.Sprintf("%s-%s", name, strings.ToLower(kind)), nil
}

func (j *Jsonnet) readParams(envName string) (string, error) {
	if envName == "" {
		return j.readNamespaceParams()
	}

	ns, err := GetModule(j.app, j.module)
	if err != nil {
		return "", err
	}

	paramsStr, err := ns.ResolvedParams()
	if err != nil {
		return "", err
	}

	data, err := j.app.EnvironmentParams(envName)
	if err != nil {
		return "", err
	}

	envParams := upgradeParams(envName, data)

	env, err := j.app.Environment(envName)
	if err != nil {
		return "", err
	}

	vm := jsonnet.NewVM()
	vm.JPaths = []string{
		env.MakePath(j.app.Root()),
		filepath.Join(j.app.Root(), "vendor"),
	}
	vm.ExtCode("__ksonnet/params", paramsStr)
	return vm.EvaluateSnippet("snippet", string(envParams))
}

func (j *Jsonnet) readNamespaceParams() (string, error) {
	b, err := afero.ReadFile(j.app.Fs(), j.paramsPath)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

func (j *Jsonnet) writeParams(src string) error {
	return afero.WriteFile(j.app.Fs(), j.paramsPath, []byte(src), 0644)
}

func (j *Jsonnet) log() *log.Entry {
	return log.WithField("component-name", j.Name(true))
}

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
	"strconv"
	"strings"

	"github.com/google/go-jsonnet/ast"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/params"
	"github.com/ksonnet/ksonnet/pkg/util/jsonnet"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/afero"
)

const (
	// TypeJsonnet is a Jsonnet component.
	TypeJsonnet = "jsonnet"
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

	if j.module == "/" || j.module == "" {
		return name
	}

	return strings.Join([]string{j.module, name}, ".")
}

// Type always returns "jsonnet".
func (j *Jsonnet) Type() string {
	return TypeJsonnet
}

// SetParam set parameter for a component.
func (j *Jsonnet) SetParam(path []string, value interface{}) error {
	paramsData, err := j.readModuleParams()
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
	paramsData, err := j.readModuleParams()
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
			Component: j.Name(true),
			Key:       k,
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

// ToNode converts a Jsonnet component to a Jsonnet node.
func (j *Jsonnet) ToNode(envName string) (string, ast.Node, error) {
	n, err := jsonnet.ImportNodeFromFs(j.source, j.app.Fs())
	if err != nil {
		return "", nil, err
	}

	return j.Name(true), n, nil
}

func (j *Jsonnet) readParams(envName string) (string, error) {
	if envName == "" {
		return j.readModuleParams()
	}

	return envParams(j.app, j.module, envName)
}

func (j *Jsonnet) readModuleParams() (string, error) {
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

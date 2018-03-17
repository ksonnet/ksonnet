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

package actions

import (
	"strings"

	"github.com/ksonnet/ksonnet/component"
	"github.com/ksonnet/ksonnet/env"
	"github.com/ksonnet/ksonnet/metadata/app"
	mp "github.com/ksonnet/ksonnet/metadata/params"
	"github.com/ksonnet/ksonnet/pkg/params"
	"github.com/pkg/errors"
)

// RunParamSet sets a parameter for a component.
func RunParamSet(ksApp app.App, componentName, path, value string, opts ...ParamSetOpt) error {
	ps, err := NewParamSet(ksApp, componentName, path, value, opts...)
	if err != nil {
		return err
	}

	return ps.Run()
}

// ParamSetOpt is an option for configuring ParamSet.
type ParamSetOpt func(*ParamSet)

// ParamSetGlobal sets if the param is global.
func ParamSetGlobal(isGlobal bool) ParamSetOpt {
	return func(ps *ParamSet) {
		ps.global = isGlobal
	}
}

// ParamSetEnv sets the env name for a param.
func ParamSetEnv(envName string) ParamSetOpt {
	return func(ps *ParamSet) {
		ps.envName = envName
	}
}

// ParamSetWithIndex sets the index for the set option.
func ParamSetWithIndex(index int) ParamSetOpt {
	return func(ParamSet *ParamSet) {
		ParamSet.index = index
	}
}

// ParamSet sets a parameter for a component.
type ParamSet struct {
	app      app.App
	name     string
	rawPath  string
	rawValue string
	index    int
	global   bool
	envName  string

	// TODO: remove once ksonnet has more robust env param handling.
	setEnv func(ksApp app.App, envName, name, pName, value string) error

	cm component.Manager
}

// NewParamSet creates an instance of ParamSet.
func NewParamSet(ksApp app.App, name, path, value string, opts ...ParamSetOpt) (*ParamSet, error) {
	ps := &ParamSet{
		app:      ksApp,
		name:     name,
		rawPath:  path,
		rawValue: value,
		cm:       component.DefaultManager,
		setEnv:   setEnv,
	}

	for _, opt := range opts {
		opt(ps)
	}

	if ps.envName != "" && ps.global {
		return nil, errors.New("unable to set global param for environments")
	}

	return ps, nil
}

// Run runs the action.
func (ps *ParamSet) Run() error {

	value, err := params.DecodeValue(ps.rawValue)
	if err != nil {
		return errors.Wrap(err, "value is invalid")
	}

	if ps.envName != "" {
		return ps.setEnv(ps.app, ps.envName, ps.name, ps.rawPath, ps.rawValue)
	}

	path := strings.Split(ps.rawPath, ".")

	if ps.global {
		return ps.setGlobal(path, value)
	}

	return ps.setLocal(path, value)
}

func (ps *ParamSet) setGlobal(path []string, value interface{}) error {
	ns, err := ps.cm.Namespace(ps.app, ps.name)
	if err != nil {
		return errors.Wrap(err, "retrieve namespace")
	}

	if err := ns.SetParam(path, value); err != nil {
		return errors.Wrap(err, "set global param")
	}

	return nil
}

func (ps *ParamSet) setLocal(path []string, value interface{}) error {
	_, c, err := ps.cm.ResolvePath(ps.app, ps.name)
	if err != nil {
		return errors.Wrap(err, "could not find component")
	}

	options := component.ParamOptions{
		Index: ps.index,
	}
	if err := c.SetParam(path, value, options); err != nil {
		return errors.Wrap(err, "set param")
	}

	return nil
}

func setEnv(ksApp app.App, envName, name, pName, value string) error {
	spc := env.SetParamsConfig{
		App: ksApp,
	}

	p := mp.Params{
		pName: value,
	}

	return env.SetParams(envName, name, p, spc)
}

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
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/env"
	"github.com/pkg/errors"
)

// EnvSetNamespace is an option for setting a new namespace name.
func EnvSetNamespace(nsName string) EnvSetOpt {
	return func(es *EnvSet) {
		es.newNsName = nsName
	}
}

// EnvSetName is an option for setting a new name.
func EnvSetName(name string) EnvSetOpt {
	return func(es *EnvSet) {
		es.newName = name
	}
}

// EnvSetOpt is an option for configuring EnvSet.
type EnvSetOpt func(*EnvSet)

// RunEnvSet runs `env set`
// func RunEnvSet(ksApp app.App, envName string, opts ...EnvSetOpt) error {
func RunEnvSet(m map[string]interface{}) error {
	et, err := NewEnvSet(m)
	if err != nil {
		return err
	}

	return et.Run()
}

// func types for renaming and updating environments
type envRenameFn func(a app.App, from, to string, override bool) error
type saveFn func(a app.App, envName, k8sAPISpec string, spec *app.EnvironmentConfig, override bool) error

// EnvSet sets targets for an environment.
type EnvSet struct {
	app        app.App
	envName    string
	newName    string
	newNsName  string
	newServer  string
	newAPISpec string

	envRenameFn envRenameFn
	saveFn      saveFn
}

// NewEnvSet creates an instance of EnvSet.
func NewEnvSet(m map[string]interface{}) (*EnvSet, error) {
	ol := newOptionLoader(m)

	es := &EnvSet{
		app:        ol.LoadApp(),
		envName:    ol.LoadString(OptionEnvName),
		newName:    ol.LoadOptionalString(OptionNewEnvName),
		newNsName:  ol.LoadOptionalString(OptionNamespace),
		newServer:  ol.LoadOptionalString(OptionServer),
		newAPISpec: ol.LoadOptionalString(OptionSpecFlag),

		envRenameFn: env.Rename,
		saveFn:      save,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return es, nil
}

// Run assigns targets to an environment.
func (es *EnvSet) Run() error {
	env, err := es.app.Environment(es.envName)
	if err != nil {
		return err
	}

	if err := es.updateName(env.IsOverride()); err != nil {
		return err
	}

	if err := es.updateEnvConfig(*env, es.newNsName, es.newServer, es.newAPISpec, env.IsOverride()); err != nil {
		return err
	}

	return nil
}

func (es *EnvSet) updateName(isOverride bool) error {
	if es.newName != "" {
		if err := es.envRenameFn(es.app, es.envName, es.newName, isOverride); err != nil {
			return err
		}

		es.envName = es.newName
	}

	return nil
}

// updateEnvConfig merges the provided environment config with optional override settings and the  creates and saves a new environment config based on the provided
// base configuration. It will be merged with the provided configuration settings.
// If isOverride is specified, Libraries will be filtered out of the merged configuration, as those should always be
// managed in the primary application config.
func (es *EnvSet) updateEnvConfig(env app.EnvironmentConfig, namespace, server, k8sAPISpec string, isOverride bool) error {
	if env.Name == "" {
		return errors.Errorf("empty environment name")
	}
	if namespace == "" && server == "" && k8sAPISpec == "" {
		// Nothing to update
		return nil
	}

	newEnv := env

	var destination *app.EnvironmentDestinationSpec
	if env.Destination != nil {
		var destCopy app.EnvironmentDestinationSpec
		destCopy = *env.Destination
		destination = &destCopy // also a copy
	}
	if destination == nil && (server != "" || namespace != "") {
		destination = &app.EnvironmentDestinationSpec{}
	}
	if server != "" {
		destination.Server = server
	}
	if namespace != "" {
		destination.Namespace = namespace
	}

	newEnv.Destination = destination

	// isOverride will be set by app.AddEnvironment
	if isOverride {
		// Libraries will always derive from the primary app.yaml
		newEnv.Libraries = nil
	}

	return es.saveFn(es.app, newEnv.Name, k8sAPISpec, &newEnv, isOverride)
}

func save(a app.App, envName, k8sAPISpec string, env *app.EnvironmentConfig, override bool) error {
	return a.AddEnvironment(env, k8sAPISpec, override)
}

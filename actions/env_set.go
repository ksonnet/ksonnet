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
	"github.com/ksonnet/ksonnet/env"
	"github.com/ksonnet/ksonnet/metadata/app"
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
func RunEnvSet(ksApp app.App, envName string, opts ...EnvSetOpt) error {
	et, err := NewEnvSet(ksApp, envName, opts...)
	if err != nil {
		return err
	}

	return et.Run()
}

// EnvSet sets targets for an environment.
type EnvSet struct {
	app       app.App
	em        env.Manager
	envName   string
	newName   string
	newNsName string
}

// NewEnvSet creates an instance of EnvSet.
func NewEnvSet(ksApp app.App, envName string, opts ...EnvSetOpt) (*EnvSet, error) {
	es := &EnvSet{
		app:     ksApp,
		em:      env.DefaultManager,
		envName: envName,
	}

	for _, opt := range opts {
		opt(es)
	}

	return es, nil
}

// Run assigns targets to an environment.
func (es *EnvSet) Run() error {
	if err := es.updateName(); err != nil {
		return err
	}

	return es.updateNamespace()
}

func (es *EnvSet) updateName() error {
	if es.newName != "" {
		config := env.RenameConfig{
			App: es.app,
		}
		if err := es.em.Rename(es.envName, es.newName, config); err != nil {
			return err
		}

		es.envName = es.newName
	}

	return nil
}

func (es *EnvSet) updateNamespace() error {
	if es.newNsName != "" {
		spec, err := es.app.Environment(es.envName)
		if err != nil {
			return err
		}

		spec.Destination.Namespace = es.newNsName
		return es.app.AddEnvironment(es.envName, "", spec)
	}

	return nil
}

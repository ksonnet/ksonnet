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
	"github.com/ksonnet/ksonnet/component"
	"github.com/ksonnet/ksonnet/metadata/app"
)

// RunEnvTargets runs `env targets`
func RunEnvTargets(ksApp app.App, envName string, nsNames []string) error {
	et, err := NewEnvTargets(ksApp, envName, nsNames)
	if err != nil {
		return err
	}

	return et.Run()
}

// EnvTargets sets targets for an environment.
type EnvTargets struct {
	app     app.App
	envName string
	nsNames []string
	cm      component.Manager
}

// NewEnvTargets creates an instance of EnvTargets.
func NewEnvTargets(ksApp app.App, envName string, nsNames []string) (*EnvTargets, error) {
	et := &EnvTargets{
		app:     ksApp,
		envName: envName,
		nsNames: nsNames,
		cm:      component.DefaultManager,
	}

	return et, nil
}

// Run assigns targets to an environment.
func (et *EnvTargets) Run() error {
	_, err := et.app.Environment(et.envName)
	if err != nil {
		return err
	}

	for _, nsName := range et.nsNames {
		_, err := et.cm.Namespace(et.app, nsName)
		if err != nil {
			return err
		}
	}

	return et.app.UpdateTargets(et.envName, et.nsNames)
}

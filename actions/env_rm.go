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

// RunEnvRm runs `env rm`
func RunEnvRm(ksApp app.App, envName string, isOverride bool) error {
	ea, err := NewEnvRm(ksApp, envName, isOverride)
	if err != nil {
		return err
	}

	return ea.Run()
}

type envDeleteFn func(a app.App, name string, override bool) error

// EnvRm sets targets for an environment.
type EnvRm struct {
	app        app.App
	envName    string
	isOverride bool

	envDeleteFn envDeleteFn
}

// NewEnvRm creates an instance of EnvRm.
func NewEnvRm(ksApp app.App, envName string, isOverride bool) (*EnvRm, error) {
	ea := &EnvRm{
		app:        ksApp,
		envName:    envName,
		isOverride: isOverride,

		envDeleteFn: env.Delete,
	}

	return ea, nil
}

// Run assigns targets to an environment.
func (er *EnvRm) Run() error {
	return er.envDeleteFn(
		er.app,
		er.envName,
		er.isOverride,
	)
}

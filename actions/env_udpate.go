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
	"fmt"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/lib"
)

// RunEnvUpdate runs `env update`.
func RunEnvUpdate(m map[string]interface{}) error {
	a, err := newEnvUpdate(m)
	if err != nil {
		return err
	}

	return a.run()
}

// EnvUpdate updates ksonnet lib for an environment.
type EnvUpdate struct {
	app     app.App
	envName string

	genLibFn func(app.App, string, string) error
}

// RunEnvUpdate runs `env update`
func newEnvUpdate(m map[string]interface{}) (*EnvUpdate, error) {
	ol := newOptionLoader(m)

	eu := &EnvUpdate{
		app:     ol.LoadApp(),
		envName: ol.LoadString(OptionEnvName),

		genLibFn: genLib,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return eu, nil
}

func (eu *EnvUpdate) run() error {
	envSpec, err := eu.app.Environment(eu.envName)
	if err != nil {
		return err
	}

	k8sSpecFlag := fmt.Sprintf("version:%s", envSpec.KubernetesVersion)

	libPath, err := eu.app.LibPath(eu.envName)
	if err != nil {
		return err
	}

	return eu.genLibFn(eu.app, k8sSpecFlag, libPath)

}

func genLib(a app.App, k8sSpecFlag, libPath string) error {
	libManager, err := lib.NewManager(k8sSpecFlag, a.Fs(), libPath)
	if err != nil {
		return err
	}

	if err = a.Fs().RemoveAll(libPath); err != nil {
		return err
	}

	return libManager.GenerateLibData(false)
}

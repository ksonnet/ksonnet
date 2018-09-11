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
	"io"
	"os"
	"sort"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/util/table"
	"github.com/pkg/errors"
)

// RunEnvList runs `env list`
func RunEnvList(m map[string]interface{}) error {
	nl, err := NewEnvList(m)
	if err != nil {
		return err
	}

	return nl.Run()
}

// EnvList lists available namespaces. To initialize EnvList,
// use the `NewEnvList` constructor.
type EnvList struct {
	envListFn       func() (app.EnvironmentConfigs, error)
	envIsOverrideFn func(name string) bool
	outputType      string
	out             io.Writer
}

// NewEnvList creates an instance of EnvList
func NewEnvList(m map[string]interface{}) (*EnvList, error) {
	ol := newOptionLoader(m)

	a := ol.LoadApp()
	outputType := ol.LoadOptionalString(OptionOutput)

	if ol.err != nil {
		return nil, ol.err
	}

	el := &EnvList{
		outputType:      outputType,
		envListFn:       a.Environments,
		envIsOverrideFn: a.IsEnvOverride,
		out:             os.Stdout,
	}

	return el, nil
}

// Run runs the env list action.
func (el *EnvList) Run() error {
	environments, err := el.envListFn()
	if err != nil {
		return err
	}

	t := table.New("envList", el.out)
	t.SetHeader([]string{"name", "override", "kubernetes-version", "namespace", "server"})

	f, err := table.DetectFormat(el.outputType)
	if err != nil {
		return errors.Wrap(err, "detecting output format")
	}
	t.SetFormat(f)

	var rows [][]string

	for name, env := range environments {
		override := ""
		if el.envIsOverrideFn(name) {
			override = "*"
		}

		rows = append(rows, []string{
			name,
			override,
			env.KubernetesVersion,
			env.Destination.Namespace,
			env.Destination.Server,
		})
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	t.AppendBulk(rows)

	return t.Render()

}

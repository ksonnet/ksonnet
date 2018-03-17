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

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/pkg/registry"
	"github.com/ksonnet/ksonnet/pkg/util/table"
)

// RunRegistryList runs `env list`
func RunRegistryList(ksApp app.App) error {
	rl, err := NewRegistryList(ksApp)
	if err != nil {
		return err
	}

	return rl.Run()
}

// RegistryList lists available registries
type RegistryList struct {
	app app.App
	rm  registry.Manager
	out io.Writer
}

// NewRegistryList creates an instance of RegistryList
func NewRegistryList(ksApp app.App) (*RegistryList, error) {
	rl := &RegistryList{
		app: ksApp,
		rm:  registry.DefaultManager,
		out: os.Stdout,
	}

	return rl, nil
}

// Run runs the env list action.
func (rl *RegistryList) Run() error {
	registries, err := rl.rm.Registries(rl.app)
	if err != nil {
		return err
	}

	t := table.New(rl.out)
	t.SetHeader([]string{"name", "protocol", "uri"})

	var rows [][]string

	for _, r := range registries {
		rows = append(rows, []string{
			r.Name(),
			r.Protocol(),
			r.URI(),
		})
	}

	t.AppendBulk(rows)

	return t.Render()
}

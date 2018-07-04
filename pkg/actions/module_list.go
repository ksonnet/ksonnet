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
	"github.com/ksonnet/ksonnet/pkg/component"
	"github.com/ksonnet/ksonnet/pkg/util/table"
	"github.com/pkg/errors"
)

// RunModuleList runs `module list`
func RunModuleList(m map[string]interface{}) error {
	nl, err := NewModuleList(m)
	if err != nil {
		return err
	}

	return nl.Run()
}

// ModuleList lists modules.
type ModuleList struct {
	app        app.App
	envName    string
	outputType string
	out        io.Writer
	cm         component.Manager
}

// NewModuleList creates an instance of ModuleList.
func NewModuleList(m map[string]interface{}) (*ModuleList, error) {
	ol := newOptionLoader(m)

	nl := &ModuleList{
		app:        ol.LoadApp(),
		envName:    ol.LoadString(OptionEnvName),
		outputType: ol.LoadOptionalString(OptionOutput),

		out: os.Stdout,
		cm:  component.DefaultManager,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return nl, nil
}

// Run lists modules.
func (nl *ModuleList) Run() error {
	modules, err := nl.cm.Modules(nl.app, nl.envName)
	if err != nil {
		return err
	}

	t := table.New("moduleList", nl.out)
	t.SetHeader([]string{"module"})

	f, err := table.DetectFormat(nl.outputType)
	if err != nil {
		return errors.Wrap(err, "detecting output format")
	}
	t.SetFormat(f)

	names := make([]string, len(modules))
	for i := range modules {
		names[i] = modules[i].Name()
	}

	sort.Strings(names)

	for _, name := range names {
		t.Append([]string{name})
	}

	return t.Render()
}

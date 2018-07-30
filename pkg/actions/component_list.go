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

// RunComponentList runs `component list`
func RunComponentList(m map[string]interface{}) error {
	cl, err := NewComponentList(m)
	if err != nil {
		return err
	}

	return cl.Run()
}

// ComponentList create a list of components in a module.
type ComponentList struct {
	app    app.App
	module string
	output string
	cm     component.Manager
	out    io.Writer
}

// NewComponentList creates an instance of ComponentList.
func NewComponentList(m map[string]interface{}) (*ComponentList, error) {
	ol := newOptionLoader(m)

	cl := &ComponentList{
		app:    ol.LoadApp(),
		module: ol.LoadString(OptionModule),
		output: ol.LoadString(OptionOutput),

		cm:  component.DefaultManager,
		out: os.Stdout,
	}

	if ol.err != nil {
		return nil, ol.err
	}

	return cl, nil
}

// Run runs the ComponentList action.
func (cl *ComponentList) Run() error {
	components, err := cl.cm.Components(cl.app, cl.module)
	if err != nil {
		return err
	}

	return cl.listComponents(components)
}

func (cl *ComponentList) listComponents(components []component.Component) error {
	var rows [][]string
	for _, c := range components {
		summary, err := c.Summarize()
		if err != nil {
			return err
		}

		row := []string{
			summary.ComponentName,
			summary.Type,
			summary.APIVersion,
			summary.Kind,
			summary.Name,
		}

		rows = append(rows, row)
	}

	sort.Slice(rows, func(i, j int) bool {
		return rows[i][0] < rows[j][0]
	})

	t := table.New("componentList", cl.out)
	f, err := table.DetectFormat(cl.output)
	if err != nil {
		return errors.Wrap(err, "detecting output format")
	}

	t.SetFormat(f)
	t.SetHeader([]string{"component", "type", "apiversion", "kind", "name"})
	t.AppendBulk(rows)
	return t.Render()
}

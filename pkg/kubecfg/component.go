// Copyright 2017 The kubecfg authors
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

package kubecfg

import (
	"io"
	"sort"

	"github.com/ksonnet/ksonnet/component"
	"github.com/ksonnet/ksonnet/pkg/util/table"
)

const (
	componentNameHeader = "COMPONENT"
)

// ComponentListCmd stores the information necessary to display components.
type ComponentListCmd struct {
}

// NewComponentListCmd acts as a constructor for ComponentListCmd.
func NewComponentListCmd() *ComponentListCmd {
	return &ComponentListCmd{}
}

// Run executes the displaying of components.
func (c *ComponentListCmd) Run(out io.Writer) error {
	manager, err := manager()
	if err != nil {
		return err
	}

	components, err := manager.GetAllComponents()
	if err != nil {
		return err
	}

	return printComponents(out, components)
}

// ComponentRmCmd stores the information necessary to remove a component from
// the ksonnet application.
type ComponentRmCmd struct {
	component string
}

// NewComponentRmCmd acts as a constructor for ComponentRmCmd.
func NewComponentRmCmd(component string) *ComponentRmCmd {
	return &ComponentRmCmd{component: component}
}

// Run executes the removing of the component.
func (c *ComponentRmCmd) Run() error {
	manager, err := manager()
	if err != nil {
		return err
	}

	return manager.DeleteComponent(c.component)
}

func printComponents(out io.Writer, components []component.Component) error {
	t := table.New(out)
	t.SetHeader([]string{componentNameHeader})

	sort.SliceStable(components, func(i, j int) bool {
		return components[i].Name(true) < components[j].Name(true)
	})

	for _, component := range components {
		t.Append([]string{component.Name(false)})
	}

	return t.Render()
}

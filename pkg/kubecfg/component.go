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
	"fmt"
	"io"
	"sort"
	"strings"

	str "github.com/ksonnet/ksonnet/strings"
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

	_, err = printComponents(out, components)
	return err
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

func printComponents(out io.Writer, components []string) (string, error) {
	rows := [][]string{
		[]string{componentNameHeader},
		[]string{strings.Repeat("=", len(componentNameHeader))},
	}

	sort.Strings(components)
	for _, component := range components {
		rows = append(rows, []string{component})
	}

	formatted, err := str.PadRows(rows)
	if err != nil {
		return "", err
	}
	_, err = fmt.Fprint(out, formatted)
	return formatted, err
}

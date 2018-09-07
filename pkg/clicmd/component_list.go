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

package clicmd

import (
	"fmt"

	"github.com/ksonnet/ksonnet/pkg/actions"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	vComponentListNamespace = "component-list-namespace"
	vComponentListOutput    = "component-list-output"

	componentListLong = `
The ` + "`list`" + ` command displays all known components.

### Syntax
`
	componentListExample = `
# List all components
ks component list`
)

func newComponentListCmd() *cobra.Command {
	componentListCmd := &cobra.Command{
		Use:     "list",
		Short:   "List known components",
		Long:    componentListLong,
		Example: componentListExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("'component list' takes zero arguments")
			}

			m := map[string]interface{}{
				actions.OptionModule: viper.GetString(vComponentListNamespace),
				actions.OptionOutput: viper.GetString(vComponentListOutput),
			}
			addGlobalOptions(m)

			return runAction(actionComponentList, m)
		},
	}

	addCmdOutput(componentListCmd, vComponentListOutput)
	componentListCmd.Flags().String(flagModule, "", "Component module")
	viper.BindPFlag(vComponentListNamespace, componentListCmd.Flags().Lookup(flagModule))

	return componentListCmd
}

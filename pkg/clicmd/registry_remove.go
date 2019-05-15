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
)

var (
	registryRemoveLong = `
The ` + "`remove`" + ` command allows added registries to be removed from your ksonnet app.

### Related Commands

* ` + "`ks registry add` " + `— ` + regShortDesc["add"] + `

### Syntax
`
	registryRemoveExample = `# Remove a registry with the name 'databases'
ks registry remove databases`
)

func newRegistryRemoveCmd() *cobra.Command {
	registryRemoveCmd := &cobra.Command{
		Use:     "remove <registry-name>",
		Short:   regShortDesc["remove"],
		Long:    registryRemoveLong,
		Example: registryRemoveExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("Command 'registry remove' takes one argument, the name of the registry to remove")
			}

			m := map[string]interface{}{
				actions.OptionName: args[0],
			}
			addGlobalOptions(m)

			return runAction(actionRegistryRemove, m)
		},
	}

	return registryRemoveCmd
}

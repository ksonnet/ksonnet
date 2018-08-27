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

	"github.com/spf13/viper"

	"github.com/ksonnet/ksonnet/pkg/actions"
	"github.com/spf13/cobra"
)

var (
	vPkgRemoveEnv = "pkg-remove-env"

	pkgRemoveLong = `
The ` + "`remove`" + ` command removes a reference to a ksonnet library.  The reference can either be
global or scoped to an environment. If the last reference to a library version is removed, the cached
files will be removed as well.

### Syntax
`
	pkgRemoveExample = `
# Remove an nginx dependency
ks pkg remove incubator/nginx

# Remove an nginx dependency from the stage environment
ks pkg remove incubator/nginx --env stage
`
)

func newPkgRemoveCmd() *cobra.Command {

	pkgRemoveCmd := &cobra.Command{
		Use:     "remove <registry>/<library>",
		Short:   pkgShortDesc["remove"],
		Long:    pkgRemoveLong,
		Example: pkgRemoveExample,
		Aliases: []string{"get"},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("Command requires a single argument of the form <registry>/<library>@<version>\n\n%s", cmd.UsageString())
			}

			m := map[string]interface{}{
				actions.OptionPkgName: args[0],
				actions.OptionEnvName: viper.GetString(vPkgRemoveEnv),
			}
			addGlobalOptions(m)

			return runAction(actionPkgRemove, m)
		},
	}

	pkgRemoveCmd.Flags().String(flagEnv, "", "Environment to remove package from (optional)")
	viper.BindPFlag(vPkgRemoveEnv, pkgRemoveCmd.Flags().Lookup(flagEnv))

	return pkgRemoveCmd
}

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
	envUpdateLong = `
The ` + "`update`" + ` command updates packages for an environment.

### Related Commands

* ` + "`ks env list` " + `— ` + envShortDesc["list"] + `
* ` + "`ks env add` " + `— ` + envShortDesc["add"] + `
* ` + "`ks env set` " + `— ` + envShortDesc["set"] + `
* ` + "`ks delete` " + `— ` + `Delete all the app components running in an environment (cluster)` + `

### Syntax
`
	envUpdateExample = `
# Update the environment 'us-west/staging' packages.
ks env update us-west/staging`
)

func newEnvUpdateCmd() *cobra.Command {
	envUpdateCmd := &cobra.Command{
		Use:     "update <env-name>",
		Short:   envShortDesc["update"],
		Long:    envUpdateLong,
		Example: envUpdateExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("'env update' takes a single argument, that is the name of the environment")
			}

			m := map[string]interface{}{
				actions.OptionEnvName: args[0],
			}
			addGlobalOptions(m)

			return runAction(actionEnvUpdate, m)
		},
	}

	return envUpdateCmd

}

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

package cmd

import (
	"fmt"

	"github.com/ksonnet/ksonnet/actions"
	"github.com/spf13/cobra"
)

var envUpdateCmd = &cobra.Command{
	Use:   "update <env-name>",
	Short: envShortDesc["update"],
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("'env update' takes a single argument, that is the name of the environment")
		}

		m := map[string]interface{}{
			actions.OptionApp:     ka,
			actions.OptionEnvName: args[0],
		}

		return runAction(actionEnvUpdate, m)
	},
	Long: `
The ` + "`update`" + ` command updates libraries for an environment.

### Related Commands

* ` + "`ks env list` " + `— ` + protoShortDesc["list"] + `
* ` + "`ks env add` " + `— ` + protoShortDesc["add"] + `
* ` + "`ks env set` " + `— ` + protoShortDesc["set"] + `
* ` + "`ks delete` " + `— ` + `Delete all the app components running in an environment (cluster)` + `

### Syntax
`,
	Example: `
# Update the environment 'us-west/staging' libs.
ks env update us-west/staging`,
}

func init() {
	envCmd.AddCommand(envUpdateCmd)
}

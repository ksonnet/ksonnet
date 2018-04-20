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

	"github.com/ksonnet/ksonnet/actions"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	vEnvCurrentSet   = "env-current-set"
	vEnvCurrentUnset = "env-current-unset"
)

var envCurrentCmd = &cobra.Command{
	Use:   "current [--set <name> | --unset]",
	Short: envShortDesc["current"],
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("'env current' takes no arguments")
		}

		m := map[string]interface{}{
			actions.OptionApp:     ka,
			actions.OptionEnvName: viper.GetString(vEnvCurrentSet),
			actions.OptionUnset:   viper.GetBool(vEnvCurrentUnset),
		}

		return runAction(actionEnvCurrent, m)
	},
	Long: `
The ` + "`current`" + ` command lets you set the current ksonnet environment.

### Related Commands

* ` + "`ks env list` " + `— ` + envShortDesc["list"] + `

### Syntax
`,
	Example: `#Update the current environment to 'us-west/staging'
ks env current --set us-west/staging

#Retrieve the current environment
ks env current

#Unset the current environment
ks env current --unset`,
}

func init() {
	envCmd.AddCommand(envCurrentCmd)

	envCurrentCmd.Flags().String(flagSet, "",
		"Environment to set as current")
	viper.BindPFlag(vEnvCurrentSet, envCurrentCmd.Flags().Lookup(flagSet))

	envCurrentCmd.Flags().Bool(flagUnset, false,
		"Unset current environment")
	viper.BindPFlag(vEnvCurrentUnset, envCurrentCmd.Flags().Lookup(flagUnset))
}

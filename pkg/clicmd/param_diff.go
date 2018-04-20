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

const (
	vParamDiffComponent = "param-diff-component"
)

var paramDiffCmd = &cobra.Command{
	Use:   "diff <env1> <env2> [--component <component-name>]",
	Short: paramShortDesc["diff"],
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 2 {
			return fmt.Errorf("'param diff' takes exactly two arguments: the respective names of the environments being diffed")
		}

		m := map[string]interface{}{
			actions.OptionApp:           ka,
			actions.OptionEnvName1:      args[0],
			actions.OptionEnvName2:      args[1],
			actions.OptionComponentName: viper.GetString(vParamDiffComponent),
		}

		return runAction(actionParamDiff, m)
	},
	Long: `
The ` + "`diff`" + ` command pretty prints differences between the component parameters
of two environments.

By default, the diff is performed for all components. Diff-ing for a single component
is supported via a component flag.

### Related Commands

* ` + "`ks param set` " + `— ` + paramShortDesc["set"] + `
* ` + "`ks apply` " + `— ` + applyShortDesc + `

### Syntax
`,
	Example: `
# Diff between all component parameters for environments 'dev' and 'prod'
ks param diff dev prod

# Diff only between the parameters for the 'guestbook' component for environments
# 'dev' and 'prod'
ks param diff dev prod --component=guestbook`,
}

func init() {
	paramCmd.AddCommand(paramDiffCmd)

	paramListCmd.PersistentFlags().String(flagEnv, "", "Specify environment to list parameters for")
	paramListCmd.Flags().String(flagModule, "", "Specify module to list parameters for")

	paramDiffCmd.Flags().String(flagComponent, "", "Specify the component to diff against")
	viper.BindPFlag(vParamDiffComponent, paramDiffCmd.Flags().Lookup(flagComponent))
}

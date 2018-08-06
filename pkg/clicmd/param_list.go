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
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	vParamListOutput         = "param-list-output"
	vParamListWithoutModules = "param-without-modules"
)

var (
	paramListLong = `
The ` + "`list`" + ` command displays all known component parameters or environment parameters.

If a component is specified, this command displays all of its specific parameters.
If a component is NOT specified, parameters for **all** components are listed.
Furthermore, parameters can be listed on a per-environment basis.

### Related Commands

* ` + "`ks param set` " + `— ` + paramShortDesc["set"] + `

### Syntax
`
	paramListExample = `
# List all component parameters
ks param list

# List all parameters for the component "guestbook"
ks param list guestbook

# List all parameters for the environment "dev"
ks param list --env=dev

# List all parameters for the component "guestbook" in the environment "dev"
ks param list guestbook --env=dev`
)

func newParamListCmd(a app.App) *cobra.Command {
	paramListCmd := &cobra.Command{
		Use:     "list [<component-name>] [--env <env-name>]",
		Short:   paramShortDesc["list"],
		Long:    paramListLong,
		Example: paramListExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			if len(args) > 1 {
				return fmt.Errorf("'param list' takes at most one argument, that is the name of the component")
			}

			component := ""
			if len(args) == 1 {
				component = args[0]
			}

			env, err := flags.GetString(flagEnv)
			if err != nil {
				return err
			}

			module, err := flags.GetString(flagModule)
			if err != nil {
				return err
			}

			m := map[string]interface{}{
				actions.OptionApp:            a,
				actions.OptionComponentName:  component,
				actions.OptionEnvName:        env,
				actions.OptionModule:         module,
				actions.OptionOutput:         viper.GetString(vParamListOutput),
				actions.OptionWithoutModules: viper.GetBool(vParamListWithoutModules),
			}

			return runAction(actionParamList, m)
		},
	}

	addCmdOutput(paramListCmd, vParamListOutput)
	paramListCmd.PersistentFlags().String(flagEnv, "", "Specify environment to list parameters for")
	paramListCmd.Flags().String(flagModule, "", "Specify module to list parameters for")

	paramListCmd.Flags().Bool(flagWithoutModules, false, "Exclude module defaults")
	viper.BindPFlag(vParamListWithoutModules, paramListCmd.Flags().Lookup(flagWithoutModules))

	return paramListCmd

}

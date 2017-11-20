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

package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/ksonnet/ksonnet/pkg/kubecfg"
)

const (
	flagParamEnv       = "env"
	flagParamComponent = "component"
)

func init() {
	RootCmd.AddCommand(paramCmd)

	paramCmd.AddCommand(paramSetCmd)
	paramCmd.AddCommand(paramListCmd)
	paramCmd.AddCommand(paramDiffCmd)

	paramSetCmd.PersistentFlags().String(flagParamEnv, "", "Name of the environment to set parameters for")
	paramListCmd.PersistentFlags().String(flagParamEnv, "", "Name of the environment to list parameters for")
	paramDiffCmd.PersistentFlags().String(flagParamComponent, "", "Base name of the component file to diff against")
}

var paramCmd = &cobra.Command{
	Use:   "param",
	Short: `Manage parameters for ksonnet components`,
	Long: `Parameters are predetermined fields that can be used to customize ksonnet
components. Potential examples include replica count, component name, and deployment
image.

Parameters can also be defined on a per-environment basis. For example, this
allows a "development" environment to only run a single replica for component X,
whereas a production environment might run multiple replicas for component X, in
order to meet heavier production load demands.

Environments can be thought of as "named cluster contexts". For more information,
run ` + "`ks env --help`" + ` or see the concept documentation.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Command 'param' requires a subcommand\n\n%s", cmd.UsageString())
	},
}

var paramSetCmd = &cobra.Command{
	Use:   "set <component-name> <param-key> <param-value>",
	Short: "Set component or environment parameters such as replica count or name",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		if len(args) != 3 {
			return fmt.Errorf("'param set' takes exactly three arguments: (1) the name of the component, in addition to (2) the key and (3) value of the parameter, respectively")
		}

		component := args[0]
		param := args[1]
		value := args[2]

		env, err := flags.GetString(flagParamEnv)
		if err != nil {
			return err
		}

		c := kubecfg.NewParamSetCmd(component, env, param, value)

		return c.Run()
	},
	Long: `Set component or environment parameters such as replica count or name.

Parameters are set individually, one at a time. If you need to customize multiple
fields at once, we suggest that you modify your ksonnet application's
` + " `components/params.libsonnet` " + `file directly. Likewise, for greater customization
of environment parameters, we suggest modifying the
` + " `environments/:name/params.libsonnet` " + `file.

Additional information about parameters is available in the concept documentation.
`,
	Example: `# Update the replica count of the 'guestbook' component to 4, across
# all environments
ks param set guestbook replicas 4

# Update the replica count of the 'guestbook' component to 2, but only for the environment
# 'dev'
ks param set guestbook replicas 2 --env=dev`,
}

var paramListCmd = &cobra.Command{
	Use:   "list <component-name>",
	Short: "List all parameters for a component(s)",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		if len(args) > 1 {
			return fmt.Errorf("'param list' takes at most one argument: the component name")
		}

		component := ""
		if len(args) == 1 {
			component = args[0]
		}

		env, err := flags.GetString(flagParamEnv)
		if err != nil {
			return err
		}

		c := kubecfg.NewParamListCmd(component, env)

		return c.Run(cmd.OutOrStdout())
	},
	Long: `List all component parameters or environment parameters.

This command displays all parameters for the specified component. If a
component is not specified, parameters for *all* components are listed.

Furthermore, parameters can be listed on a per-environment basis.
`,
	Example: `# List all component parameters
ks param list

# List all parameters for the component "guestbook"
ks param list guestbook

# List all parameters for the environment "dev"
ks param list --env=dev

# List all parameters for the component "guestbook" in the environment "dev"
ks param list guestbook --env=dev`,
}

var paramDiffCmd = &cobra.Command{
	Use:   "diff <env1> <env2>",
	Short: "Display differences between the component parameters of two environments",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		if len(args) != 2 {
			return fmt.Errorf("'param diff' takes exactly two arguments: the respective names of the environments being diffed")
		}

		env1 := args[0]
		env2 := args[1]

		component, err := flags.GetString(flagParamComponent)
		if err != nil {
			return err
		}

		c := kubecfg.NewParamDiffCmd(env1, env2, component)

		return c.Run(cmd.OutOrStdout())
	},
	Long: `Pretty print differences between the component parameters of two environments.

By default, the diff is performed for all components. Diff-ing for a single component
is supported via a component flag.
`,
	Example: `# Diff between all component parameters for environments 'dev' and 'prod'
ks param diff dev prod

# Diff only between the parameters for the 'guestbook' component for environments 'dev' and 'prod'
ks param diff dev prod --component=guestbook`,
}

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
	"strings"

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

	paramSetCmd.PersistentFlags().String(flagParamEnv, "", "Specify environment to set parameters for")
	paramListCmd.PersistentFlags().String(flagParamEnv, "", "Specify environment to list parameters for")
	paramDiffCmd.PersistentFlags().String(flagParamComponent, "", "Specify the component to diff against")
}

var paramCmd = &cobra.Command{
	Use:   "param",
	Short: `Manage ksonnet component parameters`,
	Long: `Parameters are the customizable fields defining ksonnet components. For
example, replica count, component name, or deployment image.

Parameters are also able to be defined separately across environments. Meaning,
this supports features to allow a "development" environment to only run a
single replication instance for it's components, whereas allowing a "production"
environment to run more replication instances to meet heavier production load
demands.

Environments are ksonnet "named clusters". For more information on environments,
run:

    ks env --help
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("%s is not a valid subcommand\n\n%s", strings.Join(args, " "), cmd.UsageString())
		}
		return fmt.Errorf("Command 'param' requires a subcommand\n\n%s", cmd.UsageString())
	},
}

var paramSetCmd = &cobra.Command{
	Use:   "set <component-name> <param-key> <param-value>",
	Short: "Set component or environment parameters such as replica count or name",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		if len(args) != 3 {
			return fmt.Errorf("'param set' takes exactly three arguments, the name of the component, and the key and value of the parameter, respectively")
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

Parameters are set individually, one at a time. If you require customization of
more fields, we suggest that you modify your ksonnet project's
` + " `components/params.libsonnet` " + `file directly. Likewise, for greater customization
of environment parameters, we suggest modifying the
` + " `environments/:name/params.libsonnet` " + `file.
`,
	Example: `# Updates the replica count of the 'guestbook' component to 4.
ks param set guestbook replicas 4

# Updates the replica count of the 'guestbook' component to 2 for the environment
# 'dev'
ks param set guestbook replicas 2 --env=dev`,
}

var paramListCmd = &cobra.Command{
	Use:   "list <component-name>",
	Short: "List all parameters for a component(s)",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		if len(args) > 1 {
			return fmt.Errorf("'param list' takes at most one argument, that is the name of the component")
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

This command will display all parameters for the component specified. If a
component is not specified, parameters for all components will be listed.

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
			return fmt.Errorf("'param diff' takes exactly two arguments, that is the name of the environments to diff against")
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
	Long: `Pretty prints differences between the component parameters of two environments.

A component flag is accepted to diff against a single component. By default, the
diff is performed against all components.
`,
	Example: `# Diff between the component parameters on environments 'dev' and 'prod'
ks param diff dev prod

# Diff between the component 'guestbook' on environments 'dev' and 'prod'
ks param diff dev prod --component=guestbook`,
}

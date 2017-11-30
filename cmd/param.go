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

var paramShortDesc = map[string]string{
	"set":  `Change component or environment parameters (e.g. replica count, name)`,
	"list": `List known component parameters`,
	"diff": `Display differences between the component parameters of two environments`,
}

var paramCmd = &cobra.Command{
	Use:   "param",
	Short: `Manage ksonnet parameters for components and environments`,
	Long: `
Parameters are customizable fields that are used to expand and define ksonnet
*components*. Examples might include a deployment's 'name' or 'image'. Parameters
can also be defined on a *per-environment* basis. (Environments are ksonnet
deployment targets, e.g. specific clusters. For more info, run ` + "`ks env --help`" + `.)

For example, this allows a ` + "`dev`" + ` and ` + "`prod`" + ` environment to use the same component
manifest for an nginx deployment, but customize ` + "`prod`" + ` to use more replicas to meet
heavier load demands.

Params are structured as follows:

* App params (stored in ` + "`components/params.libsonnet`" + `)
    * Component-specific params
        * Originally populated from ` + "`ks generate`" + `
        * e.g. 80 for ` + "`deployment-example.port`" + `
    * Global params
        * Out of scope for CLI (requires Jsonnet editing)
        * Use to make a variable accessible to multiple components (e.g. service name)

* Per-environment params (stored in + ` + "`environments/<env-name>/params.libsonnet`" + `)
    * Component-specific params ONLY
    * Override app params (~inheritance)

Note that all of these params are tracked **locally** in version-controllable
Jsonnet files.

----
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
	Short: paramShortDesc["set"],
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		if len(args) != 3 {
			return fmt.Errorf("'param set' takes exactly three arguments, (1) the name of the component, in addition to (2) the key and (3) value of the parameter")
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
	Long: `
The ` + "`set`" + ` command sets component or environment parameters such as replica count
or name. Parameters are set individually, one at a time. All of these changes are
reflected in the ` + "`params.libsonnet`" + ` files.

For more details on how parameters are organized, see ` + "`ks param --help`" + `.

*(If you need to customize multiple parameters at once, we suggest that you modify
your ksonnet application's ` + " `components/params.libsonnet` " + `file directly. Likewise,
for greater customization of environment parameters, we suggest modifying the
` + " `environments/:name/params.libsonnet` " + `file.)*

### Related Commands

* ` + "`ks param diff` " + `— ` + paramShortDesc["diff"] + `
* ` + "`ks apply` " + `— ` + applyShortDesc + `

### Syntax
`,
	Example: `
# Update the replica count of the 'guestbook' component to 4.
ks param set guestbook replicas 4

# Update the replica count of the 'guestbook' component to 2, but only for the
# 'dev' environment
ks param set guestbook replicas 2 --env=dev`,
}

var paramListCmd = &cobra.Command{
	Use:   "list [<component-name>] [--env <env-name>]",
	Short: paramShortDesc["list"],
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
	Long: `
The ` + "`list`" + ` command displays all known component parameters or environment parameters.

If a component is specified, this command displays all of its specific parameters.
If a component is NOT specified, parameters for **all** components are listed.
Furthermore, parameters can be listed on a per-environment basis.

### Related Commands

* ` + "`ks param set` " + `— ` + paramShortDesc["set"] + `

### Syntax
`,
	Example: `
# List all component parameters
ks param list

# List all parameters for the component "guestbook"
ks param list guestbook

# List all parameters for the environment "dev"
ks param list --env=dev

# List all parameters for the component "guestbook" in the environment "dev"
ks param list guestbook --env=dev`,
}

var paramDiffCmd = &cobra.Command{
	Use:   "diff <env1> <env2> [--component <component-name>]",
	Short: paramShortDesc["diff"],
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

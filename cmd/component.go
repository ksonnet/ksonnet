// Copyright 2017 The ksonnet authors
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
	"github.com/spf13/viper"

	"github.com/ksonnet/ksonnet/actions"
	"github.com/ksonnet/ksonnet/pkg/kubecfg"
)

func init() {
	RootCmd.AddCommand(componentCmd)

	componentCmd.AddCommand(componentListCmd)
	componentListCmd.Flags().StringP(flagOutput, shortOutput, "", "Output format. Valid options: wide")
	viper.BindPFlag(vComponentListOutput, componentListCmd.Flags().Lookup(flagOutput))
	componentListCmd.Flags().String(flagNamespace, "", "Namespace")
	viper.BindPFlag(vComponentListNamespace, componentListCmd.Flags().Lookup(flagNamespace))

	componentCmd.AddCommand(componentRmCmd)

	componentRmCmd.PersistentFlags().String(flagComponent, "", "The component to be removed from components/")
}

var (
	vComponentListNamespace = "component-list-namespace"
	vComponentListOutput    = "component-list-output"
)

var componentCmd = &cobra.Command{
	Use:   "component",
	Short: "Manage ksonnet components",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("%s is not a valid subcommand\n\n%s", strings.Join(args, " "), cmd.UsageString())
		}
		return fmt.Errorf("Command 'component' requires a subcommand\n\n%s", cmd.UsageString())
	},
}

var componentListCmd = &cobra.Command{
	Use:   "list",
	Short: "List known components",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("'component list' takes zero arguments")
		}

		nsName := viper.GetString(vComponentListNamespace)
		output := viper.GetString(vComponentListOutput)

		ka, err := ksApp()
		if err != nil {
			return err
		}

		return actions.RunComponentList(ka, nsName, output)
	},
	Long: `
The ` + "`list`" + ` command displays all known components.

### Syntax
`,
	Example: `
# List all components
ks component list`,
}

var componentRmCmd = &cobra.Command{
	Use:   "rm <component-name>",
	Short: "Delete a component from the ksonnet application",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("'component rm' takes a single argument, that is the name of the component")
		}

		component := args[0]

		c := kubecfg.NewComponentRmCmd(component)
		return c.Run()
	},
	Long: `Delete a component from the ksonnet application. This is equivalent to deleting the
component file in the components directory and cleaning up all component
references throughout the project.`,
	Example: `# Remove the component 'guestbook'. This is equivalent to deleting guestbook.jsonnet
# in the components directory, and cleaning up references to the component
# throughout the ksonnet application.
ks component rm guestbook`,
}

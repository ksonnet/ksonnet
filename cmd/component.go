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

	"github.com/ksonnet/ksonnet/pkg/kubecfg"
)

func init() {
	RootCmd.AddCommand(componentCmd)

	componentCmd.AddCommand(componentListCmd)
}

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

		c := kubecfg.NewComponentListCmd()

		return c.Run(cmd.OutOrStdout())
	},
	Long: `
The ` + "`list`" + ` command displays all known components.

### Syntax
`,
	Example: `
# List all components
ks component list`,
}

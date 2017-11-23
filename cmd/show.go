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
	"os"

	"github.com/spf13/cobra"

	"github.com/ksonnet/ksonnet/metadata"
	"github.com/ksonnet/ksonnet/pkg/kubecfg"
)

const (
	flagFormat = "format"
)

func init() {
	RootCmd.AddCommand(showCmd)
	addEnvCmdFlags(showCmd)
	bindJsonnetFlags(showCmd)
	showCmd.PersistentFlags().StringP(flagFormat, "o", "yaml", "Output format.  Supported values are: json, yaml")
}

var showCmd = &cobra.Command{
	Use:   "show <env> [-c <component-filename>]",
	Short: "Show expanded manifests for a specific environment.",
	Long: `Show expanded manifests (resource definitions) for a specific environment. Jsonnet manifests,
each defining a ksonnet component, are expanded into their JSON or YAML equivalents (YAML is the default).
Any parameters for a component are resolved based on environment-specific values.

When NO component is specified via the` + "`-c`" + `flag, this command expands all of the files in the ` +
"`components/`" + ` directory into a list of resource definitions. This is the YAML version
of what gets deployed to your cluster with ` + "`ks apply <env>`" + `.

When a component IS specified via the` + "`-c`" + `flag, this command only expands the manifest for that
particular component.`,
	Example: `# Show all of the components for the 'dev' environment, in YAML
# (In other words, expands all manifests in the components/ directory)
ks show dev

# Show a single component from the 'prod' environment, in JSON
ks show prod -c redis -o json

# Show multiple components from the 'dev' environment, in YAML
ks show dev -c redis -c nginx-server
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("'show' takes at most a single argument, that is the name of the environment")
		}

		flags := cmd.Flags()
		var err error

		c := kubecfg.ShowCmd{}

		c.Format, err = flags.GetString(flagFormat)
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		wd := metadata.AbsPath(cwd)

		envSpec, err := parseEnvCmd(cmd, args)
		if err != nil {
			return err
		}

		objs, err := expandEnvCmdObjs(cmd, envSpec, wd)
		if err != nil {
			return err
		}

		return c.Run(objs, cmd.OutOrStdout())
	},
}

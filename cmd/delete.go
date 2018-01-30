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
	flagGracePeriod = "grace-period"
	deleteShortDesc = "Remove component-specified Kubernetes resources from remote clusters"
)

func init() {
	RootCmd.AddCommand(deleteCmd)
	addEnvCmdFlags(deleteCmd)
	bindClientGoFlags(deleteCmd)
	bindJsonnetFlags(deleteCmd)
	deleteCmd.PersistentFlags().Int64(flagGracePeriod, -1, "Number of seconds given to resources to terminate gracefully. A negative value is ignored")
}

var deleteCmd = &cobra.Command{
	Use:   "delete [env-name] [-c <component-name>]",
	Short: deleteShortDesc,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("'delete' requires an environment name; use `env list` to see available environments\n\n%s", cmd.UsageString())
		}
		env := args[0]

		flags := cmd.Flags()
		var err error

		c := kubecfg.DeleteCmd{}

		c.GracePeriod, err = flags.GetInt64(flagGracePeriod)
		if err != nil {
			return err
		}

		componentNames, err := flags.GetStringArray(flagComponent)
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		wd := metadata.AbsPath(cwd)

		c.ClientPool, c.Discovery, err = restClientPool(cmd, &env)
		if err != nil {
			return err
		}

		c.Namespace, err = namespace()
		if err != nil {
			return err
		}

		te := newCmdObjExpander(cmdObjExpanderConfig{
			cmd:        cmd,
			env:        env,
			components: componentNames,
			cwd:        wd,
		})
		objs, err := te.Expand()
		if err != nil {
			return err
		}

		return c.Run(objs)
	},
	Long: `
The ` + "`delete`" + ` command removes Kubernetes resources (described in local
*component* manifests) from a cluster. This cluster is determined by the mandatory
` + "`<env-name>`" + `argument.

An entire ksonnet application can be removed from a cluster, or just its specific
components.

**This command can be considered the inverse of the ` + "`ks apply`" + ` command.**

### Related Commands

* ` + "`ks diff` " + `— Compare manifests, based on environment or location (local or remote)
* ` + "`ks apply` " + `— ` + applyShortDesc + `

### Syntax
`,
	Example: `# Delete resources from the 'dev' environment, based on ALL of the manifests in your
# ksonnet app's 'components/' directory. This command works in any subdirectory
# of the app.
ks delete dev

# Delete resources described by the 'nginx' component. $KUBECONFIG is overridden by
# the CLI-specified './kubeconfig', so these changes are deployed to the current
# context's cluster (not the 'default' environment)
ks delete --kubeconfig=./kubeconfig -c nginx`,
}

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

package clicmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/ksonnet/ksonnet/pkg/actions"
	"github.com/ksonnet/ksonnet/pkg/client"
)

const (
	diffShortDesc = "Compare manifests, based on environment or location (local or remote)"

	vDiffComponentNames = "diff-component-names"
)

var (
	diffClientConfig *client.Config
)

func init() {
	diffClientConfig = client.NewDefaultClientConfig(ka)
	diffClientConfig.BindClientGoFlags(diffCmd)
	bindJsonnetFlags(diffCmd, "diff")

	diffCmd.Flags().StringSliceP(flagComponent, shortComponent, nil, "Name of a specific component")
	viper.BindPFlag(vDiffComponentNames, diffCmd.Flags().Lookup(flagComponent))

	RootCmd.AddCommand(diffCmd)
}

var diffCmd = &cobra.Command{
	Use:   "diff <location1:env1> [location2:env2]",
	Short: diffShortDesc,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("'diff' requires at least one argument, that is the name of the environment\n\n%s", cmd.UsageString())
		}
		if len(args) > 2 {
			return fmt.Errorf("'diff' takes at most two arguments, that are the name of the environments\n\n%s", cmd.UsageString())
		}

		m := map[string]interface{}{
			actions.OptionApp:            ka,
			actions.OptionClientConfig:   diffClientConfig,
			actions.OptionSrc1:           args[0],
			actions.OptionComponentNames: viper.GetStringSlice(vDiffComponentNames),
		}

		if len(args) == 2 {
			m[actions.OptionSrc2] = args[1]
		}

		if err := extractJsonnetFlags("apply"); err != nil {
			return errors.Wrap(err, "handle jsonnet flags")
		}

		return runAction(actionDiff, m)
	},
	Long: `
The ` + "`diff`" + ` command displays standard file diffs, and can be used to compare manifests
based on *environment* or location ('local' ksonnet app manifests or what's running
on a 'remote' server).

Using this command, you can compare:

1. *Remote* and *local* manifests for a single environment
2. *Remote* manifests for two separate environments
3. *Local* manifests for two separate environments
4. A *remote* manifest in one environment and a *local* manifest in another environment

To see the official syntax, see the examples below. Make sure that your $KUBECONFIG
matches what you've defined in environments.

When NO component is specified (no ` + "`-c`" + ` flag), this command diffs all of
the files in the ` + "`components/`" + ` directory.

When a component IS specified via the ` + "`-c`" + ` flag, this command only checks
the manifest for that particular component.

### Related Commands

* ` + "`ks param diff` " + `— ` + paramShortDesc["diff"] + `

### Syntax
`,
	Example: `
# Show diff between remote and local manifests for a single 'dev' environment.
# This command diffs *all* components in the ksonnet app, and can be used in any
# of that app's subdirectories.
ks diff remote:dev local:dev

# Shorthand for the previous command (remote 'dev' and local 'dev')
ks diff dev

# Show diff between the remote resources running in two different ksonnet environments
# 'us-west/dev' and 'us-west/prod'. This command diffs all resources defined in
# the ksonnet app.
ks diff remote:us-west/dev remote:us-west/prod

# Show diff between local manifests in the 'us-west/dev' environment and remote
# resources in the 'us-west/prod' environment, for an entire ksonnet app
ks diff local:us-west/dev remote:us-west/prod

# Show diff between what's in the local manifest and what's actually running in the
# 'dev' environment, but for the Redis component ONLY
ks diff dev -c redis
`,
}

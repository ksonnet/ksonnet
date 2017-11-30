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
	flagCreate = "create"
	flagSkipGc = "skip-gc"
	flagGcTag  = "gc-tag"
	flagDryRun = "dry-run"

	// AnnotationGcTag annotation that triggers
	// garbage collection. Objects with value equal to
	// command-line flag that are *not* in config will be deleted.
	AnnotationGcTag = "kubecfg.ksonnet.io/garbage-collect-tag"

	// AnnotationGcStrategy controls gc logic.  Current values:
	// `auto` (default if absent) - do garbage collection
	// `ignore` - never garbage collect this object
	AnnotationGcStrategy = "kubecfg.ksonnet.io/garbage-collect-strategy"

	// GcStrategyAuto is the default automatic gc logic
	GcStrategyAuto = "auto"
	// GcStrategyIgnore means this object should be ignored by garbage collection
	GcStrategyIgnore = "ignore"
)

var applyShortDesc = `Apply local Kubernetes manifests (components) to remote clusters`

func init() {
	RootCmd.AddCommand(applyCmd)

	addEnvCmdFlags(applyCmd)
	bindClientGoFlags(applyCmd)
	bindJsonnetFlags(applyCmd)
	applyCmd.PersistentFlags().Bool(flagCreate, true, "Option to create resources if they do not already exist on the cluster")
	applyCmd.PersistentFlags().Bool(flagSkipGc, false, "Option to skip garbage collection, even with --"+flagGcTag+" specified")
	applyCmd.PersistentFlags().String(flagGcTag, "", "A tag that's (1) added to all updated objects (2) used to garbage collect existing objects that are no longer in the manifest")
	applyCmd.PersistentFlags().Bool(flagDryRun, false, "Option to preview the list of operations without changing the cluster state")
}

var applyCmd = &cobra.Command{
	Use:   "apply <env-name> [-c <component-name>] [--dry-run]",
	Short: applyShortDesc,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("'apply' requires an environment name; use `env list` to see available environments\n\n%s", cmd.UsageString())
		}
		env := args[0]

		flags := cmd.Flags()
		var err error

		c := kubecfg.ApplyCmd{}

		c.Create, err = flags.GetBool(flagCreate)
		if err != nil {
			return err
		}

		c.GcTag, err = flags.GetString(flagGcTag)
		if err != nil {
			return err
		}

		c.SkipGc, err = flags.GetBool(flagSkipGc)
		if err != nil {
			return err
		}

		c.DryRun, err = flags.GetBool(flagDryRun)
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

		objs, err := expandEnvCmdObjs(cmd, env, componentNames, wd)
		if err != nil {
			return err
		}

		return c.Run(objs, wd)
	},
	Long: `
The ` + "`apply`" + `command uses local manifest(s) to update (and optionally create)
Kubernetes resources on a remote cluster. This cluster is determined by the
mandatory ` + "`<env-name>`" + ` argument.

The manifests themselves correspond to the components of your app, and reside
in your app's ` + "`components/`" + ` directory. When applied, the manifests are fully
expanded using the parameters of the specified environment.

By default, all component manifests are applied. To apply a subset of components,
use the ` + "`--component` " + `flag, as seen in the examples below.

Note that this command needs to be run *within* a ksonnet app directory.

### Related Commands

* ` + "`ks diff` " + `— Compare manifests, based on environment or location (local or remote)
* ` + "`ks delete` " + deleteShortDesc + `

### Syntax
`,
	Example: `
# Create or update all resources described in the ksonnet application, specifically
# the ones running in the 'dev' environment. This command works in any subdirectory
# of the app.
#
# This essentially deploys all components in the 'components/' directory.
ks apply dev

# Similar to the previous command, but does not immediately execute. Use this to
# see a preview of the cluster-changing actions.
ks apply dev --dry-run

# Create or update the single 'guestbook-ui' component of a ksonnet app, specifically
# the instance running in the 'dev' environment.
#
# This essentially deploys 'components/guestbook-ui.jsonnet'.
ks apply dev -c guestbook-ui

# Create or update multiple components in a ksonnet application (e.g. 'guestbook-ui'
# and 'ngin-depl') for the 'dev' environment. Does not create resources that are
# not already present on the cluster.
#
# This essentially deploys 'components/guestbook-ui.jsonnet' and
# 'components/nginx-depl.jsonnet'.
ks apply dev -c guestbook-ui -c nginx-depl --create false
`,
}

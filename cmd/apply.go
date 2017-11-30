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

func init() {
	RootCmd.AddCommand(applyCmd)

	addEnvCmdFlags(applyCmd)
	bindClientGoFlags(applyCmd)
	bindJsonnetFlags(applyCmd)
	applyCmd.PersistentFlags().Bool(flagCreate, true, "Create missing resources")
	applyCmd.PersistentFlags().Bool(flagSkipGc, false, "Don't perform garbage collection, even with --"+flagGcTag)
	applyCmd.PersistentFlags().String(flagGcTag, "", "Add this tag to updated objects, and garbage collect existing objects with this tag and not in config")
	applyCmd.PersistentFlags().Bool(flagDryRun, false, "Perform only read-only operations")
}

var applyCmd = &cobra.Command{
	Use:   "apply <env-name>",
	Short: `Apply local Kubernetes manifests to remote clusters`,
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
Update (or optionally create) Kubernetes resources on the cluster using your
local Kubernetes manifests. Use the` + " `--create` " + `flag to control whether
they are created if they do not exist (default: true).

The local Kubernetes manifests that are applied reside in your ` + "`components/`" + `
directory. When applied, the manifests are fully expanded using the paremeters
of the specified environment.

By default, all manifests are applied. To apply a subset of manifests, use the
` + "`--component` " + `flag, as seen in the examples below.

### Related Commands

* ` + "`ks delete` " + `— Delete the component manifests on your cluster

### Syntax
`,
	Example: `
# Create or update all resources described in a ksonnet application, and
# running in the 'dev' environment. Can be used in any subdirectory of the
# application.
#
# This is equivalent to applying all components in the 'components/' directory.
ks apply dev

# Create or update the single resource 'guestbook-ui' described in a ksonnet
# application, and running in the 'dev' environment. Can be used in any
# subdirectory of the application.
#
# This is equivalent to applying the component with the same file name (excluding
# the extension) 'guestbook-ui' in the 'components/' directory.
ks apply dev -c guestbook-ui

# Create or update the multiple resources, 'guestbook-ui' and 'nginx-depl'
# described in a ksonnet application, and running in the 'dev' environment. Can
# be used in any subdirectory of the application.
#
# This is equivalent to applying the component with the same file name (excluding
# the extension) 'guestbook-ui' and 'nginx-depl' in the 'components/' directory.
ks apply dev -c guestbook-ui -c nginx-depl
`,
}

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
	applyCmd.PersistentFlags().Bool(flagCreate, true, "Create resources if they do not already exist on the cluster")
	applyCmd.PersistentFlags().Bool(flagSkipGc, false, "Opt out of garbage collection, even with --"+flagGcTag+" specified")
	applyCmd.PersistentFlags().String(flagGcTag, "", "The specified tag is (1) added to all updated objects (2) used to identify what needs to be removed from the cluster (previously tagged objects that are no longer in the specified manifest)")
	applyCmd.PersistentFlags().Bool(flagDryRun, false, "Preview the list of operations without changing the cluster state")
}

var applyCmd = &cobra.Command{
	Use:   "apply [env-name] [-f <file-or-dir>] [--dry-run]",
	Short: `Apply local manifests to a remote cluster`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("'apply' takes at most a single argument, that is the name of the environment")
		}

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

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		wd := metadata.AbsPath(cwd)

		envSpec, err := parseEnvCmd(cmd, args)
		if err != nil {
			return err
		}

		c.ClientPool, c.Discovery, err = restClientPool(cmd, envSpec.env)
		if err != nil {
			return err
		}

		c.Namespace, err = namespace()
		if err != nil {
			return err
		}

		objs, err := expandEnvCmdObjs(cmd, envSpec, wd)
		if err != nil {
			return err
		}

		return c.Run(objs, wd)
	},
	Long: `Use local manifest(s) to update (and optionally create) Kubernetes resources on the cluster.
This cluster is determined by the mandatory environment argument.
To control whether resources are created if they do not exist, use the` +
" `--create` " + `flag (defaults to ` + "`true`" +`).

An entire ksonnet application can be applied, or individual JSON, YAML, or Jsonnet files.

Note that this command needs to be run *within* a ksonnet app directory.`,
	Example: `# Preview the set of actions that are executed by 'apply'.
ks apply dev --dry-run

# Update (but do not create) all resources described in a ksonnet
# application. Changes are deployed to the cluster pointed at the 'prod'
# environment. Can be used in any subdirectory of that application.
ks apply prod --create=false

# Create or update resources described in a JSON file. Changes are deployed
# to the cluster pointed at the 'dev' environment.
ks apply dev -f ./lib/pod.json

# Create or update resources described in a YAML file. Changes are deployed to the
# cluster pointed to by the current context of './kubeconfig'. This kubeconfig
# context overrides the 'default' environment.
ks apply default --kubeconfig=./kubeconfig -f ./lib/pod.yaml`,
}

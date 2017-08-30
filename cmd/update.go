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
	"os"

	"github.com/spf13/cobra"

	"github.com/ksonnet/kubecfg/metadata"
	"github.com/ksonnet/kubecfg/pkg/kubecfg"
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
	RootCmd.AddCommand(updateCmd)

	addEnvCmdFlags(updateCmd)
	updateCmd.PersistentFlags().Bool(flagCreate, true, "Create missing resources")
	updateCmd.PersistentFlags().Bool(flagSkipGc, false, "Don't perform garbage collection, even with --"+flagGcTag)
	updateCmd.PersistentFlags().String(flagGcTag, "", "Add this tag to updated objects, and garbage collect existing objects with this tag and not in config")
	updateCmd.PersistentFlags().Bool(flagDryRun, false, "Perform only read-only operations")
}

var updateCmd = &cobra.Command{
	Use: "update [<env>|-f <file-or-dir>]",
	Short: `Update (or optionally create) Kubernetes resources on the cluster using the
local configuration. Accepts JSON, YAML, or Jsonnet.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		var err error

		c := kubecfg.UpdateCmd{}

		c.Environment, c.Files, err = parseEnvCmd(cmd, args)
		if err != nil {
			return err
		}

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

		c.ClientPool, c.Discovery, err = restClientPool(cmd)
		if err != nil {
			return err
		}

		c.DefaultNamespace, _, err = clientConfig.Namespace()
		if err != nil {
			return err
		}

		c.Expander, err = newExpander(cmd)
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		return c.Run(metadata.AbsPath(cwd))
	},
	Long: `Update (or optionally create) Kubernetes resources on the cluster using the
local configuration. Use the '--create' flag to control whether we create them
if they do not exist (default: true).

ksonnet applications are accepted, as well as normal JSON, YAML, and Jsonnet
files.`,
	Example: `  # Create or update all resources described in a ksonnet application, and
  # running in the 'dev' environment. Can be used in any subdirectory of the
  # application.
  ksonnet update dev

  # Create or update resources described in a YAML file. Automatically picks up
  # the cluster's location from '$KUBECONFIG'.
  ksonnet appy -f ./pod.yaml

  # Update resources described in a YAML file, and running in cluster referred
  # to by './kubeconfig'.
  ksonnet update --kubeconfig=./kubeconfig -f ./pod.yaml

  # Display set of actions we will execute when we run 'update'.
  ksonnet update dev --dry-run`,
}

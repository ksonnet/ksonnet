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

func init() {
	RootCmd.AddCommand(updateCmd)

	addEnvCmdFlags(updateCmd)
	bindClientGoFlags(updateCmd)
	bindJsonnetFlags(updateCmd)

	updateCmd.PersistentFlags().Bool(flagCreate, true, "Option to create resources if they do not already exist on the cluster")
	updateCmd.PersistentFlags().Bool(flagSkipGc, false, "Option to skip garbage collection, even with --"+flagGcTag+" specified")
	updateCmd.PersistentFlags().String(flagGcTag, "", "The specified tag is (1) added to all updated objects (2) used to identify what needs to be removed from the cluster (previously tagged objects that are no longer in the specified manifest)")
	updateCmd.PersistentFlags().Bool(flagDryRun, false, "Option to preview the list of operations without changing the cluster state")
}

var updateCmd = &cobra.Command{
	Deprecated: "NOTE: Use 'apply' instead",
	Hidden:     true,
	Use:        "update [<env>|-f <file-or-dir>]",
	Short: `[DEPRECATED] Use local manifest(s) to update (and optionally create) Kubernetes resources on the cluster.
Accepts JSON, YAML, or Jsonnet files.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("'update' takes at most a single argument, that is the name of the environment")
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

		c.ClientPool, c.Discovery, err = restClientPool(cmd, nil)
		if err != nil {
			return err
		}

		c.Namespace, err = namespace()
		if err != nil {
			return err
		}

		envSpec, err := parseEnvCmd(cmd, args)
		if err != nil {
			return err
		}

		objs, err := expandEnvCmdObjs(cmd, envSpec, wd)
		if err != nil {
			return err
		}

		return c.Run(objs, wd)
	},
}

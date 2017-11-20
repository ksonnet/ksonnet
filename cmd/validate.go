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
	RootCmd.AddCommand(validateCmd)
	addEnvCmdFlags(validateCmd)
	bindJsonnetFlags(validateCmd)
	bindClientGoFlags(validateCmd)
}

var validateCmd = &cobra.Command{
	Use:   "validate [env-name] [-f <file-or-dir>]",
	Short: "Compare generated manifests against the server's OpenAPI spec",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) > 1 {
			return fmt.Errorf("'validate' takes at most a single argument, that is the name of the environment")
		}

		var err error

		c := kubecfg.ValidateCmd{}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		wd := metadata.AbsPath(cwd)

		envSpec, err := parseEnvCmd(cmd, args)
		if err != nil {
			return err
		}

		_, c.Discovery, err = restClientPool(cmd, nil)
		if err != nil {
			return err
		}

		objs, err := expandEnvCmdObjs(cmd, envSpec, wd)
		if err != nil {
			return err
		}

		return c.Run(objs, cmd.OutOrStdout())
	},
	Long: `Validate that a ksonnet application or file is compliant with an environment's
Kubernetes specification.`,
	Example: `# Validate all resources described in a ksonnet application, expanding
# ksonnet code with 'dev' environment parameters where necessary.
# Note that this ONLY validates manifests that are written with ksonnet-lib. It does
# not check YAML, JSON, or vanilla Jsonnet.
ksonnet validate dev

# Validate that resources described in a YAML file adhere to the Kubernetes spec
# used by the 'dev' environment cluster.
ksonnet validate dev -f ./pod.yaml`,
}

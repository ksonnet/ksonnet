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
	"github.com/spf13/cobra"

	"github.com/ksonnet/kubecfg/pkg/kubecfg"
)

func init() {
	RootCmd.AddCommand(validateCmd)
	addEnvCmdFlags(validateCmd)
}

var validateCmd = &cobra.Command{
	Use:   "validate [env-name] [-f <file-or-dir>]",
	Short: "Compare generated manifest against server OpenAPI spec",
	RunE: func(cmd *cobra.Command, args []string) error {
		var err error

		c := kubecfg.ValidateCmd{}

		_, c.Discovery, err = restClientPool(cmd)
		if err != nil {
			return err
		}

		objs, err := expandEnvCmdObjs(cmd, args)
		if err != nil {
			return err
		}

		return c.Run(objs, cmd.OutOrStdout())
	},
	Long: `Validate that an application or file is compliant with the Kubernetes
specification.

ksonnet applications are accepted, as well as normal JSON, YAML, and Jsonnet
files.`,
	Example: `  # Validate all resources described in a ksonnet application, expanding
  # ksonnet code with 'dev' environment where necessary (i.e., not YAML, JSON,
  # or non-ksonnet Jsonnet code).
  ksonnet validate dev

  # Validate resources described in a YAML file.
  ksonnet validate -f ./pod.yaml

  # Validate resources described in the JSON file against existing resources
  # in the cluster the 'dev' environment is pointing at.
  ksonnet validate dev -f ./pod.yaml

  # Validate resources described in a Jsonnet file. Does not expand using
  # environment bindings.
  ksonnet validate -f ./pod.jsonnet`,
}

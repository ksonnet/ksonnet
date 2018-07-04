// Copyright 2018 The ksonnet authors
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
	"github.com/ksonnet/ksonnet/pkg/actions"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	vModuleListEnv    = "module-list-env"
	vModuleListOutput = "module-list-output"
)

func newModuleListCmd(a app.App) *cobra.Command {
	moduleListCmd := &cobra.Command{
		Use:   "list",
		Short: "List modules",
		Long:  `List modules`,
		RunE: func(cmd *cobra.Command, args []string) error {
			m := map[string]interface{}{
				actions.OptionApp:     a,
				actions.OptionEnvName: viper.GetString(vModuleListEnv),
				actions.OptionOutput:  viper.GetString(vModuleListOutput),
			}

			return runAction(actionModuleList, m)
		},
	}

	addCmdOutput(moduleListCmd, vModuleListOutput)
	moduleListCmd.Flags().String(flagEnv, "", "Environment to list modules for")
	viper.BindPFlag(vModuleListEnv, moduleListCmd.Flags().Lookup(flagEnv))

	return moduleListCmd
}

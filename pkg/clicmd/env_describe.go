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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newEnvDescribeCmd(a app.App) *cobra.Command {
	envDescribeCmd := &cobra.Command{
		Use:   "describe <env>",
		Short: "Describe an environment",
		Long:  `describe`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("env describe <environment>")
			}

			m := map[string]interface{}{
				actions.OptionApp:     a,
				actions.OptionEnvName: args[0],
			}

			return runAction(actionEnvDescribe, m)
		},
	}

	return envDescribeCmd

}

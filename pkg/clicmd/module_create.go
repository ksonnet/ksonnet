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

func newModuleCreateCmd(a app.App) *cobra.Command {
	moduleCreateCmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a module",
		Long:  `Create a module`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("module create <module name>")
			}

			m := map[string]interface{}{
				actions.OptionApp:    a,
				actions.OptionModule: args[0],
			}

			return runAction(actionModuleCreate, m)
		},
	}

	return moduleCreateCmd
}

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
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	vParamDeleteEnv = "param-delete-env"
	paramDeleteLong = `
The ` + "`delete`" + ` command deletes component or environment parameters.

### Related Commands

* ` + "`ks param set` " + `— ` + paramShortDesc["set"] + `
* ` + "`ks param diff` " + `— ` + paramShortDesc["diff"] + `
* ` + "`ks apply` " + `— ` + applyShortDesc + `

### Syntax
`
	paramDeleteExample = `
# Delete 'guestbook' component replica parameter
ks param delete guestbook replicas

# Delete 'guestbook' component replicate in 'dev' environment
ks param delete guestbook replicas --env=dev`
)

func newParamDeleteCmd() *cobra.Command {
	paramDeleteCmd := &cobra.Command{
		Use:     "delete [component-name] <param-key>",
		Short:   paramShortDesc["delete"],
		Long:    paramDeleteLong,
		Example: paramDeleteExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			var name string
			var path string

			switch len(args) {
			default:
				return errors.New("invalid arguments for 'param delete'")
			case 2:
				name = args[0]
				path = args[1]
			case 1:
				path = args[0]
			}

			m := map[string]interface{}{
				actions.OptionName:    name,
				actions.OptionPath:    path,
				actions.OptionEnvName: viper.GetString(vParamDeleteEnv),
			}
			addGlobalOptions(m)

			return runAction(actionParamDelete, m)
		},
	}

	paramDeleteCmd.Flags().String(flagEnv, "", "Specify environment to delete parameter from")
	viper.BindPFlag(vParamDeleteEnv, paramDeleteCmd.Flags().Lookup(flagEnv))

	return paramDeleteCmd
}

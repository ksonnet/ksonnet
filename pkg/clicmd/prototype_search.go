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
	"fmt"

	"github.com/ksonnet/ksonnet/pkg/actions"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	vPrototypeSearchOutput = "prototype-search-output"
)

var (
	prototypeSearchLong = `
The ` + "`prototype search`" + ` command allows you to search for specific prototypes by name.
Specifically, it matches any prototypes with names that contain the string <name-substring>.

### Related Commands

* ` + "`ks prototype describe` " + `— ` + protoShortDesc["describe"] + `
* ` + "`ks prototype list` " + `— ` + protoShortDesc["list"] + `

### Syntax
`
	prototypeSearchExample = `
# Search for prototypes with names that contain the string 'service'.
ks prototype search service`
)

func newPrototypeSearchCmd(a app.App) *cobra.Command {
	prototypeSearchCmd := &cobra.Command{
		Use:     "search <name-substring>",
		Short:   protoShortDesc["search"],
		Long:    prototypeSearchLong,
		Example: prototypeSearchExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("Command 'prototype search' requires a prototype name\n\n%s", cmd.UsageString())
			}

			m := map[string]interface{}{
				actions.OptionApp:    a,
				actions.OptionQuery:  args[0],
				actions.OptionOutput: viper.GetString(vPrototypeSearchOutput),
			}

			return runAction(actionPrototypeSearch, m)
		},
	}

	addCmdOutput(prototypeSearchCmd, vPrototypeSearchOutput)
	return prototypeSearchCmd
}

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
)

var (
	prototypePreviewLong = `
This ` + "`preview`" + ` command expands a prototype with CLI flag parameters, and
emits the resulting manifest to stdout. This allows you to see the potential
output of a ` + "`ks generate`" + ` command without actually creating a new component file.

The output is formatted in Jsonnet. To see YAML or JSON equivalents, first create
a component with ` + "`ks generate`" + ` and then use ` + "`ks show`" + `.

### Related Commands

* ` + "`ks generate` " + `— ` + protoShortDesc["use"] + `

### Syntax
`
	prototypePreviewExample = `
# Preview prototype 'io.ksonnet.pkg.single-port-deployment', using the
# 'nginx' image, and port 80 exposed.
ks prototype preview single-port-deployment \
  --name=nginx                              \
  --image=nginx                             \
  --port=80`
)

func newPrototypePreviewCmd(a app.App) *cobra.Command {
	prototypePreviewCmd := &cobra.Command{
		Use:                "preview <prototype-name> [parameter-flags]",
		Short:              protoShortDesc["preview"],
		Long:               prototypePreviewLong,
		Example:            prototypePreviewExample,
		DisableFlagParsing: true,
		RunE: func(cmd *cobra.Command, rawArgs []string) error {
			if len(rawArgs) == 1 && (rawArgs[0] == "--help" || rawArgs[0] == "-h") {
				cmd.Help()
				return nil
			}

			if len(rawArgs) < 1 {
				return fmt.Errorf("Command 'prototype preview' requires a prototype name\n\n%s", cmd.UsageString())
			}

			m := map[string]interface{}{
				actions.OptionApp:       a,
				actions.OptionQuery:     rawArgs[0],
				actions.OptionArguments: rawArgs[1:],
			}

			return runAction(actionPrototypePreview, m)
		},
	}

	return prototypePreviewCmd
}

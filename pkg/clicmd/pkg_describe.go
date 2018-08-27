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
	"github.com/spf13/cobra"
)

var (
	pkgDescribeLong = `
The ` + "`describe`" + ` command outputs documentation for a package that is available
(e.g. downloaded) in the current ksonnet application. (This must belong to an already
known ` + "`<registry-name>`" + ` like *incubator*). The output includes:

1. The package name
2. A brief description provided by the package authors
3. A list of available prototypes provided by the package

### Related Commands

* ` + "`ks pkg list` " + `— ` + pkgShortDesc["list"] + `
* ` + "`ks prototype describe` " + `— ` + protoShortDesc["describe"] + `
* ` + "`ks generate` " + `— ` + protoShortDesc["use"] + `

### Syntax
`
)

func newPkgDescribeCmd() *cobra.Command {
	pkgDescribeCmd := &cobra.Command{
		Use:   "describe [<registry-name>/]<package-name>",
		Short: pkgShortDesc["describe"],
		Long:  pkgDescribeLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return fmt.Errorf("Command 'pkg describe' requires a package name\n\n%s", cmd.UsageString())
			}

			m := map[string]interface{}{
				actions.OptionPackageName: args[0],
			}
			addGlobalOptions(m)

			return runAction(actionPkgDescribe, m)
		},
	}

	return pkgDescribeCmd
}

package cmd

import (
	"fmt"

	"github.com/ksonnet/ksonnet/actions"
	"github.com/spf13/cobra"
)

var prototypeListCmd = &cobra.Command{
	Use:   "list",
	Short: protoShortDesc["list"],
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("Command 'prototype list' does not take any arguments")
		}

		return actions.RunPrototypeList(ka)
	},
	Long: `
The ` + "`list`" + ` command displays all prototypes that are available locally, as
well as brief descriptions of what they generate.

ksonnet comes with a set of system prototypes that you can use out-of-the-box
(e.g.` + " `io.ksonnet.pkg.configMap`" + `). However, you can use more advanced
prototypes like ` + "`io.ksonnet.pkg.redis-stateless`" + ` by downloading extra packages
from the *incubator* registry.

### Related Commands

* ` + "`ks prototype describe` " + `— ` + protoShortDesc["describe"] + `
* ` + "`ks prototype preview` " + `— ` + protoShortDesc["preview"] + `
* ` + "`ks prototype use` " + `— ` + protoShortDesc["use"] + `
* ` + "`ks pkg install` " + pkgShortDesc["install"] + `

### Syntax
`,
}

func init() {
	prototypeCmd.AddCommand(prototypeListCmd)
}

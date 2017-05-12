package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	jsonnet "github.com/strickyak/jsonnet_cgo"
)

func init() {
	RootCmd.AddCommand(versionCmd)
}

// Version is overridden by main
var Version = "(dev build)"

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(cmd *cobra.Command, args []string) {
		out := cmd.OutOrStdout()
		fmt.Fprintln(out, "kubecfg version:", Version)
		fmt.Fprintln(out, "jsonnet version:", jsonnet.Version())
	},
}

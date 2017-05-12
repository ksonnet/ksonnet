package main

import (
	"fmt"
	"os"

	"github.com/anguslees/kubecfg/cmd"
)

// Version is overridden using `-X main.version` during release builds
var version = "(dev build)"

func main() {
	cmd.Version = version

	if err := cmd.RootCmd.Execute(); err != nil {
		fmt.Println("got error")
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}

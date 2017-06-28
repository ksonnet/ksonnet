package main

import (
	log "github.com/sirupsen/logrus"

	"github.com/ksonnet/kubecfg/cmd"
)

// Version is overridden using `-X main.version` during release builds
var version = "(dev build)"

func main() {
	cmd.Version = version

	if err := cmd.RootCmd.Execute(); err != nil {
		// PersistentPreRunE may not have been run for early
		// errors, like invalid command line flags.
		logFmt := cmd.NewLogFormatter(log.StandardLogger().Out)
		log.SetFormatter(logFmt)

		log.Fatal(err.Error())
	}
}

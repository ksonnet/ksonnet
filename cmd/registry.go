package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/ksonnet/ksonnet/metadata"
	"github.com/ksonnet/ksonnet/utils"
	"github.com/spf13/cobra"
)

func init() {
	RootCmd.AddCommand(registryCmd)
	registryCmd.AddCommand(registryListCmd)
	registryCmd.AddCommand(registryDescribeCmd)
}

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: `Manage registries for current project`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Command 'registry' requires a subcommand\n\n%s", cmd.UsageString())
	},
	Long: `Manage and inspect ksonnet registries.

Registries contain a set of versioned libraries that the user can install and
manage in a ksonnet project using the CLI. A typical library contains:

  1. A set of "parts", pre-fabricated API objects which can be combined together
     to configure a Kubernetes application for some task. For example, the Redis
     library may contain a Deployment, a Service, a Secret, and a
     PersistentVolumeClaim, but if the user is operating it as a cache, they may
     only need the first three of these.
  2. A set of "prototypes", which are pre-fabricated combinations of these
     parts, made to make it easier to get started using a library. See the
     documentation for 'ks prototype' for more information.`,
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: `List all registries known to the current ksonnet app.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		const (
			nameHeader     = "NAME"
			protocolHeader = "PROTOCOL"
			uriHeader      = "URI"
		)

		if len(args) != 0 {
			return fmt.Errorf("Command 'registry list' does not take arguments")
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		wd := metadata.AbsPath(cwd)

		manager, err := metadata.Find(wd)
		if err != nil {
			return err
		}

		app, err := manager.AppSpec()
		if err != nil {
			return err
		}

		rows := [][]string{
			[]string{nameHeader, protocolHeader, uriHeader},
			[]string{
				strings.Repeat("=", len(nameHeader)),
				strings.Repeat("=", len(protocolHeader)),
				strings.Repeat("=", len(uriHeader)),
			},
		}
		for name, regRef := range app.Registries {
			rows = append(rows, []string{name, regRef.Protocol, regRef.URI})
		}

		formatted, err := utils.PadRows(rows)
		if err != nil {
			return err
		}
		fmt.Print(formatted)
		return nil
	},
}

var registryDescribeCmd = &cobra.Command{
	Use:   "describe <registry-name>",
	Short: `Describe a ksonnet registry`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Command 'registry describe' takes one argument, which is the name of the registry to describe")
		}
		name := args[0]

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		wd := metadata.AbsPath(cwd)

		manager, err := metadata.Find(wd)
		if err != nil {
			return err
		}

		app, err := manager.AppSpec()
		if err != nil {
			return err
		}

		regRef, exists := app.GetRegistryRef(name)
		if !exists {
			return fmt.Errorf("Registry '%s' doesn't exist", name)
		}

		reg, _, err := manager.GetRegistry(name)
		if err != nil {
			return err
		}

		fmt.Println(`REGISTRY NAME:`)
		fmt.Println(regRef.Name)
		fmt.Println()
		fmt.Println(`URI:`)
		fmt.Println(regRef.URI)
		fmt.Println()
		fmt.Println(`PROTOCOL:`)
		fmt.Println(regRef.Protocol)
		fmt.Println()
		fmt.Println(`LIBRARIES:`)

		for _, lib := range reg.Libraries {
			fmt.Printf("  %s\n", lib.Path)
		}

		return nil
	},

	Long: `Output documentation for some ksonnet registry prototype uniquely identified in
the current ksonnet project by some` + " `registry-name`" + `.`,
}

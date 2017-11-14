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
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: `List all registries known to current ksonnet app`,
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
	Long: ``,

	Example: ``,
}

var registryDescribeCmd = &cobra.Command{
	Use:   "describe",
	Short: `Manage registries for current project`,
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
	Long: ``,

	Example: ``,
}

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

var regShortDesc = map[string]string{
	"list":     `List all registries known to the current ksonnet app.`,
	"describe": `Describe a ksonnet registry and the packages it contains`,
}

var registryCmd = &cobra.Command{
	Use:   "registry",
	Short: `Manage registries for current project`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("%s is not a valid subcommand\n\n%s", strings.Join(args, " "), cmd.UsageString())
		}
		return fmt.Errorf("Command 'registry' requires a subcommand\n\n%s", cmd.UsageString())
	},
	Long: `
A ksonnet registry is basically a repository for *packages*. (Registry here is
used in the same sense as a container image registry). Registries are identified
by a ` + "`registry.yaml`" + ` in their root that declares which packages they contain.

Specifically, registries contain a set of versioned packages that the user can
install and manage in a given ksonnet app, using the CLI. A typical package contains:

1. **A library definining a set of "parts"**. These are pre-fabricated API objects
which can be combined together to configure a Kubernetes application for some task.
(e.g. a Deployment, a Service, and a Secret, specifically tailored for Redis).

2. **A set of "prototypes"**, which are pre-fabricated combinations of parts, as
described above. (See ` + "`ks prototype --help`" + ` for more information.)

----
`,
}

var registryListCmd = &cobra.Command{
	Use:   "list",
	Short: regShortDesc["list"],
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
	Long: `
The ` + "`list`" + ` command displays all known ksonnet registries in a table. This
table includes the following info:

1. Registry name
2. Protocol (e.g. ` + "`github`" + `)
3. Registry URI

### Related Commands

* ` + "`ks registry describe` " + `— ` + regShortDesc["describe"] + `

### Syntax
`,
}

var registryDescribeCmd = &cobra.Command{
	Use:   "describe <registry-name>",
	Short: regShortDesc["describe"],
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
		fmt.Println(`PACKAGES:`)

		for _, lib := range reg.Libraries {
			fmt.Printf("  %s\n", lib.Path)
		}

		return nil
	},

	Long: `
The ` + "`describe`" + ` command outputs documentation for the ksonnet registry identified
by `+ "`<registry-name>`" + `. Specifically, it displays the following:

1. Registry URI
2. Protocol (e.g. ` + "`github`" + `)
3. List of packages included in the registry

### Related Commands

* ` + "`ks pkg install` " + `— ` + pkgShortDesc["install"] + `

### Syntax
`,
}

// Copyright 2018 The kubecfg authors
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

package cmd

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/ksonnet/ksonnet/metadata"
	"github.com/spf13/cobra"
)

const (
	flagRegistryVersion = "version"
)

var regShortDesc = map[string]string{
	"list":     "List all registries known to the current ksonnet app.",
	"describe": "Describe a ksonnet registry and the packages it contains",
	"add":      "Add a registry to the current ksonnet app",
}

func init() {
	RootCmd.AddCommand(registryCmd)
	registryCmd.AddCommand(registryDescribeCmd)
	registryCmd.AddCommand(registryAddCmd)

	registryAddCmd.PersistentFlags().String(flagRegistryVersion, "", "Version of the registry to add")
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

		manager, err := metadata.Find(cwd)
		if err != nil {
			return err
		}

		app, err := manager.App()
		if err != nil {
			return err
		}

		appRegistries, err := app.Registries()
		if err != nil {
			return err
		}
		regRef, exists := appRegistries[name]
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

		libs := make([]string, 0, len(reg.Libraries))
		for _, lib := range reg.Libraries {
			libs = append(libs, lib.Path)
		}
		sort.Strings(libs)
		for _, libPath := range libs {
			fmt.Printf("  %s\n", libPath)
		}

		return nil
	},

	Long: `
The ` + "`describe`" + ` command outputs documentation for the ksonnet registry identified
by ` + "`<registry-name>`" + `. Specifically, it displays the following:

1. Registry URI
2. Protocol (e.g. ` + "`github`" + `)
3. List of packages included in the registry

### Related Commands

* ` + "`ks pkg install` " + `— ` + pkgShortDesc["install"] + `

### Syntax
`,
}

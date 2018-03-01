// Copyright 2017 The kubecfg authors
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
	"strings"

	"github.com/ksonnet/ksonnet/metadata"
	"github.com/ksonnet/ksonnet/metadata/parts"
	str "github.com/ksonnet/ksonnet/strings"
	"github.com/spf13/cobra"
)

const (
	flagName = "name"
)

var pkgShortDesc = map[string]string{
	"install":  "Install a package (e.g. extra prototypes) for the current ksonnet app",
	"describe": "Describe a ksonnet package and its contents",
	"list":     "List all packages known (downloaded or not) for the current ksonnet app",
}

var errInvalidSpec = fmt.Errorf("Command 'pkg install' requires a single argument of the form <registry>/<library>@<version>")

func init() {
	RootCmd.AddCommand(pkgCmd)
	pkgCmd.AddCommand(pkgInstallCmd)
	pkgCmd.AddCommand(pkgListCmd)
	pkgCmd.AddCommand(pkgDescribeCmd)
	pkgInstallCmd.PersistentFlags().String(flagName, "", "Name to give the dependency, to use within the ksonnet app")
}

var pkgCmd = &cobra.Command{
	Use:   "pkg",
	Short: `Manage packages and dependencies for the current ksonnet application`,
	Long: `
A ksonnet package contains:

* A set of prototypes (see ` + "`ks prototype --help`" + ` for more info on prototypes), which
generate similar types of components (e.g. ` + "`redis-stateless`" + `, ` + "`redis-persistent`" + `)
* Associated helper libraries that define the prototype parts (e.g. ` + "`redis.libsonnet`" + `)

Packages allow you to easily distribute and reuse code in any ksonnet application.
Packages come from registries, such as Github repositories. (For more info, see
` + "`ks registry --help`" + `).

To be recognized and imported by ksonnet, packages need to follow a specific schema.
See the annotated file tree below, as an example:

` + "```" + `
.
├── README.md                      // Human-readable description of the package
├── parts.yaml                     // Provides metadata about the package
├── prototypes                     // Can be imported and used to generate components
│   ├── redis-all-features.jsonnet
│   ├── redis-persistent.jsonnet
│   └── redis-stateless.jsonnet
└── redis.libsonnet                // Helper library, includes prototype parts
` + "```" + `
---
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("%s is not a valid subcommand\n\n%s", strings.Join(args, " "), cmd.UsageString())
		}
		return fmt.Errorf("Command 'pkg' requires a subcommand\n\n%s", cmd.UsageString())
	},
}

var pkgInstallCmd = &cobra.Command{
	Use:     "install <registry>/<library>@<version>",
	Short:   pkgShortDesc["install"],
	Aliases: []string{"get"},
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Command requires a single argument of the form <registry>/<library>@<version>\n\n%s", cmd.UsageString())
		}

		registry, libID, name, version, err := parseDepSpec(cmd, args[0])
		if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		manager, err := metadata.Find(cwd)
		if err != nil {
			return err
		}

		_, err = manager.CacheDependency(registry, libID, name, version)
		if err != nil {
			return err
		}

		return nil
	},
	Long: `
The ` + "`install`" + ` command caches a ksonnet library locally, and makes it available
for use in the current ksonnet application. Enough info and metadata is recorded in
` + "`app.yaml` " + `that new users can retrieve the dependency after a fresh clone of this app.

The library itself needs to be located in a registry (e.g. Github repo). By default,
ksonnet knows about two registries: *incubator* and *stable*, which are the release
channels for official ksonnet libraries.

### Related Commands

* ` + "`ks pkg list` " + `— ` + pkgShortDesc["list"] + `
* ` + "`ks prototype list` " + `— ` + protoShortDesc["list"] + `
* ` + "`ks registry describe` " + `— ` + regShortDesc["describe"] + `

### Syntax
`,
	Example: `
# Install an nginx dependency, based on the latest branch.
# In a ksonnet source file, this can be referenced as:
#   local nginx = import "incubator/nginx/nginx.libsonnet";
ks pkg install incubator/nginx

# Install an nginx dependency, based on the 'master' branch.
# In a ksonnet source file, this can be referenced as:
#   local nginx = import "incubator/nginx/nginx.libsonnet";
ks pkg install incubator/nginx@master
`,
}

var pkgDescribeCmd = &cobra.Command{
	Use:   "describe [<registry-name>/]<package-name>",
	Short: pkgShortDesc["describe"],
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Command 'pkg describe' requires a package name\n\n%s", cmd.UsageString())
		}

		registryName, libID, err := parsePkgSpec(args[0])
		if err == errInvalidSpec {
			registryName = ""
			libID = args[0]
		} else if err != nil {
			return err
		}

		cwd, err := os.Getwd()
		if err != nil {
			return err
		}

		manager, err := metadata.Find(cwd)
		if err != nil {
			return err
		}

		var lib *parts.Spec
		if registryName == "" {
			lib, err = manager.GetDependency(libID)
			if err != nil {
				return err
			}
		} else {
			lib, err = manager.GetPackage(registryName, libID)
			if err != nil {
				return err
			}
		}

		fmt.Println(`LIBRARY NAME:`)
		fmt.Println(lib.Name)
		fmt.Println()
		fmt.Println(`DESCRIPTION:`)
		fmt.Println(lib.Description)
		fmt.Println()
		fmt.Println(`PROTOTYPES:`)
		for _, proto := range lib.Prototypes {
			fmt.Printf("  %s\n", proto)
		}
		fmt.Println()

		return nil
	},

	Long: `
The ` + "`describe`" + ` command outputs documentation for a package that is available
(e.g. downloaded) in the current ksonnet application. (This must belong to an already
known ` + "`<registry-name>`" + ` like *incubator*). The output includes:

1. The library name
2. A brief description provided by the library authors
3. A list of available prototypes provided by the library

### Related Commands

* ` + "`ks pkg list` " + `— ` + pkgShortDesc["list"] + `
* ` + "`ks prototype describe` " + `— ` + protoShortDesc["describe"] + `
* ` + "`ks generate` " + `— ` + protoShortDesc["use"] + `

### Syntax
`,
}

var pkgListCmd = &cobra.Command{
	Use:   "list",
	Short: pkgShortDesc["list"],
	RunE: func(cmd *cobra.Command, args []string) error {
		const (
			nameHeader      = "NAME"
			registryHeader  = "REGISTRY"
			installedHeader = "INSTALLED"
			installed       = "*"
		)

		if len(args) != 0 {
			return fmt.Errorf("Command 'pkg list' does not take arguments")
		}

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

		rows := [][]string{
			[]string{registryHeader, nameHeader, installedHeader},
			[]string{
				strings.Repeat("=", len(registryHeader)),
				strings.Repeat("=", len(nameHeader)),
				strings.Repeat("=", len(installedHeader))},
		}
		for name := range app.Registries() {
			reg, _, err := manager.GetRegistry(name)
			if err != nil {
				return err
			}

			for libName := range reg.Libraries {
				_, isInstalled := app.Libraries()[libName]
				if isInstalled {
					rows = append(rows, []string{name, libName, installed})
				} else {
					rows = append(rows, []string{name, libName})
				}
			}
		}

		formatted, err := str.PadRows(rows)
		if err != nil {
			return err
		}
		fmt.Print(formatted)
		return nil
	},
	Long: `
The ` + "`list`" + ` command outputs a table that describes all *known* packages (not
necessarily downloaded, but available from existing registries). This includes
the following info:

1. Library name
2. Registry name
3. Installed status — an asterisk indicates 'installed'

### Related Commands

* ` + "`ks pkg install` " + `— ` + pkgShortDesc["install"] + `
* ` + "`ks pkg describe` " + `— ` + pkgShortDesc["describe"] + `
* ` + "`ks registry describe` " + `— ` + regShortDesc["describe"] + `

### Syntax
`,
}

func parsePkgSpec(spec string) (registry, libID string, err error) {
	split := strings.SplitN(spec, "/", 2)
	if len(split) < 2 {
		return "", "", errInvalidSpec
	}
	registry = split[0]
	// Strip off the trailing `@version`.
	libID = strings.SplitN(split[1], "@", 2)[0]
	return
}

func parseDepSpec(cmd *cobra.Command, spec string) (registry, libID, name, version string, err error) {
	registry, libID, err = parsePkgSpec(spec)
	if err != nil {
		return "", "", "", "", err
	}

	split := strings.Split(spec, "@")
	if len(split) > 2 {
		return "", "", "", "", fmt.Errorf("Symbol '@' is only allowed once, at the end of the argument of the form <registry>/<library>@<version>")
	}
	version = ""
	if len(split) == 2 {
		version = split[1]
	}

	name, err = cmd.Flags().GetString(flagName)
	if err != nil {
		return "", "", "", "", err
	} else if name == "" {
		// Get last component, strip off trailing `@<version>`.
		split = strings.Split(spec, "/")
		lastComponent := split[len(split)-1]
		name = strings.SplitN(lastComponent, "@", 2)[0]
	}

	return
}

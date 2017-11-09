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
	"github.com/ksonnet/ksonnet/utils"
	"github.com/spf13/cobra"
)

const (
	flagName = "name"
)

func init() {
	RootCmd.AddCommand(depCmd)
	depCmd.AddCommand(depAddCmd)
	depCmd.AddCommand(depListCmd)
	depAddCmd.PersistentFlags().String(flagName, "", "Name to give the dependency")
}

var depCmd = &cobra.Command{
	Use:   "dep",
	Short: `Manage dependencies for the current ksonnet project`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Command 'dep' requires a subcommand\n\n%s", cmd.UsageString())
	},
}

var depAddCmd = &cobra.Command{
	Use:   "add <registry>/<library>@<version>",
	Short: `Add a dependency to current ksonnet application`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("Command 'dep add' requires a single argument of the form <registry>/<library>@<version>")
		}

		registry, libID, name, version, err := parseDepSpec(cmd, args[0])
		if err != nil {
			return err
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

		_, err = manager.CacheDependency(registry, libID, name, version)
		if err != nil {
			return err
		}

		return nil
	},
	Long: `Cache a ksonnet library locally, and make it available for use in the current
ksonnet project. This particularly means that we record enough information in
'app.yaml' for new users to retrieve the dependency after a fresh clone of the
app repository.

For example, inside a ksonnet application directory, run:

  ks dep get incubator/nginx@v0.1

This can then be referenced in a source file in the ksonnet project:

  local nginx = import "kspkg://nginx";

By default, ksonnet knows about two registries: incubator and stable, which are
the release channels for official ksonnet libraries. Additional registries can
be added with the 'ks registry' command.

Note that multiple versions of the same ksonnet library can be cached and used
in the same project, by explicitly passing in the '--name' flag. For example:

  ks dep get incubator/nginx@v0.1 --name nginxv1
  ks dep get incubator/nginx@v0.2 --name nginxv2

With these commands, a user can 'import "kspkg://nginx1"', and
'import "kspkg://nginx2"' with no conflict.`,
}

var depListCmd = &cobra.Command{
	Use:   "list",
	Short: `Lists information about all dependencies known to the current ksonnet app`,
	RunE: func(cmd *cobra.Command, args []string) error {
		const (
			nameHeader      = "NAME"
			registryHeader  = "REGISTRY"
			installedHeader = "INSTALLED"
			installed       = "*"
		)

		if len(args) != 0 {
			return fmt.Errorf("Command 'dep list' does not take arguments")
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
			[]string{nameHeader, registryHeader, installedHeader},
			[]string{
				strings.Repeat("=", len(nameHeader)),
				strings.Repeat("=", len(registryHeader)),
				strings.Repeat("=", len(installedHeader))},
		}
		for name := range app.Registries {
			reg, _, err := manager.GetRegistry(name)
			if err != nil {
				return err
			}

			for libName := range reg.Libraries {
				_, isInstalled := app.Libraries[libName]
				if isInstalled {
					rows = append(rows, []string{libName, name, installed})
				} else {
					rows = append(rows, []string{libName, name})
				}
			}
		}

		formatted, err := utils.PadRows(rows)
		if err != nil {
			return err
		}
		fmt.Print(formatted)
		return nil
	},
}

func parseDepSpec(cmd *cobra.Command, spec string) (registry, libID, name, version string, err error) {
	split := strings.SplitN(spec, "/", 2)
	if len(split) < 2 {
		return "", "", "", "", fmt.Errorf("Command 'dep add' requires a single argument of the form <registry>/<library>@<version>")
	}
	registry = split[0]
	// Strip off the trailing `@version`.
	libID = strings.SplitN(split[1], "@", 2)[0]

	split = strings.Split(spec, "@")
	if len(split) > 2 {
		return "", "", "", "", fmt.Errorf("Symbol '@' is only allowed once, at the end of the argument of the form <registry>/<library>@<version>")
	}
	version = split[len(split)-1]

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

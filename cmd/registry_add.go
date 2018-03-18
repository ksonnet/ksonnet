// Copyright 2018 The ksonnet authors
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

	"github.com/ksonnet/ksonnet/actions"
	"github.com/spf13/cobra"
)

var registryAddCmd = &cobra.Command{
	Use:   "add <registry-name> <registry-uri>",
	Short: regShortDesc["add"],
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()

		if len(args) != 2 {
			return fmt.Errorf("Command 'registry add' takes two arguments, which is the name and the repository address of the registry to add")
		}

		name := args[0]
		uri := args[1]

		version, err := flags.GetString(flagRegistryVersion)
		if err != nil {
			return err
		}

		return actions.RunRegistryAdd(ka, name, uri, version)
	},

	Long: `
The ` + "`add`" + ` command allows custom registries to be added to your ksonnet app,
provided that their file structures follow the appropriate schema. *You can look
at the ` + "`incubator`" + ` repo (https://github.com/ksonnet/parts/tree/master/incubator)
as an example.*

A registry is uniquely identified by its:

1. Name (e.g. ` + "`incubator`" + `)
2. Version (e.g. ` + "`master`" + `)

Currently, only registries supporting the **GitHub protocol** can be added.

During creation, all registries must specify a unique name and URI where the
registry lives. Optionally, a version can be provided (e.g. the *Github branch
name*). If a version is not specified, it will default to ` + "`latest`" + `.


### Related Commands

* ` + "`ks registry list` " + `— ` + regShortDesc["list"] + `

### Syntax
`,
	Example: `# Add a registry with the name 'databases' at the uri 'github.com/example'
ks registry add databases github.com/example

# Add a registry with the name 'databases' at the uri
# 'github.com/example/tree/master/reg' and the version (branch name) 0.0.1
# NOTE that "0.0.1" overrides the branch name in the URI ("master")
ks registry add databases github.com/example/tree/master/reg --version=0.0.1`,
}

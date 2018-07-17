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

package clicmd

import (
	"fmt"

	"github.com/spf13/viper"

	"github.com/ksonnet/ksonnet/pkg/actions"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/spf13/cobra"
)

const (
	vRegistryAddOverride = "registry-add-override"
)

var (
	registryAddLong = `
The ` + "`add`" + ` command allows custom registries to be added to your ksonnet app,
provided that their file structures follow the appropriate schema. *You can look
at the ` + "`incubator`" + ` repo (https://github.com/ksonnet/parts/tree/master/incubator)
as an example.*

A registry is given a string identifier, which must be unique within a ksonnet application.

There are three supported registry protocols: **github**, **fs**, and **Helm**.

GitHub registries expect a path in a GitHub repository, and filesystem based
registries expect a path on the local filesystem.

During creation, all registries must specify a unique name and URI where the
registry lives. GitHub registries can specify a commit, tag, or branch to follow as part of the URI.

Registries can be overridden with ` + "`--override`" + `.  Overridden registries
are stored in ` + "`app.override.yaml`" + ` and can be safely ignored using your
SCM configuration.

### Related Commands

* ` + "`ks registry list` " + `— ` + regShortDesc["list"] + `

### Syntax
`
	registryAddExample = `# Add a registry with the name 'databases' at the uri 'github.com/example'
ks registry add databases github.com/example

# Add a registry with the name 'databases' at the uri
# 'github.com/org/example/tree/0.0.1/registry' (0.0.1 is the branch name)
ks registry add databases github.com/org/example/tree/0.0.1/registry

# Add a registry with a Helm Charts Repository uri
ks registry add helm-stable https://kubernetes-charts.storage.googleapis.com`
)

func newRegistryAddCmd(a app.App) *cobra.Command {
	registryAddCmd := &cobra.Command{
		Use:     "add <registry-name> <registry-uri>",
		Short:   regShortDesc["add"],
		Long:    registryAddLong,
		Example: registryAddExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 2 {
				return fmt.Errorf("Command 'registry add' takes two arguments, which is the name and the repository address of the registry to add")
			}

			m := map[string]interface{}{
				actions.OptionApp:      a,
				actions.OptionName:     args[0],
				actions.OptionURI:      args[1],
				actions.OptionOverride: viper.GetBool(vRegistryAddOverride),
			}

			return runAction(actionRegistryAdd, m)
		},
	}

	registryAddCmd.Flags().BoolP(flagOverride, shortOverride, false, "Store in override configuration")
	viper.BindPFlag(vRegistryAddOverride, registryAddCmd.Flags().Lookup(flagOverride))

	return registryAddCmd
}

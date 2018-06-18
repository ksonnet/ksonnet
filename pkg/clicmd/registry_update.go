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
	"github.com/ksonnet/ksonnet/pkg/actions"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	vRegistryUpdateVersion = "registry-update-version"
)

var (
	registryUpdateLong = `
The ` + "`update`" + ` command updates a set of configured registries in your ksonnet app.
Unless a specific version is specified with ` + "`--version`" + `, update will attempt to
fetch the lastest registry version matching the configured floating version specifer.

With ` + "`--version`" + `, a specific version specifier (floating or concrete) can be set.

### Syntax
`
	registryUpdatExample = `# Update *all* registries to their latest matching versions
ks registry update

# Update a registry with the name 'databases' to version 0.0.2
ks registry update databases --version=0.0.1`
)

func newRegistryUpdateCmd(a app.App) *cobra.Command {
	registryUpdateCmd := &cobra.Command{
		Use:     "update [registry-name]",
		Short:   regShortDesc["update"],
		Long:    registryUpdateLong,
		Example: registryUpdatExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			var registryName string
			var registryVersion string

			switch {
			// case order matters
			case len(args) >= 2:
				registryVersion = args[1]
				registryName = args[0]
			case len(args) >= 1:
				registryName = args[0]
			default:
			}

			m := map[string]interface{}{
				actions.OptionApp:     a,
				actions.OptionName:    registryName,
				actions.OptionVersion: registryVersion,

				// TODO: decide what to do with these
				//actions.OptionVersion:  viper.GetString(vRegistryAddVersion), TODO decide if this is positional or not
				// actions.OptionOverride: viper.GetBool(vRegistryAddOverride),
			}

			return runAction(actionRegistryUpdate, m)
		},
	}

	registryUpdateCmd.Flags().String(flagVersion, "", "Version to update registry to")
	viper.BindPFlag(vRegistryUpdateVersion, registryUpdateCmd.Flags().Lookup(flagVersion))

	// TODO: decide what to do with these
	// registryUpdateCmd.Flags().BoolP(flagOverride, shortOverride, false, "Store in override configuration")
	// viper.BindPFlag(vRegistryAddOverride, registryAddCmd.Flags().Lookup(flagOverride))

	return registryUpdateCmd
}

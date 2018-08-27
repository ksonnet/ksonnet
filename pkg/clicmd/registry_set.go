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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	vRegistrySetURI = "registry-set-uri"
)

var (
	registrySetLong = `
The ` + "`set`" + ` command sets configuration parameters on a configured registry.
The following parameters can be set:

* --uri: The uri a registry points to. For GitHub-based registries, this can be used to select a specific branch.
`
	registrySetExample = `
	# Set the incubator registry to the experimental branch:
	ks registry set incubator --uri https://github.com/ksonnet/parts/tree/experimental/incubator
`
)

func newRegistrySetCmd() *cobra.Command {
	var registrySetCmd = &cobra.Command{
		Use:     "set [registry-name]",
		Short:   regShortDesc["set"],
		Long:    registrySetLong,
		Example: registrySetExample,
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			registryName := args[0] // len(args) was verified

			m := map[string]interface{}{
				actions.OptionName: registryName,
				actions.OptionURI:  viper.GetString(vRegistrySetURI),
			}

			return runAction(actionRegistrySet, m)
		},
	}

	flagURI := "uri"
	registrySetCmd.Flags().String(flagURI, "", "URI to configure the registry")
	viper.BindPFlag(vRegistrySetURI, registrySetCmd.Flags().Lookup(flagURI))

	return registrySetCmd
}

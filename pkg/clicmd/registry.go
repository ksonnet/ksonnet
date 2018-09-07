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

package clicmd

import (
	"github.com/spf13/cobra"
)

var (
	regShortDesc = map[string]string{
		"list":     "List all registries known to the current ksonnet app",
		"describe": "Describe a ksonnet registry and the packages it contains",
		"add":      "Add a registry to the current ksonnet app",
		"set":      "Set configuration options for registry",
	}
	registryLong = `
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
`
)

func newRegistryCmd() *cobra.Command {
	registryCmd := &cobra.Command{
		Use:   "registry",
		Short: `Manage registries for current project`,
		Long:  registryLong,
	}

	registryCmd.AddCommand(newRegistryAddCmd())
	registryCmd.AddCommand(newRegistryDescribeCmd())
	registryCmd.AddCommand(newRegistryListCmd())
	registryCmd.AddCommand(newRegistrySetCmd())

	return registryCmd
}

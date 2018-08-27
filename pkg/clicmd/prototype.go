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

package clicmd

import (
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
)

var (
	protoShortDesc = map[string]string{
		"list":     "List all locally available ksonnet prototypes",
		"describe": "See more info about a prototype's output and usage",
		"preview":  "Preview a prototype's output without creating a component (stdout)",
		"search":   "Search for a prototype",
		"use":      "Use the specified prototype to generate a component manifest",
	}
	protoLong = `
Use the` + " `prototype` " + `subcommands to manage, inspect, instantiate, and get
examples for ksonnet prototypes.

Prototypes are pre-written but incomplete Kubernetes manifests, with "holes"
(parameters) that can be filled in with the ksonnet CLI or manually. For example,
the prototype` + " `io.ksonnet.pkg.single-port-deployment` " + `requires a name and image,
and the ksonnet CLI can expand this into a fully-formed 'Deployment' object.

These complete manifests are output into your ` + "`components/`" + ` directory. In other
words, prototypes provide the basis for the **components** of your app. You can
use prototypes to autogenerate boilerplate code and focus on customizing them
for your use case.

----
`
)

func newPrototypeCmd(fs afero.Fs) *cobra.Command {
	prototypeCmd := &cobra.Command{
		Use:   "prototype",
		Short: `Instantiate, inspect, and get examples for ksonnet prototypes`,
		Long:  protoLong,
	}

	prototypeCmd.AddCommand(newPrototypeDescribeCmd())
	prototypeCmd.AddCommand(newPrototypeListCmd())
	prototypeCmd.AddCommand(newPrototypePreviewCmd())
	prototypeCmd.AddCommand(newPrototypeSearchCmd())
	prototypeCmd.AddCommand(newPrototypeUseCmd(fs))

	return prototypeCmd
}

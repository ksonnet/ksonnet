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
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/spf13/cobra"
)

const (
	flagName = "name"
)

var (
	pkgShortDesc = map[string]string{
		"install":  "Install a package (e.g. extra prototypes) for the current ksonnet app",
		"remove":   "Remove a package from the app or environment scope",
		"describe": "Describe a ksonnet package and its contents",
		"list":     "List all packages known (downloaded or not) for the current ksonnet app",
	}
	pkgLong = `
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
`
)

func newPkgCmd(a app.App) *cobra.Command {
	pkgCmd := &cobra.Command{
		Use:   "pkg",
		Short: `Manage packages and dependencies for the current ksonnet application`,
		Long:  pkgLong,
	}

	pkgCmd.AddCommand(newPkgListCmd(a))
	pkgCmd.AddCommand(newPkgInstallCmd(a))
	pkgCmd.AddCommand(newPkgDescribeCmd(a))
	pkgCmd.AddCommand(newPkgRemoveCmd(a))

	return pkgCmd
}

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
	"fmt"
	"strings"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/client"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	flagEnvName      = "name"
	flagEnvServer    = "server"
	flagEnvNamespace = "namespace"
	flagEnvContext   = "context"
)

var (
	envShortDesc = map[string]string{
		"add":     "Add a new environment to a ksonnet application",
		"current": "Sets the current environment",
		"list":    "List all environments in a ksonnet application",
		"rm":      "Delete an environment from a ksonnet application",
		"set":     "Set environment-specific fields (name, namespace, server)",
		"update":  "Updates the libs for an environment",
	}

	envLong = `
An environment is a deployment target for your ksonnet app and its constituent
components. You can use ksonnet to deploy a given app to *multiple* environments,
such as ` + "`dev`" + ` and ` + "`prod`" + `.

Intuitively, an environment acts as a sort of "named cluster", similar to a
Kubernetes context. (Running ` + "`ks env add --help`" + ` provides more detail
about the fields that you need to create an environment).

**All of this environment info is cached in local files**. Metadata such as k8s version, API server address, and namespace are found in ` + "`app.yaml`. " + `Environments are
represented as a hierarchy in the ` + "`environments/`" + ` directory of a ksonnet app, like
'default' and 'us-west/staging' in the example below.

` + "```" + `
├── environments
│   ├── base.libsonnet
│   ├── default
│   │   ├── globals.libsonnet        // Default generated environment ('ks init')
│   │   ├── main.jsonnet
│   │   └── params.libsonnet
│   └── us-west
│       └── staging                  // Example of user-generated env ('ks env add')
│           ├── globals.libsonnet
│           ├── main.jsonnet         // Main file that imports all components (expanded on apply, delete, etc). Add environment-specific logic here.
│           └── params.libsonnet     // Customize components *per-environment* here.
` + "```" + `
----
`
)

func newEnvCmd(a app.App) *cobra.Command {
	envCmd := &cobra.Command{
		Use:   "env",
		Short: `Manage ksonnet environments`,
		Long:  envLong,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 0 {
				return fmt.Errorf("%s is not a valid subcommand\n\n%s", strings.Join(args, " "), cmd.UsageString())
			}
			return fmt.Errorf("Command 'env' requires a subcommand\n\n%s", cmd.UsageString())
		},
	}

	envCmd.AddCommand(newEnvAddCmd(a))
	envCmd.AddCommand(newEnvCurrentCmd(a))
	envCmd.AddCommand(newEnvDescribeCmd(a))
	envCmd.AddCommand(newEnvListCmd(a))
	envCmd.AddCommand(newEnvRmCmd(a))
	envCmd.AddCommand(newEnvSetCmd(a))
	envCmd.AddCommand(newEnvTargetsCmd(a))
	envCmd.AddCommand(newEnvUpdateCmd(a))

	return envCmd

}

func commonEnvFlags(flags *pflag.FlagSet) (server, namespace, context string, err error) {
	server, err = flags.GetString(flagEnvServer)
	if err != nil {
		return "", "", "", err
	}

	namespace, err = flags.GetString(flagEnvNamespace)
	if err != nil {
		return "", "", "", err
	}

	context, err = flags.GetString(flagEnvContext)
	if err != nil {
		return "", "", "", err
	}

	if flags.Changed(flagEnvContext) && flags.Changed(flagEnvServer) {
		return "", "", "", fmt.Errorf("flags '%s' and '%s' are mutually exclusive, because '%s' has a server. Try setting '%s', '%s' to the desired values",
			flagEnvContext, flagEnvServer, flagEnvContext, flagEnvServer, flagEnvNamespace)
	}

	return server, namespace, context, nil
}

func resolveEnvFlags(flags *pflag.FlagSet, config *client.Config) (string, string, error) {
	defaultNamespace := "default"

	server, envNs, context, err := commonEnvFlags(flags)
	if err != nil {
		return "", "", err
	}

	var ctxNs string
	if server == "" {
		// server is not provided -- use the context.
		server, ctxNs, err = config.ResolveContext(context)
		if err != nil {
			return "", "", err
		}
	}

	ns := defaultNamespace
	if envNs != "" {
		ns = envNs
	} else if ctxNs != "" {
		ns = ctxNs
	}

	return server, ns, nil
}

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

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ksonnet/ksonnet/client"
	"github.com/ksonnet/ksonnet/metadata"
	"github.com/ksonnet/ksonnet/pkg/kubecfg"
)

const (
	flagEnvName      = "name"
	flagEnvServer    = "server"
	flagEnvNamespace = "namespace"
	flagEnvContext   = "context"
)

var (
	envClientConfig *client.Config
	envShortDesc    = map[string]string{
		"add":  "Add a new environment to a ksonnet application",
		"list": "List all environments in a ksonnet application",
		"rm":   "Delete an environment from a ksonnet application",
		"set":  "Set environment-specific fields (name, namespace, server)",
	}
)

func init() {
	RootCmd.AddCommand(envCmd)
	envClientConfig = client.NewDefaultClientConfig()
	envClientConfig.BindClientGoFlags(envCmd)

	envCmd.AddCommand(envAddCmd)
	envCmd.AddCommand(envRmCmd)
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envSetCmd)

	// TODO: We need to make this default to checking the `kubeconfig` file.
	envAddCmd.PersistentFlags().String(flagAPISpec, "version:v1.7.0",
		"Manually specify API version from OpenAPI schema, cluster, or Kubernetes version")

	envSetCmd.PersistentFlags().String(flagEnvName, "",
		"Name used to uniquely identify the environment. Must not already exist within the ksonnet app")
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: `Manage ksonnet environments`,
	Long: `
An environment is a deployment target for your ksonnet app and its constituent
components. You can use ksonnet to deploy a given app to *multiple* environments,
such as ` + "`dev`" + ` and ` + "`prod`" + `.

Intuitively, an environment acts as a sort of "named cluster", similar to a
Kubernetes context. (Running ` + "`ks env add --help`" + ` provides more detail
about the fields that you need to create an environment).

**All of this environment info is cached in local files**. Environments are
represented as a hierarchy in the ` + "`environments/`" + ` directory of a ksonnet app, like
'default' and 'us-west/staging' in the example below.

` + "```" + `
├── environments
│   ├── base.libsonnet
│   ├── default                      // Default generated environment ('ks init')
│   │   ├── .metadata
│   │   │   ├── k.libsonnet
│   │   │   ├── k8s.libsonnet
│   │   │   └── swagger.json
│   │   ├── main.jsonnet
│   │   ├── params.libsonnet
│   │   └── spec.json
│   └── us-west
│       └── staging                  // Example of user-generated env ('ks env add')
│           ├── .metadata
│           │   ├── k.libsonnet      // Jsonnet library with Kubernetes-compatible types and definitions
│           │   ├── k8s.libsonnet
│           │   └── swagger.json
│           ├── main.libsonnet       // Main file that imports all components (expanded on apply, delete, etc). Add environment-specific logic here.
│           ├── params.libsonnet     // Customize components *per-environment* here.
│           └── spec.json            // Contains the environment's API server address and namespace
` + "```" + `
----
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("%s is not a valid subcommand\n\n%s", strings.Join(args, " "), cmd.UsageString())
		}
		return fmt.Errorf("Command 'env' requires a subcommand\n\n%s", cmd.UsageString())
	},
}

var envAddCmd = &cobra.Command{
	Use:   "add <env-name>",
	Short: envShortDesc["add"],
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		if len(args) != 1 {
			return fmt.Errorf("'env add' takes exactly one argument, which is the name of the environment")
		}

		name := args[0]

		server, namespace, err := resolveEnvFlags(flags)
		if err != nil {
			return err
		}

		appDir, err := os.Getwd()
		if err != nil {
			return err
		}

		manager, err := metadata.Find(appDir)
		if err != nil {
			return err
		}

		specFlag, err := flags.GetString(flagAPISpec)
		if err != nil {
			return err
		}
		if specFlag == "" {
			specFlag = envClientConfig.GetAPISpec(server)
		}

		c, err := kubecfg.NewEnvAddCmd(name, server, namespace, specFlag, manager)
		if err != nil {
			return err
		}

		return c.Run()
	},

	Long: `
The ` + "`add`" + ` command creates a new environment (specifically for the ksonnet app
whose directory it's executed in). This environment is cached with the following
info:

1. **Name** — A string used to uniquely identify the environment.
2. **Server** — The address and port of a Kubernetes API server (i.e. cluster).
3. **Namespace**  — A Kubernetes namespace. *Must already exist on the cluster.*
4. **Kubernetes API Version**  — Used to generate a library with compatible type defs.

(1) is mandatory. (2) and (3) can be inferred from $KUBECONFIG, *or* from the
` + "`--kubeconfig`" + ` or ` + "`--context`" + ` flags. Otherwise, (2), (3), and (4) can all be
specified by individual flags. Unless otherwise specified, (4) defaults to the
latest Kubernetes version that ksonnet supports.

Note that an environment *DOES NOT* contain user-specific data such as private keys.

### Related Commands

* ` + "`ks env list` " + `— ` + protoShortDesc["list"] + `
* ` + "`ks env rm` " + `— ` + protoShortDesc["rm"] + `
* ` + "`ks env set` " + `— ` + protoShortDesc["set"] + `
* ` + "`ks param set` " + `— ` + paramShortDesc["set"] + `
* ` + "`ks apply` " + `— ` + applyShortDesc + `

### Syntax
`,
	Example: `
# Initialize a new environment, called "staging". No flags are set, so 'server'
# and 'namespace' info are pulled from the file specified by $KUBECONFIG.
# 'version' defaults to the latest that ksonnet supports.
ks env add us-west/staging

# Initialize a new environment called "us-west/staging" with the pre-existing
# namespace 'staging'. 'version' is specified, so the OpenAPI spec from the
# Kubernetes v1.7.1 build is used to generate the helper library 'ksonnet-lib'.
#
# NOTE: "us-west/staging" indicates a hierarchical structure, so the env-specific
# files here are saved in "<ksonnet-app-root>/environments/us-west/staging".
ks env add us-west/staging --api-spec=version:v1.7.1 --namespace=staging

# Initialize a new environment "my-env" using the "dev" context in your current
# kubeconfig file ($KUBECONFIG).
ks env add my-env --context=dev

# Initialize a new environment "prod" using the address of a cluster's Kubernetes
# API server.
ks env add prod --server=https://ksonnet-1.us-west.elb.amazonaws.com`,
}

var envRmCmd = &cobra.Command{
	Use:   "rm <env-name>",
	Short: envShortDesc["rm"],
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("'env rm' takes a single argument, that is the name of the environment")
		}

		envName := args[0]

		appDir, err := os.Getwd()
		if err != nil {
			return err
		}

		manager, err := metadata.Find(appDir)
		if err != nil {
			return err
		}

		c, err := kubecfg.NewEnvRmCmd(envName, manager)
		if err != nil {
			return err
		}

		return c.Run()
	},
	Long: `
The ` + "`rm`" + ` command deletes an environment from a ksonnet application. This is
the same as removing the ` + "`<env-name>`" + ` environment directory and all files
contained. All empty parent directories are also subsequently deleted.

NOTE: This does *NOT* delete the components running in ` + "`<env-name>`" + `. To do that, you
need to use the ` + "`ks delete`" + ` command.

### Related Commands

* ` + "`ks env list` " + `— ` + protoShortDesc["list"] + `
* ` + "`ks env add` " + `— ` + protoShortDesc["add"] + `
* ` + "`ks env set` " + `— ` + protoShortDesc["set"] + `
* ` + "`ks delete` " + `— ` + `Delete all the app components running in an environment (cluster)` + `

### Syntax
`,
	Example: `
# Remove the directory 'environments/us-west/staging' and all of its contents.
# This will also remove the parent directory 'us-west' if it is empty.
ks env rm us-west/staging`,
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: envShortDesc["list"],
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("'env list' takes zero arguments")
		}

		appDir, err := os.Getwd()
		if err != nil {
			return err
		}

		manager, err := metadata.Find(appDir)
		if err != nil {
			return err
		}

		c, err := kubecfg.NewEnvListCmd(manager)
		if err != nil {
			return err
		}

		return c.Run(cmd.OutOrStdout())
	}, Long: `
The ` + "`list`" + ` command lists all of the available environments for the
current ksonnet app. Specifically, this will display the (1) *name*,
(2) *server*, and (3) *namespace* of each environment.

### Related Commands

* ` + "`ks env add` " + `— ` + envShortDesc["add"] + `
* ` + "`ks env set` " + `— ` + envShortDesc["set"] + `
* ` + "`ks env rm` " + `— ` + envShortDesc["rm"] + `

### Syntax
`,
}

var envSetCmd = &cobra.Command{
	Use:   "set <env-name>",
	Short: envShortDesc["set"],
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		if len(args) != 1 {
			return fmt.Errorf("'env set' takes a single argument, that is the name of the environment")
		}

		originalName := args[0]

		appDir, err := os.Getwd()
		if err != nil {
			return err
		}

		manager, err := metadata.Find(appDir)
		if err != nil {
			return err
		}

		name, err := flags.GetString(flagEnvName)
		if err != nil {
			return err
		}

		c, err := kubecfg.NewEnvSetCmd(originalName, name, manager)
		if err != nil {
			return err
		}

		return c.Run()
	},
	Long: `
The ` + "`set`" + ` command lets you change the fields of an existing environment.
You can currently only update your environment's name.

Note that changing the name of an environment will also update the corresponding
directory structure in ` + "`environments/`" + `.

### Related Commands

* ` + "`ks env list` " + `— ` + envShortDesc["list"] + `

### Syntax
`,
	Example: `#Update the name of the environment 'us-west/staging'.
# Updating the name will update the directory structure in 'environments/'.
ks env set us-west/staging --name=us-east/staging`,
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

func resolveEnvFlags(flags *pflag.FlagSet) (string, string, error) {
	defaultNamespace := "default"

	server, ns, context, err := commonEnvFlags(flags)
	if err != nil {
		return "", "", err
	}

	if server == "" {
		// server is not provided -- use the context.
		server, defaultNamespace, err = envClientConfig.ResolveContext(context)
		if err != nil {
			return "", "", err
		}
	}

	namespace := defaultNamespace
	if ns != "" {
		namespace = ns
	}

	return server, namespace, nil
}

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

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/ksonnet/ksonnet/metadata"
	"github.com/ksonnet/ksonnet/pkg/kubecfg"
)

const (
	flagEnvName      = "name"
	flagEnvServer    = "server"
	flagEnvNamespace = "namespace"
	flagEnvContext   = "context"

	defaultNamespace = "default"
)

func init() {
	RootCmd.AddCommand(envCmd)
	bindClientGoFlags(envCmd)

	envCmd.AddCommand(envAddCmd)
	envCmd.AddCommand(envRmCmd)
	envCmd.AddCommand(envListCmd)
	envCmd.AddCommand(envSetCmd)

	// TODO: We need to make this default to checking the `kubeconfig` file.
	envAddCmd.PersistentFlags().String(flagAPISpec, "version:v1.7.0",
		"Manually specified Kubernetes API version. The corresponding OpenAPI spec is used to generate ksonnet's Kubernetes libraries")

	envSetCmd.PersistentFlags().String(flagEnvName, "",
		"Name used to uniquely identify the environment. Must not already exist within the ksonnet app")
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: `Manage ksonnet environments`,
	Long: `An environment acts as a sort of "named cluster context", and points at a
combination of a cluster's (1) API server URI, and (2) Kubernetes namespace. *Note that
the namespace must already exist on your cluster.* This environment definition
allows commands like` + " `ks apply dev` " + `to make changes to the specified
"dev cluster".

An environment also caches additional data about the cluster it points to, so
that it has:
1. The right version of the helper Kubernetes library (ksonnet-lib),
2. The appropriate schema to later validate manifests against

Note that an environment *DOES NOT* contain user-specific data such as private keys.

Environments are represented as a hierarchy in the 'environments/' directory of a
ksonnet application. For example, in the example below, there are two
environments: "default" and "us-west/staging". Each contains a cached version of
` + " `ksonnet-lib` " + `, and a` + " `spec.json` " + `that contains the server
and server cert that uniquely identifies the cluster.

├── environments
│   ├── base.libsonnet
│   ├── default                      // default generated environment
│   │   ├── .metadata
│   │   │   ├── k.libsonnet
│   │   │   ├── k8s.libsonnet
│   │   │   └── swagger.json
│   │   ├── default.jsonnet
│   │   ├── params.libsonnet
│   │   └── spec.json
│   └── us-west
│       └── staging                  // example of user-generated env
│           ├── .metadata
│           │   ├── k.libsonnet
│           │   ├── k8s.libsonnet
│           │   └── swagger.json
│           ├── params.libsonnet
│           ├── spec.json           //This will contain the enviornment's API server address and namespace
│           └── staging.jsonnet
...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return fmt.Errorf("Command 'env' requires a subcommand\n\n%s", cmd.UsageString())
	},
}

var envAddCmd = &cobra.Command{
	Use:   "add <env-name>",
	Short: "Add a new environment to a ksonnet application",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()
		if len(args) != 1 {
			return fmt.Errorf("'env add' takes exactly one argument, which is the name of the environment")
		}

		name := args[0]

		server, namespace, context, err := commonEnvFlags(flags)
		if err != nil {
			return err
		}

		if len(namespace) == 0 {
			namespace = defaultNamespace
		}

		if len(server) == 0 {
			// If server is not provided, use the provided context.
			// If context is also not provided, use the current context.
			var ctx *string
			if len(context) != 0 {
				ctx = &context
			}

			var ns string
			server, ns, err = resolveContext(ctx)
			if err != nil {
				return err
			}

			// If namespace is not provided, use the default namespace provided in the context.
			if !flags.Changed(flagEnvNamespace) {
				namespace = ns
			}
		}

		appDir, err := os.Getwd()
		if err != nil {
			return err
		}
		appRoot := metadata.AbsPath(appDir)

		manager, err := metadata.Find(appRoot)
		if err != nil {
			return err
		}

		specFlag, err := flags.GetString(flagAPISpec)
		if err != nil {
			return err
		}

		c, err := kubecfg.NewEnvAddCmd(name, server, namespace, specFlag, manager)
		if err != nil {
			return err
		}

		return c.Run()
	},

	Long: `Add a new environment to a ksonnet application. Names are restricted to not
include punctuation, so names like` + " `../foo` " + `are not allowed.

An environment acts as a sort of "named cluster context". For more information on
what an environment is and what the autogenerated directory structure looks like,
run` + " `ks env --help` " + `or see the official website documentation.`,
	Example: `# Initialize a new staging environment at "us-west", using the current
# 'kubeconfig' context. This autogenerates a subdirectory rooted at
# 'environments/us-west'.
ks env add us-west/staging

# Initialize a new staging environment at "us-west" with the namespace "staging", using
# the OpenAPI specification generated in the Kubernetes v1.7.1 build to generate "ksonnet-lib".
ks env add us-west/staging --api-spec=version:v1.7.1 --namespace=staging

# Initialize a new environment using the "dev" context in your current 'kubeconfig' file.
ks env add my-env --context=dev

# Initialize a new environment using the address of a cluster's Kubernetes API server.
ks env add my-env --server=https://ksonnet-1.us-west.elb.amazonaws.com`,
}

var envRmCmd = &cobra.Command{
	Use:   "rm <env-name>",
	Short: "Delete an environment from a ksonnet application",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 1 {
			return fmt.Errorf("'env rm' takes a single argument, that is the name of the environment")
		}

		envName := args[0]

		appDir, err := os.Getwd()
		if err != nil {
			return err
		}
		appRoot := metadata.AbsPath(appDir)

		manager, err := metadata.Find(appRoot)
		if err != nil {
			return err
		}

		c, err := kubecfg.NewEnvRmCmd(envName, manager)
		if err != nil {
			return err
		}

		return c.Run()
	},
	Long: `Delete an environment from a ksonnet application. This is the same
as removing the <env-name> environment directory and all files contained. All empty
parent directories are also subsequently deleted.`,
	Example: `# Remove the directory "environments/us-west/staging", as well as the
# parent directory "us-west" (if it is empty).
ks env rm us-west/staging`,
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environments in a ksonnet application",
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("'env list' takes zero arguments")
		}

		appDir, err := os.Getwd()
		if err != nil {
			return err
		}
		appRoot := metadata.AbsPath(appDir)

		manager, err := metadata.Find(appRoot)
		if err != nil {
			return err
		}

		c, err := kubecfg.NewEnvListCmd(manager)
		if err != nil {
			return err
		}

		return c.Run(cmd.OutOrStdout())
	}, Long: `List all environments in a ksonnet application. This will display the
(1) name, (2) server, and (3) namespace of each environment within the ksonnet
application.`,
}

var envSetCmd = &cobra.Command{
	Use:   "set <env-name>",
	Short: "Set environment fields such as the name, server, and namespace.",
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
		appRoot := metadata.AbsPath(appDir)

		manager, err := metadata.Find(appRoot)
		if err != nil {
			return err
		}

		name, err := flags.GetString(flagEnvName)
		if err != nil {
			return err
		}

		server, namespace, context, err := commonEnvFlags(flags)
		if err != nil {
			return err
		}

		// If namespace flag is provided but without a value, use the default namespace.
		if flags.Changed(flagEnvNamespace) && len(namespace) == 0 {
			namespace = defaultNamespace
		}

		if len(context) != 0 {
			server, _, err = resolveContext(&context)
			if err != nil {
				return err
			}
		}

		c, err := kubecfg.NewEnvSetCmd(originalName, name, server, namespace, manager)
		if err != nil {
			return err
		}

		return c.Run()
	},
	Long: `Set environment fields such as the name, namespace, and server. Changing
the name of an environment will also update the directory structure in
'environments/'. However, note that resources on the original cluster will not be
automatically cleaned up.`,
	Example: `# Update the API server address of the environment "us-west/staging".
ks env set us-west/staging --server=http://example.com

# Update the namespace of the environment "us-west/staging" to "staging".
ks env set us-west/staging --namespace=staging

# Update both the name and the server of the environment "us-west/staging".
ks env set us-west/staging --server=http://example.com --name=us-east/staging

# Update the API server address of the environment "us-west/staging", based on
# server URI specified in the "staging-west" context of your current 'kubeconfig'
# file.
ks env set us-west/staging --context=staging-west`,
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

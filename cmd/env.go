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

	"github.com/ksonnet/ksonnet/metadata"
	"github.com/ksonnet/ksonnet/pkg/kubecfg"
)

const (
	flagEnvName      = "name"
	flagEnvServer    = "server"
	flagEnvNamespace = "namespace"
	flagEnvContext   = "context"
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
		"Manually specify API version from OpenAPI schema, cluster, or Kubernetes version")

	envSetCmd.PersistentFlags().String(flagEnvName, "",
		"Specify name to rename environment to. Name must not already exist")
}

var envCmd = &cobra.Command{
	Use:   "env",
	Short: `Manage ksonnet environments`,
	Long: `An environment acts as a sort of "named cluster", allowing for commands like
` + " `ks apply dev` " + `, which applies the ksonnet application to the 'dev cluster'.
Additionally, environments allow users to cache data about the cluster it points
to, including data needed to run 'verify', and a version of ksonnet-lib that is
generated based on the flags the API server was started with (e.g., RBAC enabled
or not).

An environment contains no user-specific data (such as the private key
often contained in a kubeconfig file), and

Environments are represented as a hierarchy in the 'environments' directory of a
ksonnet application. For example, in the example below, there are two
environments: 'default' and 'us-west/staging'. Each contains a cached version of
` + " `ksonnet-lib` " + `, and a` + " `spec.json` " + `that contains the server and server cert that
uniquely identifies the cluster.

    environments/
      default/           [Default generated environment]
        .metadata/
          k.libsonnet
          k8s.libsonnet
          swagger.json
        spec.json
		default.jsonnet
        params.libsonnet		
      us-west/
        staging/         [Example of user-generated env]
          .metadata/
            k.libsonnet
            k8s.libsonnet
            swagger.json
          spec.json      [This will contain the API server address of the environment and other environment metadata]
		  staging.jsonnet
          params.libsonnet`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) != 0 {
			return fmt.Errorf("%s is not a valid subcommand\n\n%s", strings.Join(args, " "), cmd.UsageString())
		}
		return fmt.Errorf("Command 'env' requires a subcommand\n\n%s", cmd.UsageString())
	},
}

var envAddCmd = &cobra.Command{
	Use:   "add <env-name>",
	Short: "Add a new environment to a ksonnet project",
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

	Long: `Add a new environment to a ksonnet project. Names are restricted to not
include punctuation, so names like` + " `../foo` " + `are not allowed.

An environment acts as a sort of "named cluster", allowing for commands like
` + " `ks apply dev` " + `, which applies the ksonnet application to the "dev cluster".
For more information on what an environment is and how they work, run` + " `ks help env` " + `.

Environments are represented as a hierarchy in the 'environments' directory of a
ksonnet application, and hence` + " `ks env add` " + `will add to this directory structure.
For example, in the example below, there are two environments: 'default' and
'us-west/staging'.` + " `ks env add` " + `will add a similar directory to this environment.

    environments/
      default/           [Default generated environment]
        .metadata/
          k.libsonnet
          k8s.libsonnet
          swagger.json
        spec.json
		default.jsonnet
        params.libsonnet
      us-west/
        staging/         [Example of user-generated env]
          .metadata/
            k.libsonnet
            k8s.libsonnet
            swagger.json
          spec.json      [This will contain the API server address of the environment and other environment metadata],
		  staging.jsonnet
          params.libsonnet`,
	Example: `# Initialize a new staging environment at 'us-west'.
# The environment will be setup using the current context in your kubecfg file. The directory
# structure rooted at 'us-west' in the documentation above will be generated.
ks env add us-west/staging

# Initialize a new staging environment at 'us-west' with the namespace 'staging', using
# the OpenAPI specification generated in the Kubernetes v1.7.1 build to generate 'ksonnet-lib'.
ks env add us-west/staging --api-spec=version:v1.7.1 --namespace=staging

# Initialize a new environment using the 'dev' context in your kubeconfig file.
ks env add my-env --context=dev

# Initialize a new environment using a server address.
ks env add my-env --server=https://ksonnet-1.us-west.elb.amazonaws.com`,
}

var envRmCmd = &cobra.Command{
	Use:   "rm <env-name>",
	Short: "Delete an environment from a ksonnet project",
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
	Long: `Delete an environment from a ksonnet project. This is the same
as removing the <env-name> environment directory and all files contained. All empty
parent directories are also subsequently deleted.`,
	Example: `# Remove the directory 'us-west/staging' and all contents in the 'environments'
# directory. This will also remove the parent directory 'us-west' if it is empty.
ks env rm us-west/staging`,
}

var envListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all environments in a ksonnet project",
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
	}, Long: `List all environments in a ksonnet project. This will
display the name, server, and namespace of each environment within the ksonnet project.`,
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

		server, namespace, err := resolveEnvFlags(flags)

		c, err := kubecfg.NewEnvSetCmd(originalName, name, server, namespace, manager)
		if err != nil {
			return err
		}

		return c.Run()
	},
	Long: `Set environment fields such as the name, and server. Changing
the name of an environment will also update the directory structure in
'environments'.`,
	Example: `# Updates the API server address of the environment 'us-west/staging'.
ks env set us-west/staging --server=http://example.com

# Updates the namespace of the environment 'us-west/staging'.
ks env set us-west/staging --namespace=staging

# Updates both the name and the server of the environment 'us-west/staging'.
# Updating the name will update the directory structure in 'environments'.
ks env set us-west/staging --server=http://example.com --name=us-east/staging
  
# Updates API server address of the environment 'us-west/staging' based on the
# server in the context 'staging-west' in your kubeconfig file.
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

func resolveEnvFlags(flags *pflag.FlagSet) (string, string, error) {
	defaultNamespace := "default"

	server, ns, context, err := commonEnvFlags(flags)
	if err != nil {
		return "", "", err
	}

	if server == "" {
		// server is not provided -- use the context.
		server, defaultNamespace, err = resolveContext(context)
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

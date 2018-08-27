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
	"os"
	"path/filepath"

	"github.com/ksonnet/ksonnet/pkg/log"
	"github.com/ksonnet/ksonnet/pkg/plugin"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	// Register auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

const (
	rootLong = `
You can use the ` + "`ks`" + ` commands to write, share, and deploy your Kubernetes
application configuration to remote clusters.

----
	`
)

func runPlugin(fs afero.Fs, root string, p plugin.Plugin, args []string) error {
	env := []string{
		fmt.Sprintf("KS_PLUGIN_DIR=%s", p.RootDir),
		fmt.Sprintf("KS_PLUGIN_NAME=%s", p.Config.Name),
		fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
	}

	appConfig := filepath.Join(root, "app.yaml")
	exists, err := afero.Exists(fs, appConfig)
	if err != nil {
		return err
	}

	if exists {
		env = append(env, fmt.Sprintf("KS_APP_DIR=%s", root))
		// TODO: make kube context or something similar available to the plugin
	}

	cmd := p.BuildRunCmd(env, args)
	return cmd.Run()
}

// addEnvCmdFlags adds the flags that are common to the family of commands
// whose form is `[<env>|-f <file-name>]`, e.g., `apply` and `delete`.
func addEnvCmdFlags(cmd *cobra.Command) {
	cmd.PersistentFlags().StringSliceP(flagComponent, shortComponent, nil, "Name of a specific component (multiple -c flags accepted, allows YAML, JSON, and Jsonnet)")
}

// NewRoot constructs the root cobra command
func NewRoot(appFs afero.Fs, wd string, args []string) (*cobra.Command, error) {
	if appFs == nil {
		appFs = afero.NewOsFs()
	}

	rootCmd := &cobra.Command{
		Use:           "ks",
		Short:         `Configure your application to deploy to a Kubernetes cluster`,
		Long:          rootLong,
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			flags := cmd.Flags()
			verbosity, err := flags.GetCount(flagVerbose)
			if err != nil {
				return err
			}

			log.Init(verbosity, cmd.OutOrStderr())

			return nil
		},
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cobra.NoArgs(cmd, args)
			}

			pluginName := args[0]
			_, err := plugin.Find(appFs, pluginName)
			if err != nil {
				return cobra.NoArgs(cmd, args)
			}

			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			pluginName, args := args[0], args[1:]
			p, err := plugin.Find(appFs, pluginName)
			if err != nil {
				return err
			}

			root := viper.GetString(flagDir)
			if root == "" {
				root, err = os.Getwd()
			}
			if err != nil {
				return err
			}

			return runPlugin(appFs, root, p, args)
		},
	}

	rootCmd.SetArgs(args)

	rootCmd.PersistentFlags().CountP(flagVerbose, "v", "Increase verbosity. May be given multiple times.")
	rootCmd.PersistentFlags().Set("logtostderr", "true")
	rootCmd.PersistentFlags().Bool(flagTLSSkipVerify, false, "Skip verification of TLS server certificates")
	rootCmd.PersistentFlags().String(flagDir, wd, "Ksonnet application root to use; Defaults to CWD")
	viper.BindPFlag(flagTLSSkipVerify, rootCmd.PersistentFlags().Lookup(flagTLSSkipVerify))
	viper.BindPFlag(flagDir, rootCmd.PersistentFlags().Lookup(flagDir))

	rootCmd.AddCommand(newApplyCmd(appFs))
	rootCmd.AddCommand(newComponentCmd())
	rootCmd.AddCommand(newDeleteCmd(appFs))
	rootCmd.AddCommand(newDiffCmd(appFs))
	rootCmd.AddCommand(newEnvCmd())
	rootCmd.AddCommand(newGenerateCmd(appFs))
	rootCmd.AddCommand(newImportCmd())
	rootCmd.AddCommand(newInitCmd(appFs, wd))
	rootCmd.AddCommand(newModuleCmd())
	rootCmd.AddCommand(newParamCmd())
	rootCmd.AddCommand(newPkgCmd())
	rootCmd.AddCommand(newPrototypeCmd(appFs))
	rootCmd.AddCommand(newRegistryCmd())
	rootCmd.AddCommand(newShowCmd(appFs))
	rootCmd.AddCommand(newValidateCmd(appFs))
	rootCmd.AddCommand(newUpgradeCmd())
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd, nil
}

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

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/log"
	"github.com/ksonnet/ksonnet/pkg/plugin"
	"github.com/shomron/pflag"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

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

func runPlugin(fs afero.Fs, p plugin.Plugin, args []string) error {
	env := []string{
		fmt.Sprintf("KS_PLUGIN_DIR=%s", p.RootDir),
		fmt.Sprintf("KS_PLUGIN_NAME=%s", p.Config.Name),
		fmt.Sprintf("HOME=%s", os.Getenv("HOME")),
	}

	root, err := appRoot()
	if err != nil {
		return err
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

func appRoot() (string, error) {
	return os.Getwd()
}

// parseCommand does an early parse of the command line and returns
// what will ultimately be recognized as the command by cobra.
func parseCommand(args []string) (string, error) {
	fset := pflag.NewFlagSet("", pflag.ContinueOnError)
	fset.ParseErrorsWhitelist.UnknownFlags = true
	fset.BoolP("help", "h", false, "") // Needed to avoid pflag.ErrHelp
	if err := fset.Parse(args); err != nil {
		return "", err
	}
	if len(fset.Args()) == 0 {
		return "", nil
	}

	return fset.Args()[0], nil
}

func NewRoot(appFs afero.Fs, wd string, args []string) (*cobra.Command, error) {
	if appFs == nil {
		appFs = afero.NewOsFs()
	}

	var a app.App
	var err error

	cmdName, err := parseCommand(args)
	if err != nil {
		return nil, err
	}

	if len(args) > 0 && cmdName != "init" {
		a, err = app.Load(appFs, wd, false)
		if err != nil {
			return nil, err
		}
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

			return runPlugin(appFs, p, args)
		},
	}

	rootCmd.SetArgs(args)

	rootCmd.PersistentFlags().CountP(flagVerbose, "v", "Increase verbosity. May be given multiple times.")
	rootCmd.PersistentFlags().Set("logtostderr", "true")

	rootCmd.AddCommand(newApplyCmd(a))
	rootCmd.AddCommand(newComponentCmd(a))
	rootCmd.AddCommand(newDeleteCmd(a))
	rootCmd.AddCommand(newDiffCmd(a))
	rootCmd.AddCommand(newEnvCmd(a))
	rootCmd.AddCommand(newGenerateCmd(a))
	rootCmd.AddCommand(newImportCmd(a))
	rootCmd.AddCommand(newInitCmd(appFs, wd))
	rootCmd.AddCommand(newModuleCmd(a))
	rootCmd.AddCommand(newParamCmd(a))
	rootCmd.AddCommand(newPkgCmd(a))
	rootCmd.AddCommand(newPrototypeCmd(a))
	rootCmd.AddCommand(newRegistryCmd(a))
	rootCmd.AddCommand(newShowCmd(a))
	rootCmd.AddCommand(newValidateCmd(a))
	rootCmd.AddCommand(newUpgradeCmd(a))
	rootCmd.AddCommand(newVersionCmd())

	return rootCmd, nil
}

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
	"github.com/ksonnet/ksonnet/pkg/util/strings"
	"github.com/pkg/errors"
	"github.com/shomron/pflag"
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

type earlyParseArgs struct {
	command       string
	help          bool
	tlsSkipVerify bool
}

// parseCommand does an early parse of the command line and returns
// what will ultimately be recognized as the command by cobra
// and a boolean for calling help flags.
func parseCommand(args []string) (earlyParseArgs, error) {
	var parsed earlyParseArgs
	fset := pflag.NewFlagSet("", pflag.ContinueOnError)
	fset.ParseErrorsWhitelist.UnknownFlags = true
	fset.BoolVarP(&parsed.help, "help", "h", false, "") // Needed to avoid pflag.ErrHelp
	fset.BoolVar(&parsed.tlsSkipVerify, flagTLSSkipVerify, false, "")
	if err := fset.Parse(args); err != nil {
		return earlyParseArgs{}, err
	}
	if len(fset.Args()) == 0 {
		return earlyParseArgs{}, nil
	}

	parsed.command = fset.Args()[0]
	return parsed, nil
}

// checkUpgrade runs upgrade validations unless the user is running an excluded command.
// If upgrades are found to be necessary, they will be reported to the user.
func checkUpgrade(a app.App, cmd string) error {
	skip := map[string]struct{}{
		"init":    struct{}{},
		"upgrade": struct{}{},
		"help":    struct{}{},
		"version": struct{}{},
		"":        struct{}{},
	}
	if _, ok := skip[cmd]; ok {
		return nil
	}

	if a == nil {
		return errors.Errorf("nil receiver")
	}
	_, _ = a.CheckUpgrade() // NOTE we're surpressing any validation errors here
	return nil
}

func NewRoot(appFs afero.Fs, wd string, args []string) (*cobra.Command, error) {
	if appFs == nil {
		appFs = afero.NewOsFs()
	}

	var a app.App

	parsed, err := parseCommand(args)
	if err != nil {
		return nil, err
	}
	httpClient := app.NewHTTPClient(parsed.tlsSkipVerify)

	cmds := []string{"init", "version", "help"}
	switch {
	// Commands that do not require a ksonnet application
	case strings.InSlice(parsed.command, cmds), parsed.help:
		a, err = app.Load(appFs, httpClient, wd, true)
	case len(args) > 0:
		a, err = app.Load(appFs, httpClient, wd, false)
	default:
		// noop
	}

	if err != nil {
		return nil, err
	}

	if err := checkUpgrade(a, parsed.command); err != nil {
		return nil, errors.Wrap(err, "checking if app needs upgrade")
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
	rootCmd.PersistentFlags().Bool(flagTLSSkipVerify, false, "Skip verification of TLS server certificates")
	viper.BindPFlag(flagTLSSkipVerify, rootCmd.PersistentFlags().Lookup(flagTLSSkipVerify))

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

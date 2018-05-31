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
	goflag "flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/log"
	"github.com/ksonnet/ksonnet/pkg/plugin"
	"github.com/spf13/afero"
	"github.com/spf13/cobra"

	// Register auth plugins
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)

var (
	appFs = afero.NewOsFs()
	ka    app.App
)

func init() {
	RootCmd.PersistentFlags().CountP(flagVerbose, "v", "Increase verbosity. May be given multiple times.")
	RootCmd.PersistentFlags().Set("logtostderr", "true")
}

// RootCmd is the root of cobra subcommand tree
var RootCmd = &cobra.Command{
	Use:   "ks",
	Short: `Configure your application to deploy to a Kubernetes cluster`,
	Long: `
You can use the ` + "`ks`" + ` commands to write, share, and deploy your Kubernetes
application configuration to remote clusters.

----
`,
	SilenceErrors: true,
	SilenceUsage:  true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		goflag.CommandLine.Parse([]string{})
		flags := cmd.Flags()

		verbosity, err := flags.GetCount(flagVerbose)
		if err != nil {
			return err
		}

		log.Init(verbosity, cmd.OutOrStderr())

		wd, err := os.Getwd()
		if err != nil {
			return err
		}

		var isInit bool
		if len(args) == 2 && args[0] == "init" {
			isInit = true
		}

		ka, err = app.Load(appFs, wd, false)
		if err != nil && isInit {
			return err
		}

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

		return runPlugin(p, args)
	},
}

func runPlugin(p plugin.Plugin, args []string) error {
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
	exists, err := afero.Exists(appFs, appConfig)
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

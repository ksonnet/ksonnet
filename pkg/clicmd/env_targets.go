// Copyright 2018 The ksonnet authors
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
	"github.com/ksonnet/ksonnet/pkg/actions"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	vEnvTargetModules  = "env-target-modules"
	vEnvTargetOverride = "env-target-override-flag"
)

var (
	envTargetLong = `
The ` + "`targets`" + ` command selects one or more modules to be applied by an
environment. The default environment target is the root module, ` + "`/`" + `.

Changing targets for an environment will require specifying all desired modules including the root module.
`
	envTargetExample = `
# Create a new module
ks module create db

# Generate a component and specify the module
ks generate redis-stateless redis --module db

# Change the default environment target from / to db
# The targets are tracked in app.yaml
ks env targets default --module db`
)

func newEnvTargetsCmd() *cobra.Command {
	envTargetsCmd := &cobra.Command{
		Use:     "targets",
		Short:   envShortDesc["targets"],
		Long:    envTargetLong,
		Example: envTargetExample,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("'env targets' takes only one argument. Pass the environment name followed by the module flag with the target module")
			}

			m := map[string]interface{}{
				actions.OptionEnvName:  args[0],
				actions.OptionModule:   viper.GetStringSlice(vEnvTargetModules),
				actions.OptionOverride: viper.GetBool(vEnvTargetOverride),
			}
			addGlobalOptions(m)

			return runAction(actionEnvTargets, m)
		},
	}

	envTargetsCmd.Flags().StringSlice(flagModule, nil, "Component modules to include")
	viper.BindPFlag(vEnvTargetModules, envTargetsCmd.Flags().Lookup(flagModule))

	envTargetsCmd.Flags().BoolP(flagOverride, shortOverride, false, "Set targets in environment as override")
	viper.BindPFlag(vEnvTargetOverride, envTargetsCmd.Flags().Lookup(flagOverride))

	return envTargetsCmd
}

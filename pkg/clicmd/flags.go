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
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	// For use in the commands (e.g., diff, apply, delete) that require either an
	// environment or the -f flag.
	flagAPISpec               = "api-spec"
	flagAsString              = "as-string"
	flagComponent             = "component"
	flagCreate                = "create"
	flagDir                   = "dir"
	flagDryRun                = "dry-run"
	flagEnv                   = "env"
	flagExtVar                = "ext-str"
	flagExtVarFile            = "ext-str-file"
	flagFilename              = "filename"
	flagForce                 = "force"
	flagFormat                = "format"
	flagGcTag                 = "gc-tag"
	flagGracePeriod           = "grace-period"
	flagInstalled             = "installed"
	flagJpath                 = "jpath"
	flagModule                = "module"
	flagNamespace             = "namespace"
	flagResolveImage          = "resolve-image"
	flagServer                = "server"
	flagSet                   = "set"
	flagSkipDefaultRegistries = "skip-default-registries"
	flagSkipGc                = "skip-gc"
	flagTlaVar                = "tla-str"
	flagTlaVarFile            = "tla-str-file"
	flagOutput                = "output"
	flagOverride              = "override"
	flagUnset                 = "unset"
	flagVerbose               = "verbose"
	flagVersion               = "version"
	flagWithoutModules        = "without-modules"

	shortComponent = "c"
	shortFilename  = "f"
	shortFormat    = "o"
	shortOutput    = "o"
	shortOverride  = "o"
)

// addCmdOutput adds an output flag to a command. `name` is the name
// of the viper assignment.
func addCmdOutput(cmd *cobra.Command, name string) {
	cmd.Flags().StringP(flagOutput, shortOutput, "", "Output format. Valid options: table|json")
	viper.BindPFlag(name, cmd.Flags().Lookup(flagOutput))
}

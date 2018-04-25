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
	"strings"

	"github.com/ksonnet/ksonnet/pkg/env"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

func bindJsonnetFlags(cmd *cobra.Command, name string) {
	cmd.Flags().StringSliceP(flagJpath, "J", nil, "Additional jsonnet library search path")
	viper.BindPFlag(name+"-jpath", cmd.Flags().Lookup(flagJpath))

	cmd.Flags().StringSliceP(flagExtVar, "V", nil, "Values of external variables")
	viper.BindPFlag(name+"-ext-var", cmd.Flags().Lookup(flagExtVar))

	cmd.Flags().StringSlice(flagExtVarFile, nil, "Read external variable from a file")
	viper.BindPFlag(name+"-ext-var-file", cmd.Flags().Lookup(flagExtVarFile))

	cmd.Flags().StringSliceP(flagTlaVar, "A", nil, "Values of top level arguments")
	viper.BindPFlag(name+"-tla-var", cmd.Flags().Lookup(flagTlaVar))

	cmd.Flags().StringSlice(flagTlaVarFile, nil, "Read top level argument from a file")
	viper.BindPFlag(name+"-tla-var-file", cmd.Flags().Lookup(flagTlaVarFile))
}

func extractJsonnetFlags(name string) error {
	jPaths := viper.GetStringSlice(name + "-jpath")
	env.AddJPaths(jPaths...)

	extVars := viper.GetStringSlice(name + "-ext-var")
	for _, s := range extVars {
		k, v, err := splitJsonnetFlag(s)
		if err != nil {
			return errors.Wrap(err, "ext vars flag")
		}

		env.AddExtVar(k, v)
	}

	extVarFiles := viper.GetStringSlice(name + "-ext-var-file")
	for _, s := range extVarFiles {
		k, v, err := splitJsonnetFlag(s)
		if err != nil {
			return errors.Wrap(err, "ext var files flag")
		}

		if err = env.AddExtVarFile(ka, k, v); err != nil {
			return errors.Wrap(err, "add ext var file")
		}
	}

	extTlas := viper.GetStringSlice(name + "-tla-var")
	for _, s := range extTlas {
		k, v, err := splitJsonnetFlag(s)
		if err != nil {
			return errors.Wrap(err, "tla vars flag")
		}

		env.AddTlaVar(k, v)
	}

	extTlaFiles := viper.GetStringSlice(name + "-tla-var-file")
	for _, s := range extTlaFiles {
		k, v, err := splitJsonnetFlag(s)
		if err != nil {
			return errors.Wrap(err, "tla var files flag")
		}

		if err = env.AddTlaVarFile(ka, k, v); err != nil {
			return errors.Wrap(err, "add tla var file")
		}
	}

	return nil
}

func splitJsonnetFlag(in string) (key string, value string, err error) {
	parts := strings.SplitN(in, "=", 2)
	if len(parts) != 2 {
		return "", "", errors.Errorf("unable to find key and value in %q", in)
	}

	return parts[0], parts[1], nil
}

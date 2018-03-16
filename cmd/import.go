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

package cmd

import (
	"github.com/ksonnet/ksonnet/actions"
	"github.com/pkg/errors"
	"github.com/spf13/viper"

	"github.com/spf13/cobra"
)

const (
	vImportFilename  = "import-filename"
	vImportNamespace = "import-namespace"
)

// importCmd represents the import command
var importCmd = &cobra.Command{
	Use:   "import",
	Short: "Import manifest",
	Long:  `Import manifest`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fileName := viper.GetString(vImportFilename)
		if fileName == "" {
			return errors.New("filename is required")
		}

		namespace := viper.GetString(vImportNamespace)

		return actions.RunImport(ka, namespace, fileName)
	},
}

func init() {
	RootCmd.AddCommand(importCmd)

	importCmd.Flags().StringP(flagFilename, shortFilename, "", "Filename or directory for component to import")
	viper.BindPFlag(vImportFilename, importCmd.Flags().Lookup(flagFilename))
	importCmd.Flags().String(flagNamespace, "", "Component namespace")
	viper.BindPFlag(vImportNamespace, importCmd.Flags().Lookup(flagNamespace))
}

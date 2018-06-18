// Copyright 2017 The ksonnet authors
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

package main

import (
	"os"

	"github.com/ksonnet/ksonnet/pkg/clicmd"
	"github.com/sirupsen/logrus"
	"github.com/spf13/afero"
	"github.com/spf13/cobra/doc"
)

func main() {
	outputDir := os.Args[1]

	wd, err := os.Getwd()
	if err != nil {
		logrus.WithError(err).Error("unable to find working directory")
		os.Exit(1)
	}

	rootCmd, err := clicmd.NewRoot(afero.NewOsFs(), wd, []string{})
	if err != nil {
		logrus.WithError(err).Error("unable create ksonnet command")
		os.Exit(1)
	}

	// Remove auto-generated timestamps
	rootCmd.DisableAutoGenTag = true

	err = doc.GenMarkdownTree(rootCmd, outputDir)
	if err != nil {
		logrus.Fatal(err)
	}
}

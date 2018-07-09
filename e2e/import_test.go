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

// +build e2e

package e2e

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("ks import", func() {
	var (
		a          app
		importPath string
		isErr      bool
		o          *output
	)

	BeforeEach(func() {
		a = e.initApp(nil)
		a.generateDeployedService()
		isErr = false
	})

	JustBeforeEach(func() {
		o = a.runKs("import", "-f", importPath, "--module", "/")
		if isErr {
			assertExitStatus(o, 1)
			return
		}

		assertExitStatus(o, 0)
	})

	Context("directory", func() {
		BeforeEach(func() {
			importPath = filepath.Join(e.wd(), "testdata", "input", "import")
		})

		It("imports the files in the directory", func() {
			expected := componentListRow{
				APIVersion: "apps/v1beta1",
				Component:  "",
				Kind:       "Deployment",
				Name:       "nginx-deployment",
				Type:       "yaml",
			}

			a.checkComponentPrefix(expected, "deployment-nginx-deployment-")
		})
	})

	Context("file", func() {
		BeforeEach(func() {
			importPath = filepath.Join(e.wd(), "testdata", "input", "import", "deployment.yaml")
		})

		It("imports the file", func() {
			expected := componentListRow{
				APIVersion: "apps/v1beta1",
				Component:  "",
				Kind:       "Deployment",
				Name:       "nginx-deployment",
				Type:       "yaml",
			}

			a.checkComponentPrefix(expected, "deployment-nginx-deployment-")
		})
	})

	Context("invalid path", func() {
		BeforeEach(func() {
			importPath = filepath.Join(e.wd(), "testdata", "input", "import", "invalid.yaml")
			isErr = true
		})

		It("returns an error", func() {
			assertOutputContains("import/invalid.txt", o.stderr)
		})
	})
})

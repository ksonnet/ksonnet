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
	. "github.com/onsi/gomega"
)

var _ = Describe("ks init", func() {
	var a app

	var opts = &initOptions{}

	JustBeforeEach(func() {
		a = e.initApp(opts)
	})

	Context("in general", func() {
		It("doesn't generate default registries", func() {
			a.checkRegistryExists("incubator")
		})

		It("creates a gitignore", func() {
			assertContents(filepath.Join("init", "gitignore"), filepath.Join(a.dir, ".gitignore"))
		})
	})

	Context("without default registries", func() {
		BeforeEach(func() {
			opts.skipRegistries = true
		})

		It("doesn't generate default registries", func() {
			o := a.registryList()
			tr := loadTableResponse(o.stdout)
			Expect(tr.registryList()).To(BeEmpty())
		})
	})

	Context("with a custom environment name", func() {
		BeforeEach(func() {
			opts.envName = "env-name"
		})

		It("sets the the specified environment name", func() {
			expected := []envListRow{
				{
					KubernetesVersion: e.serverVersion(),
					Name:              "env-name",
					Namespace:         "default",
					Override:          "",
					Server:            "http://example.com",
				},
			}

			a.checkEnvs(expected)
		})
	})
})

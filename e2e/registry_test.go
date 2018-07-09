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

var _ = Describe("ks registry", func() {
	var a app

	BeforeEach(func() {
		a = e.initApp(nil)
		a.generateDeployedService()
	})

	Describe("add", func() {
		var (
			uri         string
			expectedURI string
		)
		Context("global", func() {
			JustBeforeEach(func() {
				o := a.runKs("registry", "add", "local", uri)
				assertExitStatus(o, 0)
			})

			Context("a filesystem based registry", func() {
				Context("as a path", func() {
					BeforeEach(func() {
						path, err := filepath.Abs(filepath.Join("testdata", "registries", "parts-infra"))
						Expect(err).ToNot(HaveOccurred())
						uri = path
						expectedURI = path
					})

					It("adds a registry", func() {
						a.checkRegistry("local", "", "fs", expectedURI)
					})
				})
				Context("as a URL", func() {
					BeforeEach(func() {
						path, err := filepath.Abs(filepath.Join("testdata", "registries", "parts-infra"))
						Expect(err).ToNot(HaveOccurred())
						uri = convertPathToURI(path)
						expectedURI = path
					})

					It("adds a registry", func() {
						a.checkRegistry("local", "", "fs", expectedURI)
					})
				})
			})
		})

		Context("override", func() {
			var registryName string

			JustBeforeEach(func() {
				o := a.runKs("registry", "add", registryName, uri, "--override")
				assertExitStatus(o, 0)
			})

			Context("a filesystem based registry", func() {

				BeforeEach(func() {
					registryName = "local"
				})

				Context("as a path", func() {
					BeforeEach(func() {
						path, err := filepath.Abs(filepath.Join("testdata", "registries", "parts-infra"))
						Expect(err).ToNot(HaveOccurred())
						uri = path
						expectedURI = path
					})

					It("adds a registry", func() {
						a.checkRegistry("local", "*", "fs", uri)
					})
				})
				Context("as a URL", func() {
					BeforeEach(func() {
						path, err := filepath.Abs(filepath.Join("testdata", "registries", "parts-infra"))
						Expect(err).ToNot(HaveOccurred())
						uri = convertPathToURI(path)
						expectedURI = path
					})

					It("adds a registry", func() {
						a.checkRegistry("local", "*", "fs", expectedURI)
					})
				})
			})

			Context("an existing configuration", func() {
				BeforeEach(func() {
					path, err := filepath.Abs(filepath.Join("testdata", "registries", "parts-infra"))
					Expect(err).ToNot(HaveOccurred())
					uri = path
					expectedURI = path
					registryName = "incubator"
				})

				It("adds a registry", func() {
					a.checkRegistry(registryName, "*", "fs", expectedURI)
				})
			})
		})
	})

	Describe("list", func() {
		It("lists the currently configured registries", func() {
			o := a.runKs("registry", "list")
			assertExitStatus(o, 0)
			assertOutput("registry/list/output.txt", o.stdout)
		})
	})

	Describe("describe", func() {
		Context("incubator", func() {
			It("describe a registry", func() {
				o := a.runKs("registry", "describe", "incubator")
				assertExitStatus(o, 0)
				assertOutput("registry/describe/output.txt", o.stdout)
			})
		})
	})
})

// +build e2e

package e2e

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("ks prototype", func() {
	var a app

	BeforeEach(func() {
		a = e.initApp("")
		a.generateDeployedService()
	})

	Describe("describe", func() {
		Context("with a long name", func() {
			It("shows the description", func() {
				o := a.runKs("prototype", "describe", "io.ksonnet.pkg.configMap")
				assertExitStatus(o, 0)
				assertOutput("prototype/describe/output.txt", o.stdout)
			})
		})

		Context("with a short name", func() {
			It("shows the description", func() {
				o := a.runKs("prototype", "describe", "configMap")
				assertExitStatus(o, 0)
				assertOutput("prototype/describe/output.txt", o.stdout)
			})
		})
	})

	Describe("list", func() {
		It("lists available prototypes", func() {
			o := a.runKs("prototype", "list")
			assertExitStatus(o, 0)
			assertOutput("prototype/list/output.txt", o.stdout)
		})
	})

	Describe("preview", func() {
		It("shows the prototype preview", func() {
			o := a.runKs("prototype", "preview", "deployed-service",
				"--name", "aName",
				"--image", "image:tag")
			assertExitStatus(o, 0)
			assertOutput("prototype/preview/output.txt", o.stdout)
		})
	})

	Describe("search", func() {
		It("returns a list of prototypes whose name maches the search term", func() {
			o := a.runKs("prototype", "search", "service")
			assertExitStatus(o, 0)
			assertOutput("prototype/search/output.txt", o.stdout)
		})
	})
})

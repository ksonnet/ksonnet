// +build e2e

package e2e

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("ks registry", func() {
	var a app

	BeforeEach(func() {
		a = e.initApp("")
		a.generateDeployedService()
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

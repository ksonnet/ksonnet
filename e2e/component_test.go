// +build e2e

package e2e

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("ks component", func() {
	var a app

	BeforeEach(func() {
		a = e.initApp("")
		a.generateDeployedService()
	})

	Describe("list", func() {
		It("lists the components for a namespace", func() {
			o := a.runKs("component", "list")
			assertExitStatus(o, 0)
			assertOutput("component/list/output.txt", o.stdout)
		})
	})

	Describe("rm", func() {
		It("removes a component", func() {
			o := a.runKs("component", "rm", "guestbook-ui")
			assertExitStatus(o, 0)

			o = a.componentList()
			assertOutput("component/rm/output.txt", o.stdout)

			o = a.paramList()
			assertOutput("component/rm/params-output.txt", o.stdout)
		})
	})
})

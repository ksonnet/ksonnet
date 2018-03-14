// +build e2e

package e2e

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("ks param", func() {
	var a app

	BeforeEach(func() {
		a = e.initApp("")
		a.generateDeployedService()
	})

	Describe("list", func() {
		Context("at the component level", func() {
			It("lists the params for a namespace", func() {
				o := a.runKs("param", "list")
				assertExitStatus(o, 0)
				assertOutput("param/list/output.txt", o.stdout)
			})
		})

		Context("at the environment level", func() {
			It("lists the params for a namespace", func() {
				a.paramSet("guestbook-ui", "replicas", "3", "--env", "default")

				o := a.paramList()
				assertExitStatus(o, 0)
				assertOutput("param/list/output.txt", o.stdout)

				o = a.runKs("param", "list", "--env", "default")
				assertExitStatus(o, 0)
				assertOutput("param/list/env.txt", o.stdout)
			})
		})
	})

	Describe("set", func() {
		Context("at the component level", func() {
			It("updates a parameter's value", func() {
				o := a.runKs("param", "set", "guestbook-ui", "replicas", "3")
				assertExitStatus(o, 0)

				o = a.paramList()
				assertOutput("param/set/output.txt", o.stdout)
			})
		})

		Context("at the environment level", func() {
			It("updates a parameter's environment value", func() {
				o := a.runKs("param", "set", "guestbook-ui", "replicas", "3", "--env", "default")
				assertExitStatus(o, 0)

				o = a.paramList()
				assertOutput("param/set/local-output.txt", o.stdout)

				o = a.paramList("--env", "default")
				assertOutput("param/set/env-default-output.txt", o.stdout)

			})
		})
	})

})

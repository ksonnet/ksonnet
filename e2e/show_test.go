// +build e2e

package e2e

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("ks show", func() {
	var a app

	BeforeEach(func() {
		a = e.initApp("")
		a.generateDeployedService()
	})

	It("shows the generated YAML", func() {
		o := a.runKs("show", "default")
		assertExitStatus(o, 0)
		assertOutput("show/output.txt", o.stdout)
	})

})

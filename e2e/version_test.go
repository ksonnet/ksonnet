// +build e2e

package e2e

import (
	. "github.com/onsi/ginkgo"
)

var _ = Describe("ks version", func() {

	It("shows ks version", func() {
		o := e.ks("version")
		assertExitStatus(o, 0)
		assertOutput("version/default.txt", o.stdout)
	})

})

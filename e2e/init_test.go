// +build e2e

package e2e

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("ks init", func() {
	var a app

	var opts = &initOptions{}

	JustBeforeEach(func() {
		a = e.initApp(opts)
	})

	Context("in general", func() {
		It("doesn't generate default registries", func() {
			o := a.registryList()
			assertOutput(filepath.Join("init", "registry-output.txt"), o.stdout)
		})
	})

	Context("without default registries", func() {
		BeforeEach(func() {
			opts.skipRegistries = true
		})

		It("doesn't generate default registries", func() {
			o := a.registryList()
			assertOutput(filepath.Join("init", "skip-registry-output.txt"), o.stdout)
		})
	})
})

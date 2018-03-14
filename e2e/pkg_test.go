// +build e2e

package e2e

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ks pkg", func() {
	var a app

	BeforeEach(func() {
		a = e.initApp("")
	})

	Describe("describe", func() {
		Context("incubator/apache", func() {
			It("describes the package", func() {
				o := a.runKs("pkg", "describe", "incubator/apache")
				assertExitStatus(o, 0)
				assertOutput("pkg/describe/output.txt", o.stdout)
			})
		})
	})

	Describe("install", func() {
		Context("incubator/apache", func() {
			It("describes the package", func() {
				o := a.runKs("pkg", "install", "incubator/apache")
				assertExitStatus(o, 0)

				pkgDir := filepath.Join(a.dir, "vendor", "incubator", "apache")
				Expect(pkgDir).To(BeADirectory())
			})
		})
	})

	Describe("list", func() {
		It("lists available packages", func() {
			o := a.runKs("pkg", "list")
			assertExitStatus(o, 0)
			assertOutput("pkg/list/output.txt", o.stdout)
		})
	})

	Context("use", func() {
		It("generates a component using the prototype", func() {
			o := a.runKs(
				"prototype", "use", "deployed-service", "guestbook-ui",
				"--image", "gcr.io/heptio-images/ks-guestbook-demo:0.1",
				"--type", "ClusterIP")
			assertExitStatus(o, 0)

			component := filepath.Join(a.dir, "components", "guestbook-ui.jsonnet")
			assertContents("generate/guestbook-ui.jsonnet", component)

			params := filepath.Join(a.dir, "components", "params.libsonnet")
			assertContents("generate/params.libsonnet", params)
		})
	})

})

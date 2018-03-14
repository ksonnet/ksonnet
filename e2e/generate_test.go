// +build e2e

package e2e

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
)

var _ = Describe("ks generate", func() {
	var a app

	BeforeEach(func() {
		a = e.initApp("")
	})

	Context("deployed-service", func() {
		It("generates a component using the prototype", func() {
			o := a.runKs(
				"generate", "deployed-service", "guestbook-ui",
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

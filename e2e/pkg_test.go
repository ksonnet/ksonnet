// Copyright 2018 The kubecfg authors
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

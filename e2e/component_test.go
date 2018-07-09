// Copyright 2018 The ksonnet authors
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
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ks component", func() {
	var a app

	BeforeEach(func() {
		a = e.initApp(nil)
		a.generateDeployedService()
	})

	Describe("list", func() {
		It("lists the components for a namespace", func() {
			a.checkComponentName("guestbook-ui")
		})

		Context("with a module", func() {
			It("lists the components for a module", func() {
				a.checkComponentName("guestbook-ui", "--module", "/")
			})
		})
	})

	Describe("rm", func() {
		It("removes a component", func() {
			o := a.runKs("component", "rm", "guestbook-ui", "-v")
			assertExitStatus(o, 0)

			o = a.componentList()
			assertExitStatus(o, 0)

			tr := loadTableResponse(o.stdout)
			Expect(tr.componentList()).To(BeEmpty())

			o = a.paramList()
			assertExitStatus(o, 0)

			tr = loadTableResponse(o.stdout)
			Expect(tr.paramList()).To(BeEmpty())
		})
	})
})

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
	. "github.com/onsi/ginkgo"
)

var _ = Describe("ks env", func() {
	var a app

	BeforeEach(func() {
		a = e.initApp("")
		a.generateDeployedService()
	})

	Describe("add", func() {
		It("adds an environment", func() {
			o := a.runKs("env", "add", "prod",
				"--server", "http://example.com",
				"--namespace", "prod")
			assertExitStatus(o, 0)

			o = a.envList()
			assertOutput("env/add/output.txt", o.stdout)
		})
	})

	Describe("list", func() {
		It("lists environments", func() {
			o := a.runKs("env", "list")
			assertExitStatus(o, 0)
			assertOutput("env/list/output.txt", o.stdout)
		})
	})

	Describe("rm", func() {
		It("removes an environment", func() {
			o := a.envAdd("prod")

			o = a.runKs("env", "rm", "prod")
			assertExitStatus(o, 0)

			o = a.envList()
			assertOutput("env/rm/output.txt", o.stdout)
		})
	})

	Describe("set", func() {
		Context("updating env name", func() {
			It("updates the name of an environment", func() {
				o := a.envAdd("prod")

				o = a.runKs("env", "set", "prod", "--name", "us-west1/prod")
				assertExitStatus(o, 0)

				o = a.envList()
				assertOutput("env/set/rename.txt", o.stdout)
			})
		})
	})
})

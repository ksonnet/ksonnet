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
)

var _ = Describe("ks param", func() {
	var a app

	BeforeEach(func() {
		a = e.initApp(nil)

	})

	Describe("delete", func() {
		var (
			component  = "guestbook-ui"
			envName    = "default"
			local      = "local-value"
			localValue = "1"
			env        = "env-value"
			envValue   = "2"
		)

		BeforeEach(func() {
			a.generateDeployedService()

			a.paramSet(component, local, localValue)
			a.paramSet(component, env, envValue, "--env", envName)

			o := a.paramList()
			assertOutput("param/delete/pre-local.txt", o.stdout)

			o = a.paramList("--env", envName)
			assertOutput("param/delete/pre-env.txt", o.stdout)
		})

		Context("at the component level", func() {
			JustBeforeEach(func() {
				o := a.runKs("param", "delete", component, local)
				assertExitStatus(o, 0)
			})

			It("removes a parameter's value", func() {
				o := a.paramList()
				assertOutput("param/delete/local.txt", o.stdout)
			})
		})

		Context("at the environment level", func() {
			JustBeforeEach(func() {
				o := a.runKs("param", "delete", component, env, "--env", envName)
				assertExitStatus(o, 0)
			})

			It("removes a parameter's environment value", func() {
				o := a.paramList("--env", envName)
				assertOutput("param/delete/env.txt", o.stdout)
			})
		})

		Context("removing environment global", func() {
			BeforeEach(func() {
				o := a.runKs("param", "set", "department", "engineering", "--env", "default")
				assertExitStatus(o, 0)
			})

			JustBeforeEach(func() {
				o := a.runKs("param", "delete", "department", "--env", "default")
				assertExitStatus(o, 0)
			})

			It("removes the value", func() {
				o := a.paramList("--env", "default")
				assertOutput("param/delete/env-global.txt", o.stdout)
			})
		})
	})

	Describe("diff", func() {
		var extraOptions []string
		var diffOutput *output

		BeforeEach(func() {
			a.generateDeployedService()

			o := a.runKs("env", "add", "env1")
			assertExitStatus(o, 0)

			o = a.runKs("param", "set", "guestbook-ui", "replicas", "4", "--env", "env1")
			assertExitStatus(o, 0)

			o = a.runKs("env", "add", "env2")
			assertExitStatus(o, 0)

			o = a.runKs("param", "set", "guestbook-ui", "replicas", "3", "--env", "env2")
			assertExitStatus(o, 0)
		})

		JustBeforeEach(func() {
			options := append([]string{"param", "diff", "env1", "env2"}, extraOptions...)
			diffOutput = a.runKs(options...)
		})

		It("runs successfully", func() {
			assertExitStatus(diffOutput, 0)
		})

		It("lists the differences", func() {
			assertOutput(filepath.Join("param", "diff", "output.txt"), diffOutput.stdout)
		})

		Context("with a component", func() {
			BeforeEach(func() {
				extraOptions = []string{"--component", "guestbook-ui"}
			})

			It("runs successfully", func() {
				assertExitStatus(diffOutput, 0)
			})

			It("lists the differences", func() {
				assertOutput(filepath.Join("param", "diff", "output.txt"), diffOutput.stdout)
			})

		})
	})

	Describe("list", func() {
		var (
			listOutput *output
			listParams = []string{"param", "list", "-v"}
		)

		JustBeforeEach(func() {
			listOutput = a.runKs(listParams...)
		})

		Describe("at the component level", func() {
			Context("with jsonnet component params", func() {
				BeforeEach(func() {
					a.generateDeployedService()
				})
				It("should exit with 0", func() {
					assertExitStatus(listOutput, 0)
				})

				It("lists the params for a module", func() {
					assertOutput("param/list/output.txt", listOutput.stdout)
				})
			})

			Context("with yaml component params", func() {
				BeforeEach(func() {
					deployment := filepath.Join(e.wd(), "testdata", "input", "import", "deployment.yaml")
					o := a.runKs("import", "-f", deployment)
					assertExitStatus(o, 0)

					o = a.runKs("param", "set", "deployment-nginx-deployment", "metadata.labels", `{"hello": "world"}`)
					assertExitStatus(o, 0)
				})

				It("should exit with 0", func() {
					assertExitStatus(listOutput, 0)
				})

				It("should list the YAML params", func() {
					assertOutput("param/list/yaml-params.txt", listOutput.stdout)
				})
			})
		})

		Describe("at the environment level", func() {
			BeforeEach(func() {
				a.generateDeployedService()

				a.paramSet("guestbook-ui", "replicas", "3", "--env", "default")
				listParams = []string{"param", "list", "--env", "default"}
			})

			It("should exit with 0", func() {
				assertExitStatus(listOutput, 0)
			})

			It("lists the params for a module", func() {
				assertOutput("param/list/env.txt", listOutput.stdout)
			})
		})
	})

	Describe("set", func() {
		BeforeEach(func() {
			a.generateDeployedService()
		})
		Describe("at the component level", func() {
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

		Context("setting environment global", func() {
			JustBeforeEach(func() {
				o := a.runKs("param", "set", "department", "engineering", "--env", "default")
				assertExitStatus(o, 0)
			})

			It("sets the value", func() {
				o := a.paramList("--env", "default")
				assertOutput("param/set/env-global.txt", o.stdout)
			})
		})
	})

})

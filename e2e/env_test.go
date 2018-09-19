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
	ksapp "github.com/ksonnet/ksonnet/pkg/app"
	yaml "gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ks env", func() {
	var a app

	BeforeEach(func() {
		a = e.initApp(nil)
		a.generateDeployedService()
	})

	Describe("add", func() {
		It("adds an environment", func() {
			o := a.runKs("env", "add", "prod",
				"--server", "http://example.com",
				"--namespace", "prod")
			assertExitStatus(o, 0)

			expected := setEnvListRow(genEnvList(e.serverVersion()),
				"prod", e.serverVersion(), "prod", "", "http://example.com")

			a.checkEnvs(expected)
		})

		Context("override", func() {
			It("adds an environment as an override", func() {
				o := a.runKs("env", "add", "prod",
					"-o",
					"--server", "http://example.com",
					"--namespace", "prod")
				assertExitStatus(o, 0)

				expected := setEnvListRow(genEnvList(e.serverVersion()),
					"prod", e.serverVersion(), "prod", "*", "http://example.com")

				a.checkEnvs(expected)
			})
		})
	})

	Describe("describe", func() {
		It("describes an environment", func() {
			o := a.runKs("env", "describe", "default")
			assertExitStatus(o, 0)

			envConfig := &ksapp.EnvironmentConfig{
				Name:              "default",
				KubernetesVersion: e.serverVersion(),
				Path:              "default",
				Destination: &ksapp.EnvironmentDestinationSpec{
					Server:    "http://example.com",
					Namespace: "default",
				},
			}

			data, err := yaml.Marshal(envConfig)
			Expect(err).ToNot(HaveOccurred())

			Expect(o.stdout).To(Equal(string(data)))
		})

		Context("with override", func() {
			It("describes the override environment", func() {
				o := a.runKs("env", "add", "default",
					"-o",
					"--server", "http://example.com",
					"--namespace", "prod")
				assertExitStatus(o, 0)

				o = a.runKs("env", "describe", "default")
				assertExitStatus(o, 0)

				envConfig := &ksapp.EnvironmentConfig{
					Name:              "default",
					KubernetesVersion: e.serverVersion(),
					Path:              "default",
					Destination: &ksapp.EnvironmentDestinationSpec{
						Server:    "http://example.com",
						Namespace: "prod",
					},
				}

				data, err := yaml.Marshal(envConfig)
				Expect(err).ToNot(HaveOccurred())

				Expect(o.stdout).To(Equal(string(data)))
			})
		})
	})

	Describe("list", func() {
		It("lists environments", func() {
			a.checkEnvs(genEnvList(e.serverVersion()))
		})
	})

	Describe("rm", func() {
		It("removes an environment", func() {
			o := a.envAdd("prod", false)
			assertExitStatus(o, 0)

			o = a.runKs("env", "rm", "prod")
			assertExitStatus(o, 0)

			a.checkEnvs(genEnvList(e.serverVersion()))
		})

		Context("with an override", func() {
			It("removes an override", func() {
				o := a.envAdd("default", true)
				assertExitStatus(o, 0)

				expected := setEnvListRow(genEnvList(e.serverVersion()),
					"default", e.serverVersion(), "default", "*", "http://example.com")
				a.checkEnvs(expected)

				o = a.runKs("env", "rm", "-o", "default")
				assertExitStatus(o, 0)

				a.checkEnvs(genEnvList(e.serverVersion()))
			})
		})

		It("attempt to remove an invalid environment", func() {
			o := a.runKs("env", "rm", "invalid")
			assertExitStatus(o, 1)
			Expect(o.stderr).To(ContainSubstring(`environment \"invalid\" does not exist`))
		})
	})

	Describe("set", func() {
		Context("updating env name", func() {
			It("updates the name of an environment", func() {
				o := a.envAdd("prod", false)
				assertExitStatus(o, 0)

				o = a.runKs("env", "set", "prod", "--name", "us-west1/prod")
				assertExitStatus(o, 0)

				expected := []envListRow{
					{
						KubernetesVersion: e.serverVersion(),
						Name:              "default",
						Namespace:         "default",
						Override:          "",
						Server:            "http://example.com",
					},
					{
						KubernetesVersion: e.serverVersion(),
						Name:              "us-west1/prod",
						Namespace:         "prod",
						Override:          "",
						Server:            "http://example.com",
					},
				}

				a.checkEnvs(expected)
			})
		})

		Context("updating namespace", func() {
			It("updates the namespace for an environment", func() {
				o := a.runKs("env", "set", "default", "--namespace", "dev")
				assertExitStatus(o, 0)

				o = a.envDescribe("default")

				envConfig := &ksapp.EnvironmentConfig{
					Name:              "default",
					KubernetesVersion: e.serverVersion(),
					Path:              "default",
					Destination: &ksapp.EnvironmentDestinationSpec{
						Server:    "http://example.com",
						Namespace: "dev",
					},
				}

				data, err := yaml.Marshal(envConfig)
				Expect(err).ToNot(HaveOccurred())

				Expect(o.stdout).To(Equal(string(data)))
			})
		})

		Context("with override", func() {
			Context("update the namespace", func() {
				It("updates the override namespace", func() {
					o := a.envAdd("default", true)
					assertExitStatus(o, 0)

					o = a.runKs("env", "set", "default", "--override", "--namespace", "new-name")
					assertExitStatus(o, 0)

					expected := []envListRow{
						{
							KubernetesVersion: e.serverVersion(),
							Name:              "default",
							Namespace:         "new-name",
							Override:          "*",
							Server:            "http://example.com",
						},
					}

					a.checkEnvs(expected)
				})
			})

			Context("update the name", func() {
				It("renames the environment", func() {
					o := a.envAdd("default", true)
					assertExitStatus(o, 0)

					o = a.runKs("env", "set", "default", "--override", "--name", "new-name")
					assertExitStatus(o, 0)

					expected := []envListRow{
						{
							KubernetesVersion: e.serverVersion(),
							Name:              "default",
							Namespace:         "default",
							Override:          "",
							Server:            "http://example.com",
						},
						{
							KubernetesVersion: e.serverVersion(),
							Name:              "new-name",
							Namespace:         "default",
							Override:          "*",
							Server:            "http://example.com",
						},
					}

					a.checkEnvs(expected)
				})
			})
		})
	})

	Describe("targets", func() {
		Context("namespace exists", func() {
			Context("updating the targets", func() {
				It("updates the name of an environment", func() {
					o := a.runKs("env", "targets", "default",
						"--module", "/")
					assertExitStatus(o, 0)

					envConfig := &ksapp.EnvironmentConfig{
						Name:              "default",
						KubernetesVersion: e.serverVersion(),
						Path:              "default",
						Destination: &ksapp.EnvironmentDestinationSpec{
							Server:    "http://example.com",
							Namespace: "default",
						},
						Targets: []string{"/"},
					}

					data, err := yaml.Marshal(envConfig)
					Expect(err).ToNot(HaveOccurred())

					o = a.envDescribe("default")
					Expect(o.stdout).To(Equal(string(data)))
				})
			})

			Context("target module does not exist", func() {
				It("return an error", func() {
					o := a.runKs("env", "targets", "default",
						"--module", "bad")
					assertExitStatus(o, 1)
					assertOutput("env/targets/invalid-target.txt", o.stderr)
				})
			})
		})

		Context("environment does not exist", func() {
			It("returns an error", func() {
				o := a.runKs("env", "targets", "invalid",
					"--module", "/")
				assertExitStatus(o, 1)
				assertOutput("env/targets/invalid-env.txt", o.stderr)
			})
		})
	})
})

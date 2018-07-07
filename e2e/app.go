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

package e2e

import (
	"path/filepath"
	"strings"

	// gomega matchers
	// nolint: golint

	. "github.com/onsi/gomega"
)

type app struct {
	dir string
	e2e *e2e
}

func (a *app) runKs(args ...string) *output {
	return a.e2e.ksInApp(a.dir, args...)
}

func (a *app) componentList(opts ...string) *output {
	o := a.runKs(append([]string{"component", "list", "-o", "json"}, opts...)...)
	assertExitStatus(o, 0)

	return o
}

func (a *app) checkComponents(expected []componentListRow, args ...string) {
	o := a.componentList(args...)
	assertExitStatus(o, 0)

	tr := loadTableResponse(o.stdout)
	got := tr.componentList()

	ExpectWithOffset(1, got).To(Equal(expected))
}

func (a *app) checkComponentName(name string, args ...string) {
	o := a.componentList(args...)
	assertExitStatus(o, 0)

	tr := loadTableResponse(o.stdout)
	rows := tr.componentList()

	found := false
	for _, row := range rows {
		if row.Component == name {
			found = true
			break
		}
	}

	Expect(found).To(BeTrue(), "component %s was not found", name)
}

func (a *app) checkComponent(expected componentListRow, args ...string) {
	o := a.componentList(args...)
	assertExitStatus(o, 0)

	tr := loadTableResponse(o.stdout)
	rows := tr.componentList()

	found := false
	for _, row := range rows {
		if row.Component == expected.Component {
			ExpectWithOffset(1, row).To(Equal(expected))
			found = true
			break
		}
	}

	Expect(found).To(BeTrue(), "component %s was not found", expected.Component)
}

func (a *app) checkComponentPrefix(expected componentListRow, prefix string, args ...string) {
	o := a.componentList(args...)
	assertExitStatus(o, 0)

	tr := loadTableResponse(o.stdout)
	rows := tr.componentList()

	found := false
	for _, row := range rows {
		if strings.HasPrefix(row.Component, row.Component) {
			expected.Component = row.Component
			ExpectWithOffset(1, row).To(Equal(expected))
			found = true
			break
		}
	}

	Expect(found).To(BeTrue(), "component with prefix %s was not found", prefix)
}

func (a *app) apply(namespace string, opts ...string) *output {
	args := append([]string{namespace}, opts...)
	return a.runKs(args...)
}

func (a *app) envAdd(nsName string, override bool) *output {
	args := []string{
		"env",
		"add",
		nsName,
		"--server", "http://example.com",
		"--namespace", nsName,
	}

	if override {
		args = append(args, "-o")
	}

	o := a.runKs(args...)
	assertExitStatus(o, 0)

	return o
}

func (a *app) envDescribe(envName string) *output {
	o := a.runKs("env", "describe", envName)
	assertExitStatus(o, 0)

	return o
}

func (a *app) envList() *output {
	o := a.runKs("env", "list", "-o", "json")
	assertExitStatus(o, 0)

	return o
}

func (a *app) checkEnvs(expected []envListRow) {
	o := a.envList()
	assertExitStatus(o, 0)

	tr := loadTableResponse(o.stdout)
	got := tr.envList()

	ExpectWithOffset(1, got).To(Equal(expected))
}

func (a *app) paramList(args ...string) *output {
	o := a.runKs(append([]string{"param", "list", "-o", "json"}, args...)...)
	assertExitStatus(o, 0)

	return o
}

func (a *app) checkParams(expected []paramListRow, args ...string) {
	o := a.paramList(args...)
	assertExitStatus(o, 0)

	tr := loadTableResponse(o.stdout)
	got := tr.paramList()

	ExpectWithOffset(1, got).To(Equal(expected))
}

func (a *app) pkgInstall(partName string) *output {
	o := a.runKs("pkg", "install", partName)
	assertExitStatus(o, 0)

	return o
}

func (a *app) pkgList(args ...string) *output {
	o := a.runKs(append([]string{"pkg", "list", "-o", "json"}, args...)...)
	assertExitStatus(o, 0)

	return o
}

func (a *app) checkPkgs(expected []pkgListRow, args ...string) {
	o := a.pkgList(args...)
	assertExitStatus(o, 0)

	tr := loadTableResponse(o.stdout)
	got := tr.pkgList()

	ExpectWithOffset(1, got).To(Equal(expected))
}

func (a *app) checkInstalledPkg(registry, name string) {
	o := a.pkgList()
	assertExitStatus(o, 0)

	tr := loadTableResponse(o.stdout)

	found := false

	for _, row := range tr.pkgList() {
		if row.Registry == registry && row.Name == name && row.Installed == "*" {
			found = true
			break
		}
	}

	ExpectWithOffset(1, found).To(Equal(true), "%s/%s is not installed", registry, name)
}

func (a *app) paramSet(key, value string, args ...string) *output {
	o := a.runKs(append([]string{"param", "set", key, value}, args...)...)
	assertExitStatus(o, 0)

	return o
}

func (a *app) prototypeList(args ...string) *output {
	o := a.runKs(append([]string{"prototype", "list", "-o", "json"}, args...)...)
	assertExitStatus(o, 0)

	return o
}

func (a *app) checkPrototype(name string, args ...string) {
	o := a.prototypeList(args...)
	assertExitStatus(o, 0)

	tr := loadTableResponse(o.stdout)

	found := false

	for _, row := range tr.prototypeList() {
		if row.Name == name {
			found = true
			break
		}
	}

	ExpectWithOffset(1, found).To(Equal(true), "prototype %s does not exist", name)
}

func (a *app) registryAdd(registryName, uri string) *output {
	o := a.runKs("registry", "add", registryName, uri)
	assertExitStatus(o, 0)

	return o
}

func (a *app) registryList(args ...string) *output {
	o := a.runKs(append([]string{"registry", "list", "-o", "json"}, args...)...)
	assertExitStatus(o, 0)

	return o
}

func (a *app) checkRegistry(name, override, prototol, uri string, args ...string) {
	o := a.registryList(args...)
	assertExitStatus(o, 0)

	tr := loadTableResponse(o.stdout)

	found := false

	for _, row := range tr.registryList() {
		if row.Name == name && row.Override == override &&
			row.Protocol == prototol && row.URI == uri {
			found = true
			break
		}
	}

	ExpectWithOffset(1, found).To(Equal(true), "registry %s does not exist", name)
}

func (a *app) checkRegistryExists(name string, args ...string) {
	o := a.registryList(args...)
	assertExitStatus(o, 0)

	tr := loadTableResponse(o.stdout)

	found := false

	for _, row := range tr.registryList() {
		if row.Name == name {
			found = true
			break
		}
	}

	ExpectWithOffset(1, found).To(Equal(true), "registry %s does not exist", name)
}

func (a *app) generateDeployedService() {
	appDir := a.dir

	o := a.runKs(
		"generate", "deployed-service", "guestbook-ui",
		"--image", "gcr.io/heptio-images/ks-guestbook-demo:0.1",
		"--type", "ClusterIP")
	assertExitStatus(o, 0)

	component := filepath.Join(appDir, "components", "guestbook-ui.jsonnet")
	assertContents("generate/guestbook-ui.jsonnet", component)

	params := filepath.Join(appDir, "components", "params.libsonnet")
	assertContents("generate/params.libsonnet", params)
}

func (a *app) findComponent(name, kind string) string {
	o := a.componentList("-o", "json")

	tr := loadTableResponse(o.stdout)
	rows := tr.componentList()

	var component string
	for _, row := range rows {
		if name == row.Name && kind == row.Kind {
			component = row.Component
		}
	}

	ExpectWithOffset(1, component).
		ToNot(BeEmpty(), "unable to find YAML component with name %q and kind %q",
			name, kind)
	return component
}

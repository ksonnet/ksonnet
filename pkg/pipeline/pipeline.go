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

package pipeline

import (
	"bytes"
	"io"
	"regexp"

	jsonnet "github.com/google/go-jsonnet"
	"github.com/ksonnet/ksonnet/component"
	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// OverrideManager overrides the component manager interface for a pipeline.
func OverrideManager(c component.Manager) Opt {
	return func(p *Pipeline) {
		p.cm = c
	}
}

// Opt is an option for configuring Pipeline.
type Opt func(p *Pipeline)

// Pipeline is the ks build pipeline.
type Pipeline struct {
	app     app.App
	envName string
	cm      component.Manager
}

// New creates an instance of Pipeline.
func New(ksApp app.App, envName string, opts ...Opt) *Pipeline {
	p := &Pipeline{
		app:     ksApp,
		envName: envName,
		cm:      component.DefaultManager,
	}

	for _, opt := range opts {
		opt(p)
	}

	return p
}

// Namespaces returns the namespaces that belong to this pipeline.
func (p *Pipeline) Namespaces() ([]component.Namespace, error) {
	return p.cm.Namespaces(p.app, p.envName)
}

// EnvParameters creates parameters for a namespace given an environment.
func (p *Pipeline) EnvParameters(nsName string) (string, error) {
	ns, err := p.cm.Namespace(p.app, nsName)
	if err != nil {
		return "", err
	}

	paramsStr, err := p.cm.NSResolveParams(ns)
	if err != nil {
		return "", err
	}

	data, err := p.app.EnvironmentParams(p.envName)
	if err != nil {
		return "", err
	}

	envParams := upgradeParams(p.envName, data)

	vm := jsonnet.MakeVM()
	vm.ExtCode("__ksonnet/params", paramsStr)
	return vm.EvaluateSnippet("snippet", string(envParams))
}

// Components returns the components that belong to this pipeline.
func (p *Pipeline) Components(filter []string) ([]component.Component, error) {
	namespaces, err := p.Namespaces()
	if err != nil {
		return nil, err
	}

	components := make([]component.Component, 0)
	for _, ns := range namespaces {
		members, err := p.cm.Components(ns)
		if err != nil {
			return nil, err
		}

		members = filterComponents(filter, members)
		components = append(components, members...)
	}

	return components, nil
}

// Objects converts components into Kubernetes objects.
func (p *Pipeline) Objects(filter []string) ([]*unstructured.Unstructured, error) {
	namespaces, err := p.Namespaces()
	if err != nil {
		return nil, err
	}

	objects := make([]*unstructured.Unstructured, 0)
	for _, ns := range namespaces {
		paramsStr, err := p.EnvParameters(ns.Name())
		if err != nil {
			return nil, err
		}

		components, err := p.Components(filter)
		if err != nil {
			return nil, err
		}

		for _, c := range components {
			o, err := c.Objects(paramsStr, p.envName)
			if err != nil {
				return nil, err
			}

			objects = append(objects, o...)
		}
	}

	return objects, nil
}

// YAML converts components into YAML.
func (p *Pipeline) YAML(filter []string) (io.Reader, error) {
	objects, err := p.Objects(filter)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	if err := Fprint(&buf, objects, "yaml"); err != nil {
		return nil, errors.Wrap(err, "convert objects to YAML")
	}

	return &buf, nil
}

func filterComponents(filter []string, components []component.Component) []component.Component {
	if len(filter) == 0 {
		return components
	}

	var out []component.Component
	for _, c := range components {
		if stringInSlice(c.Name(true), filter) {
			out = append(out, c)
		}
	}

	return out
}

var (
	reParamSwap = regexp.MustCompile(`(?m)import "\.\.\/\.\.\/components\/params\.libsonnet"`)
)

// upgradeParams replaces relative params imports with an extVar to handle
// multiple component namespaces.
// NOTE: It warns when it makes a change. This serves as a temporary fix until
// ksonnet generates the correct file.
func upgradeParams(envName, in string) string {
	logrus.Warnf("rewriting %q environment params to not use relative paths", envName)
	return reParamSwap.ReplaceAllLiteralString(in, `std.extVar("__ksonnet/params")`)
}

func stringInSlice(s string, sl []string) bool {
	for i := range sl {
		if sl[i] == s {
			return true
		}
	}

	return false
}

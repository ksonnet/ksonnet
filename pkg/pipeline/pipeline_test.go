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
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet-lib/ksonnet-gen/astext"
	"github.com/ksonnet/ksonnet/pkg/app"
	appmocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/component"
	cmocks "github.com/ksonnet/ksonnet/pkg/component/mocks"
	"github.com/ksonnet/ksonnet/pkg/metadata"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestPipeline_Namespaces(t *testing.T) {
	withPipeline(t, func(p *Pipeline, m *cmocks.Manager, a *appmocks.App) {
		namespaces := []component.Module{}
		m.On("Modules", p.app, "default").Return(namespaces, nil)

		got, err := p.Modules()
		require.NoError(t, err)

		require.Equal(t, namespaces, got)
	})
}

func TestPipeline_EnvParameters(t *testing.T) {
	withPipeline(t, func(p *Pipeline, m *cmocks.Manager, a *appmocks.App) {
		ns := component.NewModule(p.app, "/")
		namespaces := []component.Module{ns}
		m.On("Modules", p.app, "default").Return(namespaces, nil)
		m.On("Module", p.app, "/").Return(ns, nil)
		m.On("NSResolveParams", ns).Return("", nil)
		a.On("EnvironmentParams", "default").Return("{}", nil)

		env := &app.EnvironmentConfig{Path: "default"}
		a.On("Environment", "default").Return(env, nil)

		got, err := p.EnvParameters("/")
		require.NoError(t, err)

		require.Equal(t, "{ }\n", got)
	})
}

func TestPipeline_Components(t *testing.T) {
	withPipeline(t, func(p *Pipeline, m *cmocks.Manager, a *appmocks.App) {
		cpnt := &cmocks.Component{}
		components := []component.Component{cpnt}

		ns := component.NewModule(p.app, "/")
		namespaces := []component.Module{ns}
		m.On("Modules", p.app, "default").Return(namespaces, nil)
		m.On("Module", p.app, "/").Return(ns, nil)
		m.On("NSResolveParams", ns).Return("", nil)
		a.On("EnvironmentParams", "default").Return("{}", nil)
		m.On("Components", ns).Return(components, nil)

		got, err := p.Components(nil)
		require.NoError(t, err)

		require.Equal(t, components, got)
	})
}

func mockComponent(name string) *cmocks.Component {
	c := &cmocks.Component{}
	c.On("Name", true).Return(name)
	return c
}

func TestPipeline_Components_filtered(t *testing.T) {
	withPipeline(t, func(p *Pipeline, m *cmocks.Manager, a *appmocks.App) {

		cpnt1 := mockComponent("cpnt1")
		cpnt2 := mockComponent("cpnt2")
		components := []component.Component{cpnt1, cpnt2}

		ns := component.NewModule(p.app, "/")
		namespaces := []component.Module{ns}
		m.On("Modules", p.app, "default").Return(namespaces, nil)
		m.On("Module", p.app, "/").Return(ns, nil)
		m.On("NSResolveParams", ns).Return("", nil)
		a.On("EnvironmentParams", "default").Return("{}", nil)
		m.On("Components", ns).Return(components, nil)

		got, err := p.Components([]string{"cpnt1"})
		require.NoError(t, err)

		expected := []component.Component{cpnt1}

		require.Equal(t, expected, got)
	})
}

func TestPipeline_Objects(t *testing.T) {
	withPipeline(t, func(p *Pipeline, m *cmocks.Manager, a *appmocks.App) {
		u := []*unstructured.Unstructured{
			{
				Object: map[string]interface{}{
					"apiVersion": "v1",
					"kind":       "Service",
					"metadata": map[string]interface{}{
						"labels": map[string]interface{}{
							metadata.LabelComponent: "service",
						},
						"name": "my-service",
					},
					"spec": map[string]interface{}{
						"ports": []interface{}{
							map[string]interface{}{
								"port":       int64(80),
								"protocol":   "TCP",
								"targetPort": int64(80),
							},
						},
					},
				},
			},
		}

		module := &cmocks.Module{}
		module.On("Name").Return("")
		object := &astext.Object{}
		componentMap := map[string]string{"service": "yaml"}
		module.On("Render", "default").Return(object, componentMap, nil)
		module.On("ResolvedParams").Return("", nil)

		modules := []component.Module{module}
		m.On("Modules", p.app, "default").Return(modules, nil)
		m.On("Module", p.app, "/").Return(module, nil)
		m.On("NSResolveParams", module).Return("", nil)
		a.On("EnvironmentParams", "default").Return("{}", nil)

		env := &app.EnvironmentConfig{Path: "default"}
		a.On("Environment", "default").Return(env, nil)

		serviceJSON, err := ioutil.ReadFile(filepath.Join("testdata", "components.json"))
		require.NoError(t, err)
		p.evaluateEnvFn = func(_ app.App, envName, input, params string) (string, error) {
			return string(serviceJSON), nil
		}

		p.evaluateEnvParamsFn = func(_ app.App, paramsPath, paramData, envName string) (string, error) {
			return `{"components": {}}`, nil
		}

		got, err := p.Objects(nil)
		require.NoError(t, err)

		require.Equal(t, u, got)
	})
}

func TestPipeline_YAML(t *testing.T) {
	withPipeline(t, func(p *Pipeline, m *cmocks.Manager, a *appmocks.App) {
		p.buildObjectsFn = func(_ *Pipeline, filter []string) ([]*unstructured.Unstructured, error) {
			u := []*unstructured.Unstructured{
				{
					Object: map[string]interface{}{
						"apiVersion": "v1",
						"kind":       "Service",
						"metadata": map[string]interface{}{
							"name": "my-service",
						},
						"spec": map[string]interface{}{
							"ports": []interface{}{
								map[string]interface{}{
									"port":       int64(80),
									"protocol":   "TCP",
									"targetPort": int64(80),
								},
							},
						},
					},
				},
			}

			return u, nil
		}

		r, err := p.YAML(nil)
		require.NoError(t, err)

		got, err := ioutil.ReadAll(r)
		require.NoError(t, err)

		expected, err := ioutil.ReadFile(filepath.Join("testdata", "service.yaml"))
		require.NoError(t, err)

		require.Equal(t, string(expected), string(got))
	})
}

func Test_upgradeParams(t *testing.T) {
	in := `local params = import "../../components/params.libsonnet";`
	expected := `local params = std.extVar("__ksonnet/params");`

	got := upgradeParams("default", in)
	require.Equal(t, expected, got)
}

func withPipeline(t *testing.T, fn func(p *Pipeline, m *cmocks.Manager, a *appmocks.App)) {
	a := &appmocks.App{}
	a.On("Root").Return("/")
	envName := "default"

	manager := &cmocks.Manager{}

	p := New(a, envName, OverrideManager(manager))

	fn(p, manager, a)
}

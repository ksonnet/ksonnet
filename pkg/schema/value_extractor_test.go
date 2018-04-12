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

package schema

import (
	"testing"

	jsonnetutil "github.com/ksonnet/ksonnet/pkg/util/jsonnet"
	"github.com/stretchr/testify/require"
)

const (
	crdPrefix = "apiextensions.v1beta1.customResourceDefinition."
)

type veTestCase struct {
	props    Properties
	gvk      GVK
	expected map[string]Values
}

var (
	veTestCases = []veTestCase{
		{
			props: Properties{
				"metadata": map[interface{}]interface{}{
					"name": "certificates.certmanager.k8s.io",
					"labels": map[interface{}]interface{}{
						"app":      "cert-manager",
						"chart":    "cert-manager-0.2.2",
						"release":  "cert-manager",
						"heritage": "Tiller",
					},
				},
				"spec": map[interface{}]interface{}{
					"group":   "certmanager.k8s.io",
					"version": "v1alpha1",
					"names": map[interface{}]interface{}{
						"kind":   "Certificate",
						"plural": "certificates",
					},
					"scope": "Namespaced",
				},
			},

			gvk: GVK{
				GroupPath: []string{
					"apiextensions.k8s.io",
				},
				Version: "v1beta1",
				Kind:    "customResourceDefinition",
			},

			expected: map[string]Values{
				crdPrefix + "mixin.metadata.labels": Values{
					Lookup: []string{"metadata", "labels"},
					Setter: crdPrefix + "mixin.metadata.withLabels",
					Value: map[interface{}]interface{}{
						"app":      "cert-manager",
						"chart":    "cert-manager-0.2.2",
						"release":  "cert-manager",
						"heritage": "Tiller",
					},
				},
				crdPrefix + "mixin.metadata.name": Values{
					Lookup: []string{"metadata", "name"},
					Setter: crdPrefix + "mixin.metadata.withName",
					Value:  "certificates.certmanager.k8s.io",
				},
				crdPrefix + "mixin.spec.group": Values{
					Lookup: []string{"spec", "group"},
					Setter: crdPrefix + "mixin.spec.withGroup",
					Value:  "certmanager.k8s.io",
				},
				crdPrefix + "mixin.spec.names.kind": Values{
					Lookup: []string{"spec", "names", "kind"},
					Setter: crdPrefix + "mixin.spec.names.withKind",
					Value:  "Certificate",
				},
				crdPrefix + "mixin.spec.names.plural": Values{
					Lookup: []string{"spec", "names", "plural"},
					Setter: crdPrefix + "mixin.spec.names.withPlural",
					Value:  "certificates",
				},
				crdPrefix + "mixin.spec.scope": Values{
					Lookup: []string{"spec", "scope"},
					Setter: crdPrefix + "mixin.spec.withScope",
					Value:  "Namespaced",
				},
				crdPrefix + "mixin.spec.version": Values{
					Lookup: []string{"spec", "version"},
					Setter: crdPrefix + "mixin.spec.withVersion",
					Value:  "v1alpha1",
				},
			},
		},
		{
			props: Properties{
				"metadata": map[interface{}]interface{}{
					"name": "nginx-deployment",
					"labels": map[interface{}]interface{}{
						"app": "nginx",
					},
				},
				"spec": map[interface{}]interface{}{
					"selector": map[interface{}]interface{}{
						"matchLabels": map[interface{}]interface{}{
							"app": "nginx",
						},
					},
					"template": map[interface{}]interface{}{
						"metadata": map[interface{}]interface{}{
							"labels": map[interface{}]interface{}{
								"app": "nginx",
							},
						},
						"spec": map[interface{}]interface{}{
							"containers": []interface{}{
								map[interface{}]interface{}{
									"name":  "nginx",
									"image": "nginx:1.7.9",
									"ports": []interface{}{
										map[interface{}]interface{}{
											"containerPort": 80,
										},
									},
								},
							},
						},
					},
					"replicas": 3,
				},
			},
			gvk: GVK{
				GroupPath: []string{"apps"},
				Version:   "v1beta2",
				Kind:      "deployment",
			},
			expected: map[string]Values{
				"apps.v1beta2.deployment.mixin.spec.selector.matchLabels": Values{
					Lookup: []string{"spec", "selector", "matchLabels"},
					Setter: "apps.v1beta2.deployment.mixin.spec.selector.withMatchLabels",
					Value:  map[interface{}]interface{}{"app": "nginx"}},
				"apps.v1beta2.deployment.mixin.spec.template.metadata.labels": Values{
					Lookup: []string{"spec", "template", "metadata", "labels"},
					Setter: "apps.v1beta2.deployment.mixin.spec.template.metadata.withLabels",
					Value:  map[interface{}]interface{}{"app": "nginx"}},
				"apps.v1beta2.deployment.mixin.spec.template.spec.containers": Values{
					Lookup: []string{"spec", "template", "spec", "containers"},
					Setter: "apps.v1beta2.deployment.mixin.spec.template.spec.withContainers",
					Value: []interface{}{map[interface{}]interface{}{
						"image": "nginx:1.7.9",
						"ports": []interface{}{
							map[interface{}]interface{}{"containerPort": 80}}, "name": "nginx"}}},
				"apps.v1beta2.deployment.mixin.metadata.labels": Values{
					Lookup: []string{"metadata", "labels"},
					Setter: "apps.v1beta2.deployment.mixin.metadata.withLabels",
					Value:  map[interface{}]interface{}{"app": "nginx"}},
				"apps.v1beta2.deployment.mixin.metadata.name": Values{
					Lookup: []string{"metadata", "name"},
					Setter: "apps.v1beta2.deployment.mixin.metadata.withName",
					Value:  "nginx-deployment"},
				"apps.v1beta2.deployment.mixin.spec.replicas": Values{
					Lookup: []string{"spec", "replicas"},
					Setter: "apps.v1beta2.deployment.mixin.spec.withReplicas",
					Value:  3}},
		},
	}
)

func TestValueExtractor_Extract(t *testing.T) {
	node, err := jsonnetutil.Import("testdata/k8s.libsonnet")
	require.NoError(t, err)

	cases := []struct {
		name       string
		veTestCase veTestCase
	}{
		// {
		// 	name:       "crd",
		// 	veTestCase: veTestCases[0],
		// },
		{
			name:       "deployment",
			veTestCase: veTestCases[1],
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ve := NewValueExtractor(node)
			got, err := ve.Extract(tc.veTestCase.gvk, tc.veTestCase.props)
			require.NoError(t, err)

			require.Equal(t, tc.veTestCase.expected, got)
		})
	}

}

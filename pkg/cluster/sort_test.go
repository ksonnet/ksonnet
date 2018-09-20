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

package cluster

import (
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func makeUnstructured(apiVersion string, kind string) *unstructured.Unstructured {
	return &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
			"metadata": map[string]interface{}{
				"name": "important-object",
			},
		},
	}
}

func makeUnstructuredNS(apiVersion string, kind string, namespace string) *unstructured.Unstructured {
	u := makeUnstructured(apiVersion, kind)
	metadata := u.Object["metadata"].(map[string]interface{})
	metadata["namespace"] = namespace
	return u
}

func makeUnstructuredName(apiVersion string, kind string, name string) *unstructured.Unstructured {
	u := makeUnstructured(apiVersion, kind)
	metadata := u.Object["metadata"].(map[string]interface{})
	metadata["name"] = name
	return u
}

func Test_UnstructuredSlice_Sort(t *testing.T) {
	objects := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1beta2",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "d3",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "s2",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "d1",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "d2",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "s1",
				},
			},
		},
	}

	expected := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "s1",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name": "s2",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "d1",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1beta2",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "d3",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "extensions/v1beta1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name": "d2",
				},
			},
		},
	}

	for i := 0; i < 10; i++ {
		UnstructuredSlice(objects).Sort()
	}

	require.Equal(t, expected, objects)
}

func Test_UnstructuredSlice_Sort_GroupVersion(t *testing.T) {
	objects := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"apiVersion": "custom.crd.com/v1beta1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name": "myjob",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name": "myjob",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "custom.crd.com/v1beta1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name": "a-comes-first",
				},
			},
		},
	}

	expected := []*unstructured.Unstructured{
		{
			Object: map[string]interface{}{
				"apiVersion": "batch/v1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name": "myjob",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "custom.crd.com/v1beta1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name": "a-comes-first",
				},
			},
		},
		{
			Object: map[string]interface{}{
				"apiVersion": "custom.crd.com/v1beta1",
				"kind":       "Job",
				"metadata": map[string]interface{}{
					"name": "myjob",
				},
			},
		},
	}

	UnstructuredSlice(objects).Sort()

	require.Equal(t, expected, objects)
}

func Test_UnstructuredSlice_Sort_ByKind(t *testing.T) {
	objects := []*unstructured.Unstructured{
		makeUnstructured("apiregistration.k8s.io/v1beta1", "APIService"),
		makeUnstructured("extensions/v1beta1", "Ingress"),
		makeUnstructured("batch/v1beta1", "CronJob"),
		makeUnstructured("batch/v1", "Job"),
		makeUnstructured("apps/v1", "StatefulSet"),
		makeUnstructured("apps/v1", "Deployment"),
		makeUnstructured("apps/v1", "ReplicaSet"),
		makeUnstructured("core/v1", "ReplicationController"),
		makeUnstructured("core/v1", "Pod"),
		makeUnstructured("apps/v1", "DaemonSet"),
		makeUnstructured("v1", "Service"),
		makeUnstructured("v1", "UnknownGoesToTheEnd-b"),
		makeUnstructured("v1", "UnknownGoesToTheEnd"),
		makeUnstructured("rbac.authorization.k8s.io/v1", "RoleBinding"),
		makeUnstructured("rbac.authorization.k8s.io/v1", "Role"),
		makeUnstructured("rbac.authorization.k8s.io/v1", "ClusterRoleBinding"),
		makeUnstructured("rbac.authorization.k8s.io/v1", "ClusterRole"),
		makeUnstructured("rbac.authorization.k8s.io/v1", "CustomResourceDefinition"),
		makeUnstructured("v1", "ServiceAccount"),
		makeUnstructured("v1", "PersistentVolumeClaim"),
		makeUnstructured("v1", "PersistentVolume"),
		makeUnstructured("storage.k8s.io/v1beta1", "StorageClass"),
		makeUnstructured("v1", "ConfigMap"),
		makeUnstructured("v1", "Secret"),
		makeUnstructured("extensions/v1beta1", "PodSecurityPolicy"),
		makeUnstructured("v1", "LimitRange"),
		makeUnstructured("v1", "ResourceQuota"),
		makeUnstructured("v1", "Namespace"),
	}

	expected := []*unstructured.Unstructured{
		makeUnstructured("v1", "Namespace"),
		makeUnstructured("v1", "ResourceQuota"),
		makeUnstructured("v1", "LimitRange"),
		makeUnstructured("extensions/v1beta1", "PodSecurityPolicy"),
		makeUnstructured("v1", "Secret"),
		makeUnstructured("v1", "ConfigMap"),
		makeUnstructured("storage.k8s.io/v1beta1", "StorageClass"),
		makeUnstructured("v1", "PersistentVolume"),
		makeUnstructured("v1", "PersistentVolumeClaim"),
		makeUnstructured("v1", "ServiceAccount"),
		makeUnstructured("rbac.authorization.k8s.io/v1", "CustomResourceDefinition"),
		makeUnstructured("rbac.authorization.k8s.io/v1", "ClusterRole"),
		makeUnstructured("rbac.authorization.k8s.io/v1", "ClusterRoleBinding"),
		makeUnstructured("rbac.authorization.k8s.io/v1", "Role"),
		makeUnstructured("rbac.authorization.k8s.io/v1", "RoleBinding"),
		makeUnstructured("v1", "Service"),
		makeUnstructured("apps/v1", "DaemonSet"),
		makeUnstructured("core/v1", "Pod"),
		makeUnstructured("core/v1", "ReplicationController"),
		makeUnstructured("apps/v1", "ReplicaSet"),
		makeUnstructured("apps/v1", "Deployment"),
		makeUnstructured("apps/v1", "StatefulSet"),
		makeUnstructured("batch/v1", "Job"),
		makeUnstructured("batch/v1beta1", "CronJob"),
		makeUnstructured("extensions/v1beta1", "Ingress"),
		makeUnstructured("apiregistration.k8s.io/v1beta1", "APIService"),
		makeUnstructured("v1", "UnknownGoesToTheEnd"),
		makeUnstructured("v1", "UnknownGoesToTheEnd-b"),
	}

	UnstructuredSlice(objects).Sort()

	require.Equal(t, expected, objects)
}

func Test_UnstructuredSlice_Sort_Namespaces(t *testing.T) {
	objects := []*unstructured.Unstructured{
		makeUnstructuredNS("v1", "Deployment", "first"),
		makeUnstructuredNS("v1", "Service", "first"),
		makeUnstructuredNS("v1", "ConfigMap", "first"),
		makeUnstructuredNS("v1", "ConfigMap", "second"),
		makeUnstructuredNS("v1", "Service", "second"),
		makeUnstructuredNS("v1", "Deployment", "second"),
		makeUnstructuredName("v1", "Namespace", "second"),
		makeUnstructuredName("v1", "Namespace", "first"),
	}

	expected := []*unstructured.Unstructured{
		makeUnstructuredName("v1", "Namespace", "first"),
		makeUnstructuredName("v1", "Namespace", "second"),
		makeUnstructuredNS("v1", "ConfigMap", "first"),
		makeUnstructuredNS("v1", "Service", "first"),
		makeUnstructuredNS("v1", "Deployment", "first"),
		makeUnstructuredNS("v1", "ConfigMap", "second"),
		makeUnstructuredNS("v1", "Service", "second"),
		makeUnstructuredNS("v1", "Deployment", "second"),
	}

	UnstructuredSlice(objects).Sort()

	require.Equal(t, expected, objects)
}

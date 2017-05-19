package utils

import (
	"testing"

	"k8s.io/client-go/pkg/runtime"
)

func newObj(apiVersion, kind string) *runtime.Unstructured {
	return &runtime.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": apiVersion,
			"kind":       kind,
		},
	}
}

func TestSort(t *testing.T) {
	objs := []*runtime.Unstructured{
		newObj("extensions/v1beta1", "Deployment"),
		newObj("v1", "ConfigMap"),
		newObj("v1", "Namespace"),
		newObj("v1", "Service"),
	}

	SortDepFirst(objs)

	if objs[0].GetKind() != "Namespace" {
		t.Error("Namespace should be sorted first")
	}
	if objs[3].GetKind() != "Deployment" {
		t.Error("Deployment should be sorted after other objects")
	}
}

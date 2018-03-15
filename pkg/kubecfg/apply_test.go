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

package kubecfg

import (
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/ksonnet/ksonnet/utils"
)

func TestStringListContains(t *testing.T) {
	foobar := []string{"foo", "bar"}
	if stringListContains([]string{}, "") {
		t.Error("Empty list was not empty")
	}
	if !stringListContains(foobar, "foo") {
		t.Error("Failed to find foo")
	}
	if stringListContains(foobar, "baz") {
		t.Error("Should not contain baz")
	}
}

func TestEligibleForGc(t *testing.T) {
	const myTag = "my-gctag"
	boolTrue := true
	o := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "tests/v1alpha1",
			"kind":       "Dummy",
		},
	}

	if eligibleForGc(o, myTag) {
		t.Errorf("%v should not be eligible (no tag)", o)
	}

	utils.SetMetaDataAnnotation(o, AnnotationGcTag, "unknowntag")
	if eligibleForGc(o, myTag) {
		t.Errorf("%v should not be eligible (wrong tag)", o)
	}

	utils.SetMetaDataAnnotation(o, AnnotationGcTag, myTag)
	if !eligibleForGc(o, myTag) {
		t.Errorf("%v should be eligible", o)
	}

	utils.SetMetaDataAnnotation(o, AnnotationGcStrategy, GcStrategyIgnore)
	if eligibleForGc(o, myTag) {
		t.Errorf("%v should not be eligible (strategy=ignore)", o)
	}

	utils.SetMetaDataAnnotation(o, AnnotationGcStrategy, GcStrategyAuto)
	if !eligibleForGc(o, myTag) {
		t.Errorf("%v should be eligible (strategy=auto)", o)
	}

	// Unstructured.SetOwnerReferences is broken in apimachinery release-1.6
	// See kubernetes/kubernetes#46817
	setOwnerRef := func(u *unstructured.Unstructured, ref metav1.OwnerReference) {
		// This is not a complete nor robust reimplementation
		c := map[string]interface{}{
			"kind": ref.Kind,
			"name": ref.Name,
		}
		if ref.Controller != nil {
			c["controller"] = *ref.Controller
		}
		u.Object["metadata"].(map[string]interface{})["ownerReferences"] = []map[string]interface{}{c}
	}
	setOwnerRef(o, metav1.OwnerReference{Kind: "foo", Name: "bar"})
	if !eligibleForGc(o, myTag) {
		t.Errorf("%v should be eligible (non-controller ownerref)", o)
	}

	setOwnerRef(o, metav1.OwnerReference{Kind: "foo", Name: "bar", Controller: &boolTrue})
	if eligibleForGc(o, myTag) {
		t.Errorf("%v should not be eligible (controller ownerref)", o)
	}
}

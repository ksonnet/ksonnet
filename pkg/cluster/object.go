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
	"encoding/json"

	"github.com/pkg/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/clientcmd"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/kubectl/resource"
)

// SetMetaDataLabel sets a label value
func SetMetaDataLabel(obj metav1.Object, key, value string) {
	labels := obj.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}
	labels[key] = value
	obj.SetLabels(labels)
}

// SetMetaDataAnnotation sets an annotation value
func SetMetaDataAnnotation(obj metav1.Object, key, value string) {
	a := obj.GetAnnotations()
	if a == nil {
		a = make(map[string]string)
	}
	a[key] = value
	obj.SetAnnotations(a)
}

// DefaultResourceInfo fetches objects from the cluster.
func DefaultResourceInfo(namespace string, config clientcmd.ClientConfig) *resource.Result {
	factory := kcmdutil.NewFactory(config)

	return factory.NewBuilder().
		Unstructured().
		NamespaceParam(namespace).
		ExportParam(false).
		ResourceTypeOrNameArgs(true, "all").
		LabelSelectorParam("app.kubernetes.io/deploy-manager=ksonnet").
		ContinueOnError().
		Flatten().
		IncludeUninitialized(false).
		RequireObject(true).
		Latest().
		Do()
}

type ResourceInfo interface {
	Err() error
	Infos() ([]*resource.Info, error)
}

var _ ResourceInfo = (*resource.Result)(nil)

// ManagedObjects returns a slice of ksonnet managed objects.
func ManagedObjects(r ResourceInfo) ([]*unstructured.Unstructured, error) {
	if err := r.Err(); err != nil {
		return nil, errors.WithStack(err)
	}

	infos, err := r.Infos()
	if err != nil {
		return nil, errors.WithStack(err)
	}

	lookup := make(map[types.UID]map[string]interface{})

	var objects []*unstructured.Unstructured

	for i := range infos {
		if obj, ok := infos[i].Object.(*unstructured.Unstructured); ok {
			if _, ok := lookup[obj.GetUID()]; ok {
				// we've seen this object already
				continue
			}

			lookup[obj.GetUID()] = obj.Object
			objects = append(objects, obj)
		}
	}

	return objects, nil
}

// RebuildObject rebuilds the ksonnet generated object from an object on
// the cluster.
func RebuildObject(m map[string]interface{}) (map[string]interface{}, error) {
	metadata, ok := m["metadata"].(map[string]interface{})
	if !ok {
		return nil, errors.New("metadata not found")
	}
	annotations, ok := metadata["annotations"].(map[string]interface{})
	if !ok {
		return nil, errors.New("metadata annotations not found")
	}
	descriptor, ok := annotations[annotationManaged].(string)
	if !ok {
		return m, nil
	}

	var mm managedMetadata
	if err := json.Unmarshal([]byte(descriptor), &mm); err != nil {
		return nil, errors.WithStack(err)
	}

	return mm.DecodePristine()
}

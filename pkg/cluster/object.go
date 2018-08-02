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

	clustermetadata "github.com/ksonnet/ksonnet/pkg/metadata"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/discovery"
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

// Resoucetypes to exclude when fetching managed objects
var unmanagedKinds = map[string]bool{
	"ComponentStatus": true,
	"Endpoints":       true,
}

// DefaultResourceInfo fetches objects from the cluster.
func fetchManagedObjects(namespace string, clients Clients, components []string) ([]*unstructured.Unstructured, error) {
	log := log.WithFields(log.Fields{
		"action":    "fetchManagedObjects",
		"namespace": namespace,
	})
	if clients.discovery == nil {
		return nil, errors.New("nil discovery client")
	}
	if clients.clientPool == nil {
		return nil, errors.New("nil client pool")
	}

	// TODO address cluster-wide-resources defined in other environments (see ServerPreferredResources)
	resources, err := clients.discovery.ServerPreferredNamespacedResources()
	if err != nil {
		return nil, errors.Wrap(err, "ServerPreferredNamespacedResources")
	}
	sortResources(resources) // Sift "extensions" to the end because it duplicates resources, e.g. Deployments

	// Filter out resources we can't list
	filtered := discovery.FilteredBy(
		discovery.ResourcePredicateFunc(
			func(groupVersion string, r *metav1.APIResource) bool {
				return (!unmanagedKinds[r.Kind]) &&
					discovery.SupportsAllVerbs{Verbs: []string{"list", "get"}}.Match(groupVersion, r)
			},
		),
		resources,
	)

	uids := make(map[types.UID]bool)
	results := make([]*unstructured.Unstructured, 0)

	for _, lst := range filtered {
		gv, err := schema.ParseGroupVersion(lst.GroupVersion)
		if err != nil {
			return nil, errors.Wrapf(err, "parsing GroupVersion: %s", lst.GroupVersion)
		}

		for _, resource := range lst.APIResources {
			// Create a dynamic client for this resource type
			gvr := gv.WithKind(resource.Kind)
			dynamic, err := clients.clientPool.ClientForGroupVersionKind(gvr)
			log.Debugf("listing resources: %s", gvr.String())
			if err != nil {
				return nil, errors.Wrapf(err, "creating client for resource: %s", gvr.String())
			}
			resourceClient := dynamic.Resource(&resource, namespace)

			// List managed resources of this type from the cluster
			obj, err := resourceClient.List(metav1.ListOptions{
				LabelSelector: "app.kubernetes.io/deploy-manager=ksonnet",
			})
			if err != nil {
				log.Warnf("skipping %s due to error: %v", resource.Kind, err)
				continue
			}

			if ul, ok := obj.(*unstructured.UnstructuredList); ok {
				if err := ul.EachListItem(func(o runtime.Object) error {
					if u, ok := o.(*unstructured.Unstructured); ok {
						// Filter out duplicates, e.g apps/v1/Deployment vs. extensions/v1beta1/Deployment
						if uids[u.GetUID()] {
							return nil
						}

						uids[u.GetUID()] = true
						results = append(results, u)
					}
					return nil
				}); err != nil {
					return nil, errors.Wrapf(err, "iterating %s", resource.Kind)
				}
			}
		}
	}
	return results, nil
}

// ResourceInfo holds information about cluster resources.
type ResourceInfo interface {
	Err() error
	Infos() ([]*resource.Info, error)
}

var _ ResourceInfo = (*resource.Result)(nil)

// RebuildObject rebuilds the ksonnet generated object from an object on
// the cluster.
func RebuildObject(m map[string]interface{}) (map[string]interface{}, error) {
	metadata, ok := m["metadata"].(map[string]interface{})
	if !ok {
		return nil, errors.New("metadata not found")
	}
	annotations, ok := metadata["annotations"].(map[string]interface{})
	if !ok {
		return nil, errors.Errorf("metadata annotations not found: %v", metadata)
	}
	descriptor, ok := annotations[clustermetadata.AnnotationManaged].(string)
	if !ok {
		return m, nil
	}

	var mm managedAnnotation
	if err := json.Unmarshal([]byte(descriptor), &mm); err != nil {
		return nil, errors.WithStack(err)
	}

	return mm.Decode()
}

// filterManagedObjects filters out any non-managed objects according to their labels
func filterManagedObjects(objects []*unstructured.Unstructured) []*unstructured.Unstructured {
	// see Filtering without allocating - https://github.com/golang/go/wiki/SliceTricks#filtering-without-allocating
	filtered := objects[:0]
	for _, o := range objects {
		labels := o.GetLabels()
		if labels == nil {
			continue
		}
		if labels["app.kubernetes.io/deploy-manager"] == "ksonnet" {
			filtered = append(filtered, o)
		}
	}

	return filtered
}

// CollectObjects collects objects in a cluster namespace.
func CollectObjects(namespace string, clients Clients, components []string) ([]*unstructured.Unstructured, error) {
	objects, err := fetchManagedObjects(namespace, clients, components)
	if err != nil {
		return nil, err
	}
	objects = filterManagedObjects(objects)
	if err != nil {
		return nil, err
	}

	for _, obj := range objects {
		m, err := RebuildObject(obj.Object)
		if err != nil {
			return nil, err
		}

		obj.Object = m
	}

	return objects, nil
}

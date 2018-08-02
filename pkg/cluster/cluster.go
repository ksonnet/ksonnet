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
	"fmt"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/client"
	"github.com/ksonnet/ksonnet/pkg/metadata"
	"github.com/ksonnet/ksonnet/pkg/pipeline"
	"github.com/ksonnet/ksonnet/utils"
	log "github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"
)

type genClientOptsFn func(a app.App, c *client.Config, envName string) (Clients, error)

type clientFactoryFn func(dynamic.ClientPool, discovery.DiscoveryInterface, runtime.Object, string) (dynamic.ResourceInterface, error)

type resourceClientFactoryFn func(opts Clients, object runtime.Object) (ResourceClient, error)

type discoveryFn func(a app.App, clientConfig *client.Config, envName string) (discovery.DiscoveryInterface, error)

type validateObjectFn func(d discovery.DiscoveryInterface,
	obj *unstructured.Unstructured) []error

type findObjectsFn func(a app.App, envName string,
	componentNames []string) ([]*unstructured.Unstructured, error)

func findObjects(a app.App, envName string, componentNames []string) ([]*unstructured.Unstructured, error) {
	p := pipeline.New(a, envName)
	return p.Objects(componentNames)
}

func stringListContains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

// func gcDelete(clientpool dynamic.ClientPool, disco discovery.DiscoveryInterface, version *utils.ServerVersion, o runtime.Object) error {
func gcDelete(options Clients, rcFactory resourceClientFactoryFn, version *utils.ServerVersion, o runtime.Object) error {
	obj, err := meta.Accessor(o)
	if err != nil {
		return fmt.Errorf("Unexpected object type: %s", err)
	}

	uid := obj.GetUID()
	desc := fmt.Sprintf("%s %s", utils.ResourceNameFor(options.discovery, o), utils.FqName(obj))

	deleteOpts := metav1.DeleteOptions{
		Preconditions: &metav1.Preconditions{UID: &uid},
	}
	if version.Compare(1, 6) < 0 {
		// 1.5.x option
		boolFalse := false
		deleteOpts.OrphanDependents = &boolFalse
	} else {
		// 1.6.x option (NB: Background is broken)
		fg := metav1.DeletePropagationForeground
		deleteOpts.PropagationPolicy = &fg
	}

	rc, err := rcFactory(options, o)
	if err != nil {
		return err
	}

	err = rc.Delete(&deleteOpts)
	if err != nil && (kerrors.IsNotFound(err) || kerrors.IsConflict(err)) {
		// We lost a race with something else changing the object
		log.Debugf("Ignoring error while deleting %s: %s", desc, err)
		err = nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting %s: %s", desc, err)
	}

	return nil
}

func walkObjects(co Clients, listopts metav1.ListOptions, callback func(runtime.Object) error) error {
	rsrclists, err := co.discovery.ServerResources()
	if err != nil {
		return err
	}
	for _, rsrclist := range rsrclists {
		gv, err := schema.ParseGroupVersion(rsrclist.GroupVersion)
		if err != nil {
			return err
		}
		for _, rsrc := range rsrclist.APIResources {
			gvk := gv.WithKind(rsrc.Kind)

			if !stringListContains(rsrc.Verbs, "list") {
				log.Debugf("Don't know how to list %v, skipping", rsrc)
				continue
			}
			client, err := co.clientPool.ClientForGroupVersionKind(gvk)
			if err != nil {
				return err
			}

			var ns string
			if rsrc.Namespaced {
				ns = metav1.NamespaceAll
			} else {
				ns = metav1.NamespaceNone
			}

			rc := client.Resource(&rsrc, ns)
			log.Debugf("Listing %s", gvk)
			obj, err := rc.List(listopts)
			if err != nil {
				return err
			}
			if err = meta.EachListItem(obj, callback); err != nil {
				return err
			}
		}
	}
	return nil
}

func eligibleForGc(obj metav1.Object, gcTag string) bool {
	for _, ref := range obj.GetOwnerReferences() {
		if ref.Controller != nil && *ref.Controller {
			// Has a controller ref
			return false
		}
	}

	a := obj.GetAnnotations()

	strategy, ok := a[metadata.AnnotationGcStrategy]
	if !ok {
		strategy = metadata.GcStrategyAuto
	}

	return a[metadata.AnnotationGcTag] == gcTag &&
		strategy == metadata.GcStrategyAuto
}

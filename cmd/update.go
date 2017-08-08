// Copyright 2017 The kubecfg authors
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

package cmd

import (
	"encoding/json"
	"fmt"
	"sort"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/diff"
	"k8s.io/apimachinery/pkg/util/sets"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/dynamic"

	"github.com/ksonnet/kubecfg/utils"
)

const (
	flagCreate = "create"
	flagSkipGc = "skip-gc"
	flagGcTag  = "gc-tag"
	flagDryRun = "dry-run"

	// AnnotationGcTag annotation that triggers
	// garbage collection. Objects with value equal to
	// command-line flag that are *not* in config will be deleted.
	AnnotationGcTag = "kubecfg.ksonnet.io/garbage-collect-tag"

	// AnnotationGcStrategy controls gc logic.  Current values:
	// `auto` (default if absent) - do garbage collection
	// `ignore` - never garbage collect this object
	AnnotationGcStrategy = "kubecfg.ksonnet.io/garbage-collect-strategy"

	// GcStrategyAuto is the default automatic gc logic
	GcStrategyAuto = "auto"
	// GcStrategyIgnore means this object should be ignored by garbage collection
	GcStrategyIgnore = "ignore"
)

func init() {
	RootCmd.AddCommand(updateCmd)
	updateCmd.PersistentFlags().Bool(flagCreate, true, "Create missing resources")
	updateCmd.PersistentFlags().Bool(flagSkipGc, false, "Don't perform garbage collection, even with --"+flagGcTag)
	updateCmd.PersistentFlags().String(flagGcTag, "", "Add this tag to updated objects, and garbage collect existing objects with this tag and not in config")
	updateCmd.PersistentFlags().Bool(flagDryRun, false, "Perform only read-only operations")
}

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update Kubernetes resources with local config",
	RunE: func(cmd *cobra.Command, args []string) error {
		flags := cmd.Flags()

		create, err := flags.GetBool(flagCreate)
		if err != nil {
			return err
		}

		gcTag, err := flags.GetString(flagGcTag)
		if err != nil {
			return err
		}

		skipGc, err := flags.GetBool(flagSkipGc)
		if err != nil {
			return err
		}

		dryRun, err := flags.GetBool(flagDryRun)
		if err != nil {
			return err
		}

		clientpool, disco, err := restClientPool(cmd)
		if err != nil {
			return err
		}

		defaultNs, _, err := clientConfig.Namespace()
		if err != nil {
			return err
		}

		objs, err := readObjs(cmd, args)
		if err != nil {
			return err
		}

		dryRunText := ""
		if dryRun {
			dryRunText = " (dry-run)"
		}

		sort.Sort(utils.DependencyOrder(objs))

		seenUids := sets.NewString()

		for _, obj := range objs {
			if gcTag != "" {
				utils.SetMetaDataAnnotation(obj, AnnotationGcTag, gcTag)
			}

			desc := fmt.Sprintf("%s/%s", obj.GetKind(), fqName(obj))
			log.Info("Updating ", desc, dryRunText)

			rc, err := clientForResource(clientpool, disco, obj, defaultNs)
			if err != nil {
				return err
			}

			asPatch, err := json.Marshal(obj)
			if err != nil {
				return err
			}
			var newobj metav1.Object
			if !dryRun {
				newobj, err = rc.Patch(obj.GetName(), types.MergePatchType, asPatch)
				log.Debug("Patch(%s) returned (%v, %v)", obj.GetName(), newobj, err)
			} else {
				newobj, err = rc.Get(obj.GetName())
			}
			if create && errors.IsNotFound(err) {
				log.Info(" Creating non-existent ", desc, dryRunText)
				if !dryRun {
					newobj, err = rc.Create(obj)
					log.Debug("Create(%s) returned (%v, %v)", obj.GetName(), newobj, err)
				} else {
					newobj = obj
					err = nil
				}
			}
			if err != nil {
				// TODO: retry
				return fmt.Errorf("Error updating %s: %s", desc, err)
			}

			log.Debug("Updated object: ", diff.ObjectDiff(obj, newobj))

			// Some objects appear under multiple kinds
			// (eg: Deployment is both extensions/v1beta1
			// and apps/v1beta1).  UID is the only stable
			// identifier that links these two views of
			// the same object.
			seenUids.Insert(string(newobj.GetUID()))
		}

		if gcTag != "" && !skipGc {
			version, err := utils.FetchVersion(disco)
			if err != nil {
				return err
			}

			err = walkObjects(clientpool, disco, metav1.ListOptions{}, func(o runtime.Object) error {
				meta, err := meta.Accessor(o)
				if err != nil {
					return err
				}
				gvk := o.GetObjectKind().GroupVersionKind()
				desc := fmt.Sprintf("%s/%s (%s)", gvk.Kind, fqName(meta), gvk.GroupVersion())
				log.Debugf("Considering %v for gc", desc)
				if eligibleForGc(meta, gcTag) && !seenUids.Has(string(meta.GetUID())) {
					log.Info("Garbage collecting ", desc, dryRunText)
					if !dryRun {
						err := gcDelete(clientpool, disco, &version, o)
						if err != nil {
							return err
						}
					}
				}
				return nil
			})
			if err != nil {
				return err
			}
		}

		return nil
	},
}

func fqName(o metav1.Object) string {
	if o.GetNamespace() == "" {
		return o.GetName()
	}
	return fmt.Sprintf("%s.%s", o.GetNamespace(), o.GetName())
}

func stringListContains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}

func gcDelete(clientpool dynamic.ClientPool, disco discovery.DiscoveryInterface, version *utils.ServerVersion, o runtime.Object) error {
	obj, err := meta.Accessor(o)
	if err != nil {
		return fmt.Errorf("Unexpected object type: %s", err)
	}

	uid := obj.GetUID()
	desc := fmt.Sprintf("%s/%s", o.GetObjectKind().GroupVersionKind().Kind, fqName(obj))

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

	c, err := clientForResource(clientpool, disco, o, metav1.NamespaceNone)
	if err != nil {
		return err
	}

	err = c.Delete(obj.GetName(), &deleteOpts)
	if err != nil && (errors.IsNotFound(err) || errors.IsConflict(err)) {
		// We lost a race with something else changing the object
		log.Debugf("Ignoring error while deleting %s: %s", desc, err)
		err = nil
	}
	if err != nil {
		return fmt.Errorf("Error deleting %s: %s", desc, err)
	}

	return nil
}

func walkObjects(pool dynamic.ClientPool, disco discovery.DiscoveryInterface, listopts metav1.ListOptions, callback func(runtime.Object) error) error {
	rsrclists, err := disco.ServerResources()
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
			client, err := pool.ClientForGroupVersionKind(gvk)
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

	strategy, ok := a[AnnotationGcStrategy]
	if !ok {
		strategy = GcStrategyAuto
	}

	return a[AnnotationGcTag] == gcTag &&
		strategy == GcStrategyAuto
}

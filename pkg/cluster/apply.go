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
	"sort"
	"time"

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/client"
	"github.com/ksonnet/ksonnet/pkg/metadata"
	"github.com/ksonnet/ksonnet/utils"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/sets"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

const (
	// applyConflictRetryCount sets how many times an apply is retried before giving up
	// after a conflict error is detected.
	applyConflictRetryCount = 5

	// defaultConflictTimeout sets the wait time before retrying after a conflict is detected.
	defaultConflictTimeout = 1 * time.Second

	appKsonnet = "ksonnet"
)

var (
	errApplyConflict = errors.Errorf("apply conflict detected; retried %d times", applyConflictRetryCount)
)

// ApplyConfig is configuration for Apply.
type ApplyConfig struct {
	App            app.App
	ClientConfig   *client.Config
	ComponentNames []string
	Create         bool
	DryRun         bool
	EnvName        string
	GcTag          string
	SkipGc         bool
}

// ApplyOpts are options for configuring Apply.
type ApplyOpts func(a *Apply)

// Apply applies objects to the cluster
type Apply struct {
	ApplyConfig

	// these make it easier to test Apply.
	findObjectsFn         findObjectsFn
	resourceClientFactory resourceClientFactoryFn
	clientOpts            *Clients
	objectInfo            ObjectInfo
	ksonnetObjectFactory  func() ksonnetObject
	upserterFactory       func() Upserter
	conflictTimeout       time.Duration
}

// RunApply runs apply against a cluster given a configuration.
func RunApply(config ApplyConfig, opts ...ApplyOpts) error {
	if config.ClientConfig == nil {
		return errors.New("ksonnet client config is required")
	}

	a := &Apply{
		ApplyConfig:           config,
		findObjectsFn:         findObjects,
		resourceClientFactory: resourceClientFactory,
		objectInfo:            &objectInfo{},
		ksonnetObjectFactory: func() ksonnetObject {
			factory := cmdutil.NewFactory(config.ClientConfig.Config)
			return newDefaultKsonnetObject(factory, config.DryRun)
		},
		conflictTimeout: 1 * time.Second,
	}

	for _, opt := range opts {
		opt(a)
	}

	if a.clientOpts == nil {
		co, err := GenClients(a.App, a.ClientConfig, a.EnvName)
		if err != nil {
			return err
		}

		a.clientOpts = &co
	}

	if a.upserterFactory == nil {
		u, err := newDefaultUpserter(a.ApplyConfig, a.objectInfo, *a.clientOpts, a.resourceClientFactory)
		if err != nil {
			return errors.Wrap(err, "creating upserter")
		}
		a.upserterFactory = func() Upserter {
			return u
		}
	}

	return a.Apply()
}

// Apply applies against a cluster.
func (a *Apply) Apply() error {
	apiObjects, err := a.findObjectsFn(a.App, a.EnvName, a.ComponentNames)
	if err != nil {
		return errors.Wrap(err, "find objects")
	}

	sort.Sort(utils.DependencyOrder(apiObjects))

	seenUids := sets.NewString()

	for _, obj := range apiObjects {
		var uid string
		uid, err = a.handleObject(obj)
		if err != nil {
			return errors.Wrap(err, "handle object")
		}

		// Some objects appear under multiple kinds
		// (eg: Deployment is both extensions/v1beta1
		// and apps/v1beta1).  UID is the only stable
		// identifier that links these two views of
		// the same object.
		seenUids.Insert(uid)
	}

	if a.GcTag != "" && !a.SkipGc {
		if err = a.runGc(seenUids); err != nil {
			return errors.Wrap(err, "run gc")
		}
	}

	return nil
}

func (a *Apply) handleObject(obj *unstructured.Unstructured) (string, error) {
	if err := a.preprocessObject(obj); err != nil {
		return "", errors.Wrap(err, "preprocessing object before apply")
	}

	mergedObject, err := a.patchFromCluster(obj)
	if err != nil {
		return "", errors.Wrap(err, "patching object from cluster")
	}

	a.setupGC(mergedObject)

	return a.upsert(mergedObject)
}

// preprocessObject preprocesses an object for it is applied to the cluster.
func (a *Apply) preprocessObject(obj *unstructured.Unstructured) error {
	aa := newDefaultAnnotationApplier()
	if !a.DryRun {
		return errors.Wrap(aa.SetOriginalConfiguration(obj), "tagging ksonnet managed object")
	}

	log.Info("tagging ksonnet managed object", a.dryRunText())
	return nil
}

// patchFromCluster patches an object with values that may exist in the cluster.
func (a *Apply) patchFromCluster(obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return a.ksonnetObjectFactory().MergeFromCluster(*a.clientOpts, obj)
}

func (a *Apply) upsert(obj *unstructured.Unstructured) (string, error) {
	if a.DryRun {
		log.Info("upserting object", a.dryRunText())
		return "12345", nil
	}

	u := a.upserterFactory()

	for i := applyConflictRetryCount; i > 0; i-- {
		uid, err := u.Upsert(obj)
		if err != nil {
			cause := errors.Cause(err)
			if !kerrors.IsConflict(cause) {
				return "", err
			}
			// In order for the next try to work, update the resource version on the object
			updatedObj, err := a.getUpdatedObject(obj)
			if err == nil {
				obj.SetResourceVersion(updatedObj.GetResourceVersion())
			}

			time.Sleep(a.conflictTimeout)
			continue
		}

		return uid, nil
	}

	return "", errApplyConflict
}

func (a *Apply) getUpdatedObject(obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	rc, err := a.resourceClientFactory(*a.clientOpts, obj)
	if err != nil {
		return nil, err
	}
	return rc.Get(metav1.GetOptions{})
}

// setupGC setups ksonnet's garbage collection process for objects.
func (a *Apply) setupGC(obj *unstructured.Unstructured) {
	if a.GcTag != "" {
		SetMetaDataAnnotation(obj, metadata.AnnotationGcTag, a.GcTag)
	}
}

func (a *Apply) runGc(seenUids sets.String) error {
	co := a.clientOpts

	version, err := utils.FetchVersion(co.discovery)
	if err != nil {
		return err
	}

	err = walkObjects(*co, metav1.ListOptions{}, func(o runtime.Object) error {
		var metav1Object metav1.Object
		metav1Object, err = meta.Accessor(o)
		if err != nil {
			return err
		}
		gvk := o.GetObjectKind().GroupVersionKind()
		desc := fmt.Sprintf("%s %s (%s)",
			utils.ResourceNameFor(co.discovery, o), utils.FqName(metav1Object), gvk.GroupVersion())
		log.Debugf("Considering %v for gc", desc)
		if eligibleForGc(metav1Object, a.GcTag) && !seenUids.Has(string(metav1Object.GetUID())) {
			log.Info("Garbage collecting ", desc, a.dryRunText())
			if !a.DryRun {
				err = gcDelete(*co, a.resourceClientFactory, &version, o)
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

	return nil
}

func (a *Apply) dryRunText() string {
	text := ""
	if a.DryRun {
		text = " (dry-run)"
	}

	return text
}

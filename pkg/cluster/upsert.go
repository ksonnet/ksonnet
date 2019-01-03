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
	log "github.com/sirupsen/logrus"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	kdiff "k8s.io/apimachinery/pkg/util/diff"
)

// Upserter updates or creates objects.
type Upserter interface {
	// Upsert updates or creates an object.
	Upsert(*unstructured.Unstructured) (string, error)
}

// defaultUpserter is the default implementation for updating or creating objects.
type defaultUpserter struct {
	// ApplyConfig is configuration values for applying objects to a cluster.
	ApplyConfig

	// clientOpts are Kubernetes client options.
	clientOpts Clients

	// objectInfo is the utility for location information about objects.
	objectInfo ObjectInfo

	// resourceClientFactory is a factory for creating clients for resources.
	resourceClientFactory resourceClientFactoryFn

	// objectDescriber describes an object.
	objectDescriber objectDescriber
}

var _ Upserter = (*defaultUpserter)(nil)

// newDefaultUpserter creates an instance of defaultUpserter.
func newDefaultUpserter(ac ApplyConfig, oi ObjectInfo, co Clients, rfc resourceClientFactoryFn) (*defaultUpserter, error) {
	describer, err := newDefaultObjectDescriber(co, oi)
	if err != nil {
		return nil, errors.Wrap(err, "creating object describer")
	}

	return &defaultUpserter{
		ApplyConfig:           ac,
		objectInfo:            oi,
		clientOpts:            co,
		resourceClientFactory: rfc,
		objectDescriber:       describer,
	}, nil
}

// Upsert updates or creates an object.
func (u *defaultUpserter) Upsert(obj *unstructured.Unstructured) (string, error) {
	log.Info("Applying ", u.objectDescriber.Describe(obj), u.dryRunText())

	rc, err := u.resourceClientFactory(u.clientOpts, obj)
	if err != nil {
		return "", err
	}

	patchedObject, err := u.updateObject(rc, obj)
	if err == nil {
		log.Debug("Updated object: ", kdiff.ObjectDiff(obj, patchedObject))
		return string(patchedObject.GetUID()), nil
	} else if !kerrors.IsNotFound(err) {
		return "", errors.Wrap(err, "patching existing object")
	}

	if !u.Create {
		return "", errors.New("not creating non-existent object")
	}

	log.Info("Creating non-existent ", u.objectDescriber.Describe(obj), u.dryRunText())
	newObj, err := u.createObject(u.clientOpts, rc, obj)
	if err != nil {
		return "", errors.Wrap(err, "creating object")
	}

	log.Debug("Created object: ", kdiff.ObjectDiff(obj, newObj))
	return string(newObj.GetUID()), nil
}

// updateObject attempts to update an object in the cluster.
func (u *defaultUpserter) updateObject(rc ResourceClient, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	objectData, err := json.Marshal(obj)
	if err != nil {
		return nil, err
	}

	if u.DryRun {
		log.Info("skipping patch (dry-run)")
		return obj, nil
	}

	return rc.Patch(types.MergePatchType, objectData)
}

// createObject attempts to create an object in the cluster.
func (u *defaultUpserter) createObject(co Clients, rc ResourceClient, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	newObj, err := rc.Create()
	log.Debugf("Create(%s) returned (%v, %v)", obj.GetName(), newObj, err)

	if err != nil {
		return nil, errors.Wrap(err, "creating object")
	}

	return newObj, nil
}

func (u *defaultUpserter) dryRunText() string {
	text := ""
	if u.DryRun {
		text = " (dry-run)"
	}

	return text
}

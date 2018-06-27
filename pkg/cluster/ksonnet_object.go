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
	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	cmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
)

// ksonnetObject can merge an object with its cluster state. This is required because
// some fields will be overwritten if applied again (e.g. Server NodePort).
type ksonnetObject interface {
	MergeFromCluster(co clientOpts, obj *unstructured.Unstructured) (*unstructured.Unstructured, error)
}

type defaultKsonnetObject struct {
	objectMerger objectMerger
}

var _ ksonnetObject = (*defaultKsonnetObject)(nil)

func newDefaultKsonnetObject(factory cmdutil.Factory) *defaultKsonnetObject {
	merger := newDefaultObjectMerger(factory)

	return &defaultKsonnetObject{
		objectMerger: merger,
	}
}

func (ko *defaultKsonnetObject) MergeFromCluster(co clientOpts, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	mergedObject, err := ko.objectMerger.Merge(co.namespace, obj)
	if err != nil {
		cause := errors.Cause(err)
		if !kerrors.IsNotFound(cause) {
			return nil, errors.Wrap(cause, "merging object with existing state")
		}
		mergedObject = obj
	}

	return mergedObject, nil
}

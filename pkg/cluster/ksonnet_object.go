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

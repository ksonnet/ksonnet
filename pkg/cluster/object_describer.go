package cluster

import (
	"fmt"

	"github.com/ksonnet/ksonnet/utils"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// objectDescriber describes an describe a Kubernetes object.
type objectDescriber interface {
	// Describe describes a Kubernetes object
	Describe(obj *unstructured.Unstructured) string
}

// defaultObjectDescriber is the default implementation of objectDescriber.
type defaultObjectDescriber struct {
	// clientOpts are Kubernetes client otpions.
	clientOpts clientOpts

	// objectInfo locates information for Kubernetes objects.
	objectInfo ObjectInfo
}

var _ objectDescriber = (*defaultObjectDescriber)(nil)

// newDefaultObjectDescriber creates an instance of defaultObjectDescriber.
func newDefaultObjectDescriber(co clientOpts, oi ObjectInfo) (*defaultObjectDescriber, error) {
	if oi == nil {
		return nil, errors.Errorf("object info is required")
	}

	return &defaultObjectDescriber{
		clientOpts: co,
		objectInfo: oi,
	}, nil
}

// Describe describes an object using its resource kind and fully qualified name.
func (od *defaultObjectDescriber) Describe(obj *unstructured.Unstructured) string {
	return fmt.Sprintf("%s %s", od.objectInfo.ResourceName(od.clientOpts.discovery, obj), utils.FqName(obj))
}

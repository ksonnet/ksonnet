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

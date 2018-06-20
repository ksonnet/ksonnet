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
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/discovery"
)

func Test_defaultObjectDescriber_Describe(t *testing.T) {
	co := clientOpts{}
	oi := &fakeObjectInfo{resourceName: "name"}

	od, err := newDefaultObjectDescriber(co, oi)
	require.NoError(t, err)

	obj := &unstructured.Unstructured{
		Object: genObject(),
	}

	got := od.Describe(obj)

	expected := "name guiroot"

	require.Equal(t, expected, got)
}

type fakeObjectInfo struct {
	resourceName string
}

var _ ObjectInfo = (*fakeObjectInfo)(nil)

func (oi *fakeObjectInfo) ResourceName(sri discovery.ServerResourcesInterface, o runtime.Object) string {
	return oi.resourceName
}

type fakeObjectDescriber struct {
	description string
}

var _ objectDescriber = (*fakeObjectDescriber)(nil)

func (od *fakeObjectDescriber) Describe(obj *unstructured.Unstructured) string {
	return od.description
}

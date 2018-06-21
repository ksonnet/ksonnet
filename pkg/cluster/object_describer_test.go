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

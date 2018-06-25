package cluster

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	cmdtesting "k8s.io/kubernetes/pkg/kubectl/cmd/testing"
)

type notFoundError struct{}

var _ kerrors.APIStatus = (*notFoundError)(nil)
var _ error = (*notFoundError)(nil)

func (e *notFoundError) Status() metav1.Status {
	return metav1.Status{
		Reason: metav1.StatusReasonNotFound,
	}
}

func (e *notFoundError) Error() string {
	return "not found"
}

func Test_defaultKsonnetObject_MergeFromCluster(t *testing.T) {

	sampleObj := &unstructured.Unstructured{
		Object: genObject(),
	}

	cases := []struct {
		name         string
		obj          *unstructured.Unstructured
		expected     *unstructured.Unstructured
		objectMerger *fakeObjectMerger
		isErr        bool
	}{
		{
			name: "merge object",
			obj:  sampleObj,
			objectMerger: &fakeObjectMerger{
				mergeObj: sampleObj,
			},
			expected: sampleObj,
		},
		{
			name: "unexpected error",
			obj:  sampleObj,
			objectMerger: &fakeObjectMerger{
				mergeErr: errors.Errorf("failed"),
			},
			isErr: true,
		},
		{
			name: "object not found",
			obj:  sampleObj,
			objectMerger: &fakeObjectMerger{
				mergeErr: &notFoundError{},
			},
			expected: sampleObj,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			factory := cmdtesting.NewTestFactory()
			defer factory.Cleanup()

			co := clientOpts{}

			ko := newDefaultKsonnetObject(factory)
			ko.objectMerger = tc.objectMerger

			merged, err := ko.MergeFromCluster(co, tc.obj)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, merged)
		})
	}
}

type fakeKsonnetObject struct {
	obj *unstructured.Unstructured
	err error
}

var _ (ksonnetObject) = (*fakeKsonnetObject)(nil)

func (ko *fakeKsonnetObject) MergeFromCluster(co clientOpts, obj *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return ko.obj, ko.err
}

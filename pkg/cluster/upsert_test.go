package cluster

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"

	"github.com/ksonnet/ksonnet/pkg/cluster/mocks"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/runtime"
)

func Test_defaultUpserter_Upsert(t *testing.T) {

	cases := []struct {
		name               string
		applyConfig        ApplyConfig
		initResourceClient func(*testing.T, *unstructured.Unstructured) *mocks.ResourceClient
		isErr              bool
		expectedID         string
	}{
		{
			name: "patch existing object",
			applyConfig: ApplyConfig{
				Create: true,
			},
			initResourceClient: func(t *testing.T, obj *unstructured.Unstructured) *mocks.ResourceClient {
				rc := &mocks.ResourceClient{}

				newObject := *obj
				newObject.SetUID(types.UID("12345"))

				rc.On("Patch", types.MergePatchType, mock.AnythingOfType("[]uint8")).Return(&newObject, nil)

				return rc
			},
			expectedID: "12345",
		},
		{
			name: "create new object",
			applyConfig: ApplyConfig{
				Create: true,
			},
			initResourceClient: func(t *testing.T, obj *unstructured.Unstructured) *mocks.ResourceClient {
				rc := &mocks.ResourceClient{}

				err := &notFoundError{}
				rc.On("Patch", types.MergePatchType, mock.AnythingOfType("[]uint8")).Return(nil, err)

				newObject := *obj
				newObject.SetUID(types.UID("12345"))

				rc.On("Create").Return(&newObject, nil)

				return rc
			},
			expectedID: "12345",
		},
		{
			name: "dry run create",
			applyConfig: ApplyConfig{
				Create: true,
				DryRun: true,
			},
			initResourceClient: func(t *testing.T, obj *unstructured.Unstructured) *mocks.ResourceClient {
				rc := &mocks.ResourceClient{}
				return rc
			},
		},
		{
			name: "patch error other than not found",
			applyConfig: ApplyConfig{
				Create: true,
			},
			initResourceClient: func(t *testing.T, obj *unstructured.Unstructured) *mocks.ResourceClient {
				rc := &mocks.ResourceClient{}

				rc.On("Patch", types.MergePatchType, mock.AnythingOfType("[]uint8")).Return(nil, errors.New("failed"))

				return rc
			},
			isErr: true,
		},
		{
			name: "patch only/no create",
			applyConfig: ApplyConfig{
				Create: false,
			},
			initResourceClient: func(t *testing.T, obj *unstructured.Unstructured) *mocks.ResourceClient {
				rc := &mocks.ResourceClient{}

				newObject := *obj
				newObject.SetUID(types.UID("12345"))

				rc.On("Patch", types.MergePatchType, mock.AnythingOfType("[]uint8")).Return(nil, &notFoundError{})

				return rc
			},
			isErr: true,
		},
		{
			name: "create failed",
			applyConfig: ApplyConfig{
				Create: true,
			},
			initResourceClient: func(t *testing.T, obj *unstructured.Unstructured) *mocks.ResourceClient {
				rc := &mocks.ResourceClient{}

				err := &notFoundError{}
				rc.On("Patch", types.MergePatchType, mock.AnythingOfType("[]uint8")).Return(nil, err)

				rc.On("Create").Return(nil, errors.New("failed"))

				return rc
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Fatal(r)
				}
			}()

			obj := &unstructured.Unstructured{
				Object: genObject(),
			}

			oi := &fakeObjectInfo{resourceName: "name"}

			co := clientOpts{}

			rc := tc.initResourceClient(t, obj)
			rfc := func(clientOpts, runtime.Object) (ResourceClient, error) {
				return rc, nil
			}

			u, err := newDefaultUpserter(tc.applyConfig, oi, co, rfc)
			require.NoError(t, err)

			id, err := u.Upsert(obj)

			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tc.expectedID, id)
		})
	}
}

type fakeUpserter struct {
	upsertID  string
	upsertErr error
}

var _ Upserter = (*fakeUpserter)(nil)

func (u *fakeUpserter) Upsert(*unstructured.Unstructured) (string, error) {
	return u.upsertID, u.upsertErr
}

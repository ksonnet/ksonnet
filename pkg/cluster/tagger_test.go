package cluster

import (
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func Test_defaultTagger_Tag(t *testing.T) {
	cases := []struct {
		name  string
		obj   *unstructured.Unstructured
		setup func(m *defaultTagger)
		isErr bool
	}{
		{
			name: "tag object",
			obj: &unstructured.Unstructured{
				Object: genObject(),
			},
		},
		{
			name:  "nil object",
			isErr: true,
		},
		{
			name: "nil encoder",
			obj: &unstructured.Unstructured{
				Object: genObject(),
			},
			setup: func(m *defaultTagger) {
				m.annotationEncoder = nil
			},
			isErr: true,
		},
		{
			name: "encode failure",
			obj: &unstructured.Unstructured{
				Object: genObject(),
			},
			setup: func(m *defaultTagger) {
				fm := &fakeAnnotationEncoder{
					err: errors.New("failure"),
				}
				m.annotationEncoder = fm
			},
			isErr: true,
		},
		{
			name: "marshal failure",
			obj: &unstructured.Unstructured{
				Object: genObject(),
			},
			setup: func(m *defaultTagger) {
				fm := &fakeAnnotationEncoder{
					marshalError: errors.New("failure"),
				}
				m.annotationEncoder = fm
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			m := newDefaultManaged()
			if tc.setup != nil {
				tc.setup(m)
			}

			err := m.Tag(tc.obj)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			metadata, ok := tc.obj.Object["metadata"].(map[string]interface{})
			require.True(t, ok)

			annotations, ok := metadata["annotations"].(map[string]interface{})
			require.True(t, ok)

			managed, ok := annotations["ksonnet.io/managed"].(string)
			require.True(t, ok)

			expected := "{\"pristine\":\"H4sIAAAAAAAA/1yPzWrsMAyF9/cxzjrzk93F6z5AV92UWSiJSE1iSdhKYQh+9+IMlNCV8RH6vqMdZPGDc4kqCCCzcvvuB3bq0WGJMiHgjW3VZ2JxdEjsNJETwg4SUSePKqV9l6Ii7Neot2lL6YmA11s7CCVGwLzFrOotKcZj28psaxypIPQdnJOt5NwGZ9NKA6+HhMzOnBNoVHGKwrkgfO6IieZDOebW6IvNo16OtNyWcpk3Lj6oLpeJk4b7tV38p2YH0+wv3i/+XbMj/L/XR33UWuu/HwAAAP//AQAA///Dx6kERQEAAA==\"}"
			require.Equal(t, expected, managed)
		})
	}
}

type fakeAnnotationEncoder struct {
	err error

	marshalBytes []byte
	marshalError error
}

func (m *fakeAnnotationEncoder) EncodePristine(in map[string]interface{}) error {
	return m.err
}

func (m *fakeAnnotationEncoder) Marshal() ([]byte, error) {
	return m.marshalBytes, m.marshalError
}

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
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	clientcmdapi "k8s.io/client-go/tools/clientcmd/api"
	"k8s.io/kubernetes/pkg/kubectl/resource"
)

func TestSetMetadataLabel(t *testing.T) {
	cases := []struct {
		name     string
		key      string
		value    string
		object   map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:  "with existing label",
			key:   "key",
			value: "value",
			object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"existing": "existing",
					},
				},
			},
			expected: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"existing": "existing",
						"key":      "value",
					},
				},
			},
		},
		{
			name:  "without existing label",
			key:   "key",
			value: "value",
			object: map[string]interface{}{
				"metadata": map[string]interface{}{},
			},
			expected: map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"key": "value",
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj := &unstructured.Unstructured{
				Object: tc.object,
			}

			SetMetaDataLabel(obj, tc.key, tc.value)

			require.Equal(t, tc.expected, obj.Object)
		})
	}
}

func TestSetMetadataAnnotation(t *testing.T) {
	cases := []struct {
		name     string
		key      string
		value    string
		object   map[string]interface{}
		expected map[string]interface{}
	}{
		{
			name:  "with existing annotation",
			key:   "key",
			value: "value",
			object: map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"existing": "existing",
					},
				},
			},
			expected: map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"existing": "existing",
						"key":      "value",
					},
				},
			},
		},
		{
			name:  "without existing annotation",
			key:   "key",
			value: "value",
			object: map[string]interface{}{
				"metadata": map[string]interface{}{},
			},
			expected: map[string]interface{}{
				"metadata": map[string]interface{}{
					"annotations": map[string]interface{}{
						"key": "value",
					},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			obj := &unstructured.Unstructured{
				Object: tc.object,
			}

			SetMetaDataAnnotation(obj, tc.key, tc.value)

			require.Equal(t, tc.expected, obj.Object)
		})
	}
}

type fakeResourceInfo struct {
	err      error
	infos    []*resource.Info
	infosErr error
}

func (fri *fakeResourceInfo) Err() error {
	return fri.err
}

func (fri *fakeResourceInfo) Infos() ([]*resource.Info, error) {
	return fri.infos, fri.infosErr
}

func TestRebuildObject(t *testing.T) {
	inputPath := filepath.ToSlash("testdata/deployment.json")
	b, err := ioutil.ReadFile(inputPath)
	require.NoError(t, err)

	m := make(map[string]interface{})
	err = json.Unmarshal(b, &m)
	require.NoError(t, err)

	got, err := RebuildObject(m)
	require.NoError(t, err)

	expected := map[string]interface{}{
		"apiVersion": "extensions/v1beta1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name": "guiroot",
		},
		"spec": map[string]interface{}{
			"replicas": float64(1),
			"template": map[string]interface{}{
				"metadata": map[string]interface{}{
					"labels": map[string]interface{}{
						"app": "guiroot",
					},
				},
				"spec": map[string]interface{}{
					"containers": []interface{}{
						map[string]interface{}{
							"name":  "guiroot",
							"image": "gcr.io/heptio-images/ks-guestbook-demo:0.1",
							"ports": []interface{}{
								map[string]interface{}{
									"containerPort": float64(80),
								},
							},
							"securityContext": map[string]interface{}{
								"capabilities": map[string]interface{}{
									"add": []interface{}{
										"NET_ADMIN", "SYS_TIME",
									},
								},
							},
						},
					},
				},
			},
		},
	}

	require.Equal(t, expected, got)
}

func Test_fetchManagedObjects_Fail(t *testing.T) {
	fakeClients := Clients{}
	_, err := fetchManagedObjects("default", fakeClients, []string{})

	// NOTE: yes this errors.
	require.Error(t, err)
}

type fakeClientConfig struct{}

var _ clientcmd.ClientConfig = (*fakeClientConfig)(nil)

func (fcc *fakeClientConfig) RawConfig() (clientcmdapi.Config, error) {
	return clientcmdapi.Config{}, nil
}

func (fcc *fakeClientConfig) ClientConfig() (*restclient.Config, error) {
	return &restclient.Config{}, nil
}

func (fcc *fakeClientConfig) Namespace() (string, bool, error) {
	return "default", false, nil
}

func (fcc *fakeClientConfig) ConfigAccess() clientcmd.ConfigAccess {
	return nil
}

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
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/dynamic"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/rest/fake"
	"k8s.io/kubernetes/pkg/api/legacyscheme"
	api "k8s.io/kubernetes/pkg/apis/core"
	cmdtesting "k8s.io/kubernetes/pkg/kubectl/cmd/testing"
	"k8s.io/kubernetes/pkg/kubectl/cmd/util/openapi"
	"k8s.io/kubernetes/pkg/kubectl/scheme"
)

var unstructuredSerializer = dynamic.ContentConfig().NegotiatedSerializer

func defaultHeader() http.Header {
	header := http.Header{}
	header.Set("Content-Type", runtime.ContentTypeJSON)
	return header
}

func objBody(codec runtime.Codec, obj runtime.Object) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader([]byte(runtime.EncodeOrDie(codec, obj))))
}

func convertToObject(r io.Reader) (map[string]interface{}, error) {
	var obj map[string]interface{}

	if err := json.NewDecoder(r).Decode(&obj); err != nil {
		return nil, err
	}

	return obj, nil
}

func Test_merger_merge(t *testing.T) {
	tf := cmdtesting.NewTestFactory()
	defer tf.Cleanup()

	codec := legacyscheme.Codecs.LegacyCodec(scheme.Versions...)

	servicePath := "/namespaces/testing/services/service"

	clusterService := &api.Service{
		Spec: api.ServiceSpec{
			Ports: []api.ServicePort{
				{NodePort: 30000},
			},
		},
	}

	tf.UnstructuredClient = &fake.RESTClient{
		NegotiatedSerializer: unstructuredSerializer,
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch p, m := req.URL.Path, req.Method; {
			case p == servicePath && m == "GET":
				return &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(codec, clusterService)}, nil
			case p == servicePath && m == "PATCH":
				defer req.Body.Close()
				_, err := convertToObject(req.Body)
				require.NoError(t, err)

				return &http.Response{StatusCode: 200, Header: defaultHeader(), Body: objBody(codec, clusterService)}, nil
			default:
				t.Fatalf("unexpected request using unstructured client: %#v\n%#v", req.URL, req)
				return nil, nil
			}
		}),
	}

	tf.Client = &fake.RESTClient{
		Client: fake.CreateHTTPClient(func(req *http.Request) (*http.Response, error) {
			switch p, m := req.URL.Path, req.Method; {
			case p == "/openapi/v2" && m == "GET":
				schemaPath := filepath.Join("testdata", "swagger.json")
				f, err := os.Open(schemaPath)
				require.NoError(t, err)

				return &http.Response{StatusCode: 200, Body: f}, nil
			default:
				t.Fatalf("unexpected request using client: %#v\n%#v", req.URL, req)
				return nil, errors.New("not found")
			}
		}),
	}

	tf.OpenAPISchemaFunc = func() (openapi.Resources, error) {
		return nil, errors.New("not found")
	}

	tf.ClientConfigVal = &restclient.Config{}

	om := newDefaultObjectMerger(tf)

	obj := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": "v1",
			"kind":       "Service",
			"metadata": map[string]interface{}{
				"name": "service",
				"labels": map[string]interface{}{
					"foo": "bar",
				},
			},
			"spec": map[string]interface{}{
				"ports": []interface{}{
					map[string]interface{}{
						"protocol":   "TCP",
						"targetPort": 8080,
						"port":       80,
					},
				},
				"selector": map[string]interface{}{
					"app": "MyApp",
				},
				"type": "NodePort",
			},
		},
	}

	_, err := om.Merge("testing", obj)
	require.NoError(t, err)
}

type fakeObjectMerger struct {
	mergeObj *unstructured.Unstructured
	mergeErr error
}

func (om *fakeObjectMerger) Merge(string, *unstructured.Unstructured) (*unstructured.Unstructured, error) {
	return om.mergeObj, om.mergeErr
}

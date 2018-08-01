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

package diff

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/ghodss/yaml"
	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/client"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/clientcmd"
)

type fakeYamlGenerator struct {
	b   []byte
	err error
}

func (fyg *fakeYamlGenerator) Generate(l *Location, components []string) (io.ReadSeeker, error) {
	var r io.ReadSeeker
	if fyg.b != nil {
		r = bytes.NewReader(fyg.b)
	} else {
		r = bytes.NewReader([]byte{})
	}

	return r, fyg.err
}

func TestDiffer(t *testing.T) {
	test.WithApp(t, "/", func(appMock *mocks.App, fs afero.Fs) {
		differ := New(appMock, &client.Config{}, []string{})

		localGen := &fakeYamlGenerator{}
		differ.localGen = localGen

		remoteGen := &fakeYamlGenerator{}
		differ.remoteGen = remoteGen

		l1 := NewLocation("default")
		l2 := NewLocation("default")

		r, err := differ.Diff(l1, l2)
		require.NoError(t, err)

		expected := ""

		b, err := ioutil.ReadAll(r)
		require.NoError(t, err)

		require.Equal(t, expected, string(b))
	})
}

func Test_yamlLocal(t *testing.T) {
	cases := []struct {
		name             string
		collectObjectsFn func(a app.App, envName string, componentNames []string) ([]*unstructured.Unstructured, error)
		showFn           func(io.Writer, []*unstructured.Unstructured) error
		expected         string
		isErr            bool
	}{
		{
			name: "in general",
			collectObjectsFn: func(a app.App, envName string, componentNames []string) ([]*unstructured.Unstructured, error) {
				return nil, nil
			},
			showFn: func(w io.Writer, objects []*unstructured.Unstructured) error {
				fmt.Fprint(w, "output")
				return nil
			},
			expected: "output",
		},
		{
			name: "show failed",
			collectObjectsFn: func(a app.App, envName string, componentNames []string) ([]*unstructured.Unstructured, error) {
				return nil, nil
			},
			showFn: func(w io.Writer, objects []*unstructured.Unstructured) error {
				return errors.New("fail")
			},
			isErr: true,
		},
		{
			name: "sorted",
			collectObjectsFn: func(a app.App, envName string, componentNames []string) ([]*unstructured.Unstructured, error) {
				return genObjects(), nil
			},
			showFn:   showYAML,
			expected: sortedYAML,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			test.WithApp(t, "/", func(appMock *mocks.App, fs afero.Fs) {
				location := NewLocation("default")

				yl := newYamlLocal(appMock)

				yl.collectObjectsFn = tc.collectObjectsFn
				yl.showFn = tc.showFn

				rs, err := yl.Generate(location, []string{})
				if tc.isErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				b, err := ioutil.ReadAll(rs)
				require.NoError(t, err)

				require.Equal(t, tc.expected, string(b))
			})
		})
	}

}

func Test_yamlRemote(t *testing.T) {
	validAppSetup := func(a *mocks.App) {
		myEnv := &app.EnvironmentConfig{
			Destination: &app.EnvironmentDestinationSpec{
				Namespace: "default",
			},
		}
		a.On("Environment", "default").Return(myEnv, nil)
	}

	cases := []struct {
		name      string
		appSetup  func(a *mocks.App)
		collectFn func(namespace string, config clientcmd.ClientConfig, components []string) ([]*unstructured.Unstructured, error)
		showFn    func(w io.Writer, objects []*unstructured.Unstructured) error
		expected  string
		isErr     bool
	}{
		{
			name:     "in general",
			appSetup: validAppSetup,
			collectFn: func(namespace string, config clientcmd.ClientConfig, components []string) ([]*unstructured.Unstructured, error) {
				return nil, nil
			},
			showFn: func(w io.Writer, objects []*unstructured.Unstructured) error {
				fmt.Fprintf(w, "output")
				return nil
			},
			expected: "output",
		},
		{
			name: "invalid environment",
			appSetup: func(a *mocks.App) {
				a.On("Environment", "default").Return(nil, errors.New("fail"))
			},
			isErr: true,
		},
		{
			name:     "collect objects failed",
			appSetup: validAppSetup,
			collectFn: func(namespace string, config clientcmd.ClientConfig, components []string) ([]*unstructured.Unstructured, error) {
				return nil, errors.New("fail")
			},
			isErr: true,
		},
		{
			name:     "show failed",
			appSetup: validAppSetup,
			collectFn: func(namespace string, config clientcmd.ClientConfig, components []string) ([]*unstructured.Unstructured, error) {
				return nil, nil
			},
			showFn: func(w io.Writer, objects []*unstructured.Unstructured) error {
				return errors.New("fail")
			},
			isErr: true,
		},
		{
			name:     "sorted",
			appSetup: validAppSetup,
			collectFn: func(namespace string, config clientcmd.ClientConfig, components []string) ([]*unstructured.Unstructured, error) {
				return genObjects(), nil
			},
			showFn:   showYAML,
			expected: sortedYAML,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			test.WithApp(t, "/", func(appMock *mocks.App, fs afero.Fs) {
				tc.appSetup(appMock)

				config := &client.Config{}
				yr := newYamlRemote(appMock, config)

				yr.collectObjectsFn = tc.collectFn
				yr.showFn = tc.showFn

				location := NewLocation("default")

				rs, err := yr.Generate(location, []string{})
				if tc.isErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				b, err := ioutil.ReadAll(rs)
				require.NoError(t, err)

				require.Equal(t, tc.expected, string(b))
			})
		})
	}
}

func showYAML(out io.Writer, objects []*unstructured.Unstructured) error {
	for _, obj := range objects {
		fmt.Fprintln(out, "---")
		buf, err := yaml.Marshal(obj)
		if err != nil {
			return err
		}
		_, err = out.Write(buf)
		if err != nil {
			return err
		}
	}

	return nil
}

func genObjects() []*unstructured.Unstructured {
	return []*unstructured.Unstructured{
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      "deploymentZ",
					"namespace": "default",
					"annotations": map[string]interface{}{
						"app": "Z",
					},
				},
			},
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Service",
				"metadata": map[string]interface{}{
					"name":      "serviceA",
					"namespace": "default",
				},
			},
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "apps/v1",
				"kind":       "Deployment",
				"metadata": map[string]interface{}{
					"name":      "deploymentA",
					"namespace": "default",
					"annotations": map[string]interface{}{
						"app": "Z",
					},
				},
			},
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"namespace":    "default",
					"generateName": "podZ",
				},
			},
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"namespace":    "a-before-d",
					"generateName": "podA",
				},
			},
		},
		&unstructured.Unstructured{
			Object: map[string]interface{}{
				"apiVersion": "v1",
				"kind":       "Pod",
				"metadata": map[string]interface{}{
					"namespace":    "default",
					"generateName": "podA",
				},
			},
		},
	}
}

var sortedYAML = `---
apiVersion: v1
kind: Pod
metadata:
  generateName: podA
  namespace: a-before-d
---
apiVersion: v1
kind: Pod
metadata:
  generateName: podA
  namespace: default
---
apiVersion: v1
kind: Pod
metadata:
  generateName: podZ
  namespace: default
---
apiVersion: v1
kind: Service
metadata:
  name: serviceA
  namespace: default
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    app: Z
  name: deploymentA
  namespace: default
---
apiVersion: apps/v1
kind: Deployment
metadata:
  annotations:
    app: Z
  name: deploymentZ
  namespace: default
`

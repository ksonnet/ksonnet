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

	"github.com/ksonnet/ksonnet/pkg/app"
	"github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/client"
	"github.com/ksonnet/ksonnet/pkg/cluster"
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
		name   string
		showFn func(c cluster.ShowConfig, opts ...cluster.ShowOpts) error
		isErr  bool
	}{
		{
			name: "in general",
			showFn: func(c cluster.ShowConfig, opts ...cluster.ShowOpts) error {
				fmt.Fprint(c.Out, "output")
				return nil
			},
		},
		{
			name: "show failed",
			showFn: func(c cluster.ShowConfig, opts ...cluster.ShowOpts) error {
				return errors.New("fail")
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			test.WithApp(t, "/", func(appMock *mocks.App, fs afero.Fs) {
				location := NewLocation("default")

				yl := newYamlLocal(appMock)

				yl.showFn = tc.showFn

				rs, err := yl.Generate(location, []string{})
				if tc.isErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				b, err := ioutil.ReadAll(rs)
				require.NoError(t, err)

				require.Equal(t, "output", string(b))
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

				require.Equal(t, "output", string(b))
			})
		})
	}
}

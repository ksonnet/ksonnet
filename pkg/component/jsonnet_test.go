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

package component

import (
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/util/test"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestJsonnet_Type(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {
		test.StageFile(t, fs, filepath.FromSlash("guestbook/guestbook-ui.jsonnet"), filepath.FromSlash("/app/components/guestbook-ui.jsonnet"))
		test.StageFile(t, fs, filepath.FromSlash("guestbook/params.libsonnet"), filepath.FromSlash("/app/components/params.libsonnet"))

		j := NewJsonnet(a, "", "/app/components/guestbook-ui.jsonnet", "/app/components/params.libsonnet")
		require.Equal(t, TypeJsonnet, j.Type())
	})
}

func TestJsonnet_Name(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {

		files := []string{"guestbook-ui.jsonnet", "params.libsonnet"}
		for _, file := range files {
			test.StageFile(t, fs, "guestbook/"+file, "/app/components/"+file)
			test.StageFile(t, fs, "guestbook/"+file, "/app/components/nested/"+file)
		}

		root := NewJsonnet(a, "", "/app/components/guestbook-ui.jsonnet", "/app/components/params.libsonnet")
		nested := NewJsonnet(a, "nested", "/app/components/nested/guestbook-ui.jsonnet", "/app/components/nested/params.libsonnet")

		cases := []struct {
			name         string
			isNameSpaced bool
			expected     string
			c            *Jsonnet
		}{
			{
				name:         "wants namespaced",
				isNameSpaced: true,
				expected:     "guestbook-ui",
				c:            root,
			},
			{
				name:         "no namespace",
				isNameSpaced: false,
				expected:     "guestbook-ui",
				c:            root,
			},
			{
				name:         "nested: wants namespaced",
				isNameSpaced: true,
				expected:     "nested.guestbook-ui",
				c:            nested,
			},
			{
				name:         "nested: no namespace",
				isNameSpaced: false,
				expected:     "guestbook-ui",
				c:            nested,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				require.Equal(t, tc.expected, tc.c.Name(tc.isNameSpaced))
			})
		}
	})

}

func TestJsonnet_Params(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {

		files := []string{"guestbook-ui.jsonnet", "k.libsonnet", "k8s.libsonnet", "params.libsonnet"}
		for _, file := range files {
			test.StageFile(t, fs, "guestbook/"+file, "/app/components/"+file)
		}

		c := NewJsonnet(a, "", "/app/components/guestbook-ui.jsonnet", "/app/components/params.libsonnet")

		params, err := c.Params("")
		require.NoError(t, err)

		expected := []ModuleParameter{
			{
				Component: "guestbook-ui",
				Key:       "containerPort",
				Value:     "80",
			},
			{
				Component: "guestbook-ui",
				Key:       "image",
				Value:     `"gcr.io/heptio-images/ks-guestbook-demo:0.1"`,
			},
			{
				Component: "guestbook-ui",
				Key:       "name",
				Value:     `"guiroot"`,
			},
			{
				Component: "guestbook-ui",
				Key:       "obj",
				Value:     `{"a":"b"}`,
			},
			{
				Component: "guestbook-ui",
				Key:       "replicas",
				Value:     "1",
			},
			{
				Component: "guestbook-ui",
				Key:       "servicePort",
				Value:     "80",
			},
			{
				Component: "guestbook-ui",
				Key:       "type",
				Value:     `"ClusterIP"`,
			},
		}

		require.Equal(t, expected, params)
	})
}

func TestJsonnet_Summarize(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {

		files := []string{"guestbook-ui.jsonnet", "k.libsonnet", "k8s.libsonnet", "params.libsonnet"}
		for _, file := range files {
			test.StageFile(t, fs, "guestbook/"+file, "/app/components/"+file)
		}

		c := NewJsonnet(a, "", "/app/components/guestbook-ui.jsonnet", "/app/components/params.libsonnet")

		got, err := c.Summarize()
		require.NoError(t, err)

		expected := Summary{ComponentName: "guestbook-ui", Type: "jsonnet"}

		require.Equal(t, expected, got)
	})
}

func TestJsonnet_SetParam(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {

		files := []string{"guestbook-ui.jsonnet", "k.libsonnet", "k8s.libsonnet", "params.libsonnet"}
		for _, file := range files {
			test.StageFile(t, fs, "guestbook/"+file, "/app/components/"+file)
		}

		c := NewJsonnet(a, "", "/app/components/guestbook-ui.jsonnet", "/app/components/params.libsonnet")

		err := c.SetParam([]string{"replicas"}, 4)
		require.NoError(t, err)

		b, err := afero.ReadFile(fs, "/app/components/params.libsonnet")
		require.NoError(t, err)

		expected := testdata(t, "guestbook/set-params.libsonnet")

		require.Equal(t, string(expected), string(b))
	})
}

func TestJsonnet_DeleteParam(t *testing.T) {
	test.WithApp(t, "/app", func(a *mocks.App, fs afero.Fs) {

		files := []string{"guestbook-ui.jsonnet", "k.libsonnet", "k8s.libsonnet", "params.libsonnet"}
		for _, file := range files {
			test.StageFile(t, fs, "guestbook/"+file, "/app/components/"+file)
		}

		c := NewJsonnet(a, "", "/app/components/guestbook-ui.jsonnet", "/app/components/params.libsonnet")

		err := c.DeleteParam([]string{"replicas"})
		require.NoError(t, err)

		b, err := afero.ReadFile(fs, "/app/components/params.libsonnet")
		require.NoError(t, err)

		expected := testdata(t, "guestbook/delete-params.libsonnet")

		require.Equal(t, string(expected), string(b))
	})
}

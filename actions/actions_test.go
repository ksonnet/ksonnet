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

package actions

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ksonnet/ksonnet/client"
	"github.com/ksonnet/ksonnet/pkg/app/mocks"
	cmocks "github.com/ksonnet/ksonnet/pkg/component/mocks"
	"github.com/ksonnet/ksonnet/pkg/registry"
	rmocks "github.com/ksonnet/ksonnet/pkg/registry/mocks"
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func Test_optionLoader_types(t *testing.T) {
	cases := []struct {
		name    string
		hasArg  bool
		valid   interface{}
		invalid interface{}
		keyName string
	}{
		{
			name:    "App",
			valid:   &mocks.App{},
			invalid: "invalid",
			keyName: OptionApp,
		},
		{
			name:    "Bool",
			hasArg:  true,
			valid:   true,
			invalid: "invalid",
			keyName: OptionSkipGc,
		},
		{
			name:    "Fs",
			hasArg:  true,
			valid:   afero.NewMemMapFs(),
			invalid: "invalid",
			keyName: OptionFs,
		},
		{
			name:    "Int",
			hasArg:  true,
			valid:   0,
			invalid: "invalid",
			keyName: OptionName,
		},
		{
			name:    "Int64",
			hasArg:  true,
			valid:   int64(0),
			invalid: "invalid",
			keyName: OptionName,
		},
		{
			name:    "String",
			hasArg:  true,
			valid:   "valid",
			invalid: 0,
			keyName: OptionName,
		},
		{
			name:    "StringSlice",
			hasArg:  true,
			valid:   []string{},
			invalid: "invalid",
			keyName: OptionName,
		},
		{
			name:    "ClientConfig",
			valid:   &client.Config{},
			invalid: "invalid",
			keyName: OptionClientConfig,
		},
	}

	for _, tc := range cases {
		methodName := fmt.Sprintf("Load%s", tc.name)

		t.Run(tc.name+" valid", func(t *testing.T) {
			m := map[string]interface{}{
				tc.keyName: tc.valid,
			}

			ol := newOptionLoader(m)

			loader := reflect.ValueOf(ol).MethodByName(methodName)

			callValues := make([]reflect.Value, 0)
			if tc.hasArg {
				callValues = append(callValues, reflect.ValueOf(tc.keyName))
			}

			values := loader.Call(callValues)
			require.Len(t, values, 1)
			require.EqualValues(t, tc.valid, values[0].Interface())
		})

		t.Run(tc.name+" invalid", func(t *testing.T) {
			m := map[string]interface{}{
				tc.keyName: tc.invalid,
			}

			ol := newOptionLoader(m)

			loader := reflect.ValueOf(ol).MethodByName(methodName)

			callValues := make([]reflect.Value, 0)
			if tc.hasArg {
				callValues = append(callValues, reflect.ValueOf(tc.keyName))
			}

			loader.Call(callValues)
			require.Error(t, ol.err)
		})

		t.Run(tc.name+" previous error", func(t *testing.T) {
			m := map[string]interface{}{
				tc.keyName: tc.invalid,
			}

			ol := newOptionLoader(m)
			ol.err = errors.New("error")

			loader := reflect.ValueOf(ol).MethodByName(methodName)

			callValues := make([]reflect.Value, 0)
			if tc.hasArg {
				callValues = append(callValues, reflect.ValueOf(tc.keyName))
			}

			loader.Call(callValues)
			require.Error(t, ol.err)
		})
	}
}

func Test_optionLoader_optional_types(t *testing.T) {
	cases := []struct {
		name     string
		valid    interface{}
		invalid  interface{}
		expected interface{}
		keyName  string
	}{
		{
			name:     "Bool",
			valid:    true,
			invalid:  "invalid",
			expected: false,
			keyName:  OptionApp,
		},
		{
			name:     "Int",
			valid:    9,
			invalid:  "invalid",
			expected: 0,
			keyName:  OptionApp,
		},
		{
			name:     "String",
			valid:    "valid",
			invalid:  9,
			expected: "",
			keyName:  OptionApp,
		},
	}

	for _, tc := range cases {
		methodName := fmt.Sprintf("LoadOptional%s", tc.name)

		t.Run(tc.name+" valid", func(t *testing.T) {
			m := map[string]interface{}{
				tc.keyName: tc.valid,
			}

			ol := newOptionLoader(m)

			loader := reflect.ValueOf(ol).MethodByName(methodName)

			callValues := []reflect.Value{reflect.ValueOf(tc.keyName)}

			values := loader.Call(callValues)
			require.Len(t, values, 1)
			require.EqualValues(t, tc.valid, values[0].Interface())
		})

		t.Run(tc.name+" invalid", func(t *testing.T) {
			m := map[string]interface{}{
				tc.keyName: tc.invalid,
			}

			ol := newOptionLoader(m)

			loader := reflect.ValueOf(ol).MethodByName(methodName)

			callValues := []reflect.Value{reflect.ValueOf(tc.keyName)}
			values := loader.Call(callValues)
			require.Len(t, values, 1)
			require.EqualValues(t, tc.expected, values[0].Interface())
		})

		t.Run(tc.name+" previous error", func(t *testing.T) {
			m := map[string]interface{}{
				tc.keyName: tc.invalid,
			}

			ol := newOptionLoader(m)
			ol.err = errors.New("error")

			loader := reflect.ValueOf(ol).MethodByName(methodName)

			callValues := []reflect.Value{reflect.ValueOf(tc.keyName)}
			loader.Call(callValues)
			require.Error(t, ol.err)
		})
	}
}

func withApp(t *testing.T, fn func(*mocks.App)) {
	fs := afero.NewMemMapFs()

	appMock := &mocks.App{}
	appMock.On("Fs").Return(fs)
	appMock.On("Root").Return("/")

	fn(appMock)
}

func assertOutput(t *testing.T, filename, actual string) {
	require.NotEmpty(t, filename, "filename can not be empty")
	path := filepath.Join("testdata", filename)
	b, err := ioutil.ReadFile(path)
	require.NoError(t, err)

	require.Equal(t, string(b), actual)
}

func stageFile(t *testing.T, fs afero.Fs, src, dest string) {
	in := filepath.Join("testdata", src)

	b, err := ioutil.ReadFile(in)
	require.NoError(t, err)

	dir := filepath.Dir(dest)
	err = fs.MkdirAll(dir, 0755)
	require.NoError(t, err)

	err = afero.WriteFile(fs, dest, b, 0644)
	require.NoError(t, err)
}

func mockNsWithName(name string) *cmocks.Module {
	m := &cmocks.Module{}
	m.On("Name").Return(name)
	return m
}

func mockRegistry(name string, isOverride bool) *rmocks.Registry {
	m := &rmocks.Registry{}
	m.On("Name").Return(name)
	m.On("Protocol").Return(registry.ProtocolGitHub)
	m.On("URI").Return("github.com/ksonnet/parts/tree/master/incubator")
	m.On("IsOverride").Return(isOverride)

	return m
}

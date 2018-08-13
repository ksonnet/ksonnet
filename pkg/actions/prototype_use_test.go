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
	"testing"

	param "github.com/ksonnet/ksonnet/metadata/params"
	"github.com/ksonnet/ksonnet/pkg/app"
	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/prototype"
	registrymocks "github.com/ksonnet/ksonnet/pkg/registry/mocks"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPrototypeUse(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		prototypes := prototype.Prototypes{}

		manager := &registrymocks.PackageManager{}
		manager.On("Prototypes").Return(prototypes, nil)

		args := []string{
			"single-port-deployment",
			"deployment",
			"--name", "deployment",
			"--image", "nginx",
			"--containerPort", "80",
		}

		in := map[string]interface{}{
			OptionApp:           appMock,
			OptionArguments:     args,
			OptionTLSSkipVerify: false,
		}

		a, err := NewPrototypeUse(in)
		require.NoError(t, err)

		a.packageManager = manager

		a.createComponentFn = func(_ app.App, moduleName, name, text string, params param.Params, template prototype.TemplateType) (string, error) {
			assert.Equal(t, "", moduleName)
			assert.Equal(t, "deployment", name)
			assertOutput(t, "prototype/use/text.txt", text)

			expectedParams := param.Params{
				"name":          `"deployment"`,
				"image":         `"nginx"`,
				"replicas":      "1",
				"containerPort": "80",
			}

			assert.Equal(t, expectedParams, params)
			assert.Equal(t, prototype.Jsonnet, template)

			return "", nil
		}

		err = a.Run()
		require.NoError(t, err)
	})
}

func TestPrototypeUse_bind_flags_failed(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		prototypes := prototype.Prototypes{}

		manager := &registrymocks.PackageManager{}
		manager.On("Prototypes").Return(prototypes, nil)

		args := []string{
			"single-port-deployment",
			"deployment",
			"--name", "deployment",
			"--image", "nginx",
		}

		in := map[string]interface{}{
			OptionApp:           appMock,
			OptionArguments:     args,
			OptionTLSSkipVerify: false,
		}

		a, err := NewPrototypeUse(in)
		require.NoError(t, err)

		a.createComponentFn = func(_ app.App, moduleName, name string, text string, params param.Params, template prototype.TemplateType) (string, error) {
			assert.Equal(t, "", moduleName)
			assert.Equal(t, "deployment", name)
			assertOutput(t, "prototype/use/text.txt", text)

			expectedParams := param.Params{
				"name":          `"deployment"`,
				"image":         `"nginx"`,
				"replicas":      "1",
				"containerPort": "80",
			}

			assert.Equal(t, expectedParams, params)
			assert.Equal(t, prototype.Jsonnet, template)

			return "", nil
		}

		a.bindFlagsFn = func(*prototype.Prototype) (*pflag.FlagSet, error) {
			return nil, errors.New("failed")
		}

		a.packageManager = manager

		err = a.Run()
		require.Error(t, err)
	})
}

func TestPrototypeUse_with_module_in_name(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		prototypes := prototype.Prototypes{}

		manager := &registrymocks.PackageManager{}
		manager.On("Prototypes").Return(prototypes, nil)

		args := []string{
			"single-port-deployment",
			"module.deployment",
			"--image", "nginx",
			"--containerPort", "80",
		}

		in := map[string]interface{}{
			OptionApp:           appMock,
			OptionArguments:     args,
			OptionTLSSkipVerify: false,
		}

		a, err := NewPrototypeUse(in)
		require.NoError(t, err)

		a.createComponentFn = func(_ app.App, moduleName, name string, text string, params param.Params, template prototype.TemplateType) (string, error) {
			assert.Equal(t, "module", moduleName)
			assert.Equal(t, "deployment", name)
			assertOutput(t, "prototype/use/text.txt", text)

			expectedParams := param.Params{
				"name":          `"deployment"`,
				"image":         `"nginx"`,
				"replicas":      "1",
				"containerPort": "80",
			}

			assert.Equal(t, expectedParams, params)
			assert.Equal(t, prototype.Jsonnet, template)

			return "", nil
		}

		a.packageManager = manager

		err = a.Run()
		require.NoError(t, err)
	})
}

func TestPrototypeUse_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := NewPrototypeUse(in)
	require.Error(t, err)
}

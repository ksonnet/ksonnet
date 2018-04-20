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
	"bytes"
	"path/filepath"
	"testing"

	"github.com/ksonnet/ksonnet/metadata/app"
	amocks "github.com/ksonnet/ksonnet/metadata/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/component"
	"github.com/ksonnet/ksonnet/pkg/component/mocks"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestParamDiff(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		env1 := "env1"
		env2 := "env2"

		moduleEnv1 := &mocks.Module{}
		env1Params := []component.ModuleParameter{
			{Component: "a", Key: "a", Value: "a"},
			{Component: "a", Key: "b", Value: "b1"},
			{Component: "c", Key: "c", Value: "c"},
		}
		moduleEnv1.On("Params", "env1").Return(env1Params, nil)

		moduleEnv2 := &mocks.Module{}
		env2Params := []component.ModuleParameter{
			{Component: "a", Key: "a", Value: "a"},
			{Component: "a", Key: "b", Value: "b2"},
			{Component: "d", Key: "d", Value: "d"},
		}
		moduleEnv2.On("Params", "env2").Return(env2Params, nil)

		in := map[string]interface{}{
			OptionApp:      appMock,
			OptionEnvName1: env1,
			OptionEnvName2: env2,
		}

		a, err := NewParamDiff(in)
		require.NoError(t, err)

		a.modulesFromEnvFn = func(_ app.App, envName string) ([]component.Module, error) {
			switch envName {
			case env1:
				return []component.Module{moduleEnv1}, nil
			case env2:
				return []component.Module{moduleEnv2}, nil
			default:
				return nil, errors.Errorf("unknown env %s", envName)
			}
		}

		var buf bytes.Buffer
		a.out = &buf

		err = a.Run()
		require.NoError(t, err)

		assertOutput(t, filepath.Join("param", "diff", "output.txt"), buf.String())
	})
}

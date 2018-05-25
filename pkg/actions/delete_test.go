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

	amocks "github.com/ksonnet/ksonnet/pkg/app/mocks"
	"github.com/ksonnet/ksonnet/pkg/client"
	"github.com/ksonnet/ksonnet/pkg/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	cases := []struct {
		name        string
		isSetupErr  bool
		currentName string
		envName     string
	}{
		{
			name:    "with a supplied env",
			envName: "default",
		},
		{
			name:        "with a current env",
			currentName: "default",
		},
		{
			name:       "without supplied or current env",
			isSetupErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			withApp(t, func(appMock *amocks.App) {
				appMock.On("CurrentEnvironment").Return(tc.currentName)

				in := map[string]interface{}{
					OptionApp:            appMock,
					OptionClientConfig:   &client.Config{},
					OptionComponentNames: []string{},
					OptionEnvName:        tc.envName,
					OptionGracePeriod:    int64(3),
				}

				expected := cluster.DeleteConfig{
					App:            appMock,
					ClientConfig:   &client.Config{},
					ComponentNames: []string{},
					EnvName:        "default",
					GracePeriod:    3,
				}

				runDeleteOpt := func(a *Delete) {
					a.runDeleteFn = func(config cluster.DeleteConfig, opts ...cluster.DeleteOpts) error {
						assert.Equal(t, expected, config)
						return nil
					}
				}

				a, err := newDelete(in, runDeleteOpt)
				if tc.isSetupErr {
					require.Error(t, err)
					return
				}
				require.NoError(t, err)

				err = a.run()
				require.NoError(t, err)
			})
		})
	}
}

func TestDelete_invalid_input(t *testing.T) {
	withApp(t, func(appMock *amocks.App) {
		in := map[string]interface{}{
			OptionClientConfig: "invalid",
		}

		_, err := newDelete(in)
		require.Error(t, err)
	})
}

func TestDelete_requires_app(t *testing.T) {
	in := make(map[string]interface{})
	_, err := newDelete(in)
	require.Error(t, err)
}

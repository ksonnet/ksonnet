package env

import (
	"testing"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/metadata/app/mocks"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestList(t *testing.T) {
	withEnv(t, func(fs afero.Fs) {
		appMock := &mocks.App{}

		specEnvs := app.EnvironmentSpecs{
			"default": &app.EnvironmentSpec{
				Path: "default",
				Destination: &app.EnvironmentDestinationSpec{
					Namespace: "default",
					Server:    "http://example.com",
				},
				KubernetesVersion: "v1.8.7",
			},
		}
		appMock.On("Environments").Return(specEnvs, nil)

		envs, err := List(appMock)
		require.NoError(t, err)

		expected := map[string]Env{
			"default": Env{
				KubernetesVersion: "v1.8.7",
				Name:              "default",
				Destination:       NewDestination("http://example.com", "default"),
			},
		}

		require.Equal(t, expected, envs)
	})
}

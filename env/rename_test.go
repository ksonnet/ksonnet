package env

import (
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/ksonnet/ksonnet/metadata/app"
	"github.com/ksonnet/ksonnet/metadata/app/mocks"
)

func TestRename(t *testing.T) {
	withEnv(t, func(fs afero.Fs) {
		appMock := &mocks.App{}

		envSpec := &app.EnvironmentSpec{
			Path:              "env1",
			Destination:       &app.EnvironmentDestinationSpec{Namespace: "default", Server: "http://example.com"},
			KubernetesVersion: "v1.9.2",
		}
		appMock.On("Environment", "env1").Return(envSpec, nil)

		appMock.On(
			"AddEnvironment",
			"env1",
			"version:v1.9.2",
			mock.AnythingOfType("*app.EnvironmentSpec")).Return(nil)

		config := RenameConfig{
			App:     appMock,
			AppRoot: "/",
			Fs:      fs,
		}

		checkExists(t, fs, "/environments/env1")

		err := Rename("env1", "env1-updated", config)
		require.NoError(t, err)

		checkNotExists(t, fs, "/environments/env1")
		checkExists(t, fs, "/environments/env1-updated/main.jsonnet")
	})
}

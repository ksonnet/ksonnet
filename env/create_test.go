package env

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/mock"

	"github.com/ksonnet/ksonnet/metadata/app/mocks"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestCreate(t *testing.T) {
	withEnv(t, func(fs afero.Fs) {
		appMock := &mocks.App{}
		appMock.On("Environment", "newenv").Return(nil, errors.New("it does not exist"))
		appMock.On(
			"AddEnvironment",
			"newenv",
			"version:v1.8.7",
			mock.AnythingOfType("*app.EnvironmentSpec"),
		).Return(nil)

		config := CreateConfig{
			App:         appMock,
			Fs:          fs,
			Destination: NewDestination("http://example.com", "default"),
			RootPath:    "/",
			Name:        "newenv",
			K8sSpecFlag: "version:v1.8.7",
		}

		err := Create(config)
		require.NoError(t, err)

		checkExists(t, fs, "/environments/newenv/main.jsonnet")
		checkExists(t, fs, "/environments/newenv/params.libsonnet")
	})
}

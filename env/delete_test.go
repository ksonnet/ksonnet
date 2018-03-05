package env

import (
	"testing"

	"github.com/ksonnet/ksonnet/metadata/app/mocks"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func TestDelete(t *testing.T) {
	withEnv(t, func(fs afero.Fs) {
		appMock := &mocks.App{}
		appMock.On("RemoveEnvironment", "nested/env3").Return(nil)

		config := DeleteConfig{
			App:     appMock,
			Fs:      fs,
			Name:    "nested/env3",
			AppRoot: "/",
		}

		err := Delete(config)
		require.NoError(t, err)

		checkNotExists(t, fs, "/environments/nested")
	})
}

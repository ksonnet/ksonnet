package env

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"

	"github.com/ksonnet/ksonnet/metadata/app/mocks"
)

func TestRename(t *testing.T) {
	withEnv(t, func(fs afero.Fs) {
		appMock := &mocks.App{}

		appMock.On("RenameEnvironment", "env1", "env1-updated").Return(nil)

		config := RenameConfig{
			App:     appMock,
			AppRoot: "/",
			Fs:      fs,
		}

		err := Rename("env1", "env1-updated", config)
		require.NoError(t, err)
	})
}

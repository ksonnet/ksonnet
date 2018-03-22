package app

import (
	"testing"

	"github.com/spf13/afero"
	"github.com/stretchr/testify/require"
)

func Test_findRoot(t *testing.T) {
	fs := afero.NewMemMapFs()
	stageFile(t, fs, "app010_app.yaml", "/app/app.yaml")

	dirs := []string{
		"/app/components",
		"/invalid",
	}

	for _, dir := range dirs {
		err := fs.MkdirAll(dir, DefaultFilePermissions)
		require.NoError(t, err)
	}

	cases := []struct {
		name     string
		expected string
		isErr    bool
	}{
		{
			name:     "/app",
			expected: "/app",
		},
		{
			name:     "/app/components",
			expected: "/app",
		},
		{
			name:  "/invalid",
			isErr: true,
		},
		{
			name:  "/missing",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			root, err := findRoot(fs, tc.name)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.Equal(t, tc.expected, root)
		})
	}

}

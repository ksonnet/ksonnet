package component

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_locateTarget(t *testing.T) {
	cases := []struct {
		name     string
		in       string
		expected string
	}{
		{
			name:     "root path",
			in:       ".",
			expected: filepath.ToSlash("/components"),
		},
		{
			name:     "nested path 1",
			in:       "foo",
			expected: filepath.ToSlash("/components/foo"),
		},
		{
			name:     "nested path 2",
			in:       "foo.bar",
			expected: filepath.ToSlash("/components/foo/bar"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := locateTarget(filepath.FromSlash("/"), tc.in)
			require.Equal(t, tc.expected, got)
		})
	}
}

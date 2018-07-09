package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMake(t *testing.T) {
	cases := []struct {
		name  string
		in    string
		out   string
		isErr bool
	}{
		{
			name: "semantic version",
			in:   "1.0.0",
			out:  "1.0.0",
		},
		{
			name: "major.minor",
			in:   "1.0",
			out:  "1.0",
		},
		{
			name: "major",
			in:   "1",
			out:  "1",
		},
		{
			name: "version has v prefix",
			in:   "v1.2.3",
			out:  "v1.2.3",
		},
		{
			name: "semantic version with additional info",
			in:   "1.2.3-beta.1",
			out:  "1.2.3-beta.1",
		},
		{
			name:  "empty version string",
			in:    "",
			isErr: true,
		},
		{
			name:  "invalid version",
			in:    "invalid",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			v, err := Make(tc.in)

			if tc.isErr {
				require.Error(t, err)
				return
			}

			assert.Equal(t, tc.out, v.String())
		})
	}
}

package prototype

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTemplateType(t *testing.T) {
	cases := []struct {
		name     string
		expected TemplateType
		isErr    bool
	}{
		{
			name:     "yaml",
			expected: YAML,
		},
		{
			name:     "yml",
			expected: YAML,
		},
		{
			name:     "json",
			expected: JSON,
		},
		{
			name:     "jsonnet",
			expected: Jsonnet,
		},
		{
			name:  "unknown",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tt, err := ParseTemplateType(tc.name)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.Equal(t, tc.expected, tt)
		})
	}
}

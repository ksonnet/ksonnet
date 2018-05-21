package prototype

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_parseParamType(t *testing.T) {
	cases := []struct {
		name     string
		expected ParamType
		isErr    bool
	}{
		{
			name:     "number",
			expected: Number,
		},
		{
			name:     "string",
			expected: String,
		},
		{
			name:     "numberOrString",
			expected: NumberOrString,
		},
		{
			name:     "object",
			expected: Object,
		},
		{
			name:     "array",
			expected: Array,
		},
		{
			name:  "invalid",
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseParamType(tc.name)
			if tc.isErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tc.expected, got)
		})
	}
}

func TestParseParam_String(t *testing.T) {
	cases := []struct {
		name     string
		in       ParamType
		expected string
	}{
		{
			name: "number",
			in:   Number,
		},
		{
			name: "string",
			in:   String,
		},
		{
			name: "numberOrString",
			in:   NumberOrString,
		},
		{
			name: "object",
			in:   Object,
		},
		{
			name: "array",
			in:   Array,
		},
		{
			name: "unknown",
			in:   ParamType("unknown"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s := tc.in.String()
			require.Equal(t, tc.name, s)
		})
	}
}

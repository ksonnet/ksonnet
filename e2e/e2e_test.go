package e2e

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_initOptions_toSlice(t *testing.T) {
	cases := []struct {
		name     string
		io       initOptions
		expected []string
		isErr    bool
	}{
		{
			name: "namespace",
			io: initOptions{
				namespace: "ns",
			},
			expected: []string{
				"--server", "http://example.com",
				"--namespace", "ns",
			},
		},
		{
			name: "no server or context",
			io:   initOptions{},
			expected: []string{
				"--server", "http://example.com",
			},
		},
		{
			name: "server",
			io: initOptions{
				server: "http://example.com/2",
			},
			expected: []string{
				"--server", "http://example.com/2",
			},
		},
		{
			name: "context",
			io: initOptions{
				context: "minikube",
			},
			expected: []string{
				"--context", "minikube",
			},
		},
		{
			name: "server and context",
			io: initOptions{
				context: "minikube",
				server:  "http://example.com",
			},
			isErr: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sl, err := tc.io.toSlice()
			if tc.isErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			assert.Equal(t, tc.expected, sl)
		})
	}
}

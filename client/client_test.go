package client

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfig_GetAPISpec(t *testing.T) {
	b, err := ioutil.ReadFile("testdata/swagger.json")
	require.NoError(t, err)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, string(b))
	}))
	defer ts.Close()

	cases := []struct {
		name      string
		serverURL string
		expected  string
	}{
		{
			name:      "invalid server URL",
			serverURL: "http://+++",
			expected:  defaultVersion,
		},
		{
			name:      "with a server",
			serverURL: ts.URL,
			expected:  "version:v1.9.3",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c := Config{}
			got := c.GetAPISpec(tc.serverURL)
			require.Equal(t, tc.expected, got)
		})
	}

}

package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSort(t *testing.T) {
	versionNames := []string{
		"6.5.1",
		"0.1.3",
		"11.3",
	}

	var versions []Version

	for _, s := range versionNames {
		v, err := Make(s)
		require.NoError(t, err)

		versions = append(versions, v)
	}

	Sort(versions)

	var got []string
	for _, v := range versions {
		got = append(got, v.String())
	}

	expected := []string{"0.1.3", "6.5.1", "11.3"}

	assert.Equal(t, expected, got)
}

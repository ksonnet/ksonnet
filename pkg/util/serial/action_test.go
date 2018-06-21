package serial

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_RunActions(t *testing.T) {
	fn1 := func() error { return nil }
	fn2 := func() error { return nil }

	err := RunActions(fn1, fn2)
	require.NoError(t, err)
}

func Test_RunActions_failure(t *testing.T) {
	fn1 := func() error { return errors.New("failed") }
	fn2 := func() error { t.Fatal("should not have run"); return nil }

	err := RunActions(fn1, fn2)
	require.Error(t, err)
}

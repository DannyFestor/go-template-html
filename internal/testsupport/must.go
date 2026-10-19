package testsupport

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Must[T any](value T, err error) func(t *testing.T) T {
	return func(t *testing.T) T {
		t.Helper()

		require.NoError(t, err)

		return value
	}
}

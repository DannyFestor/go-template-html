//go:build integration

package valkey_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/testsupport/testvalkey"
)

func TestPingReachesValkey(t *testing.T) {
	t.Parallel()

	assert.NoError(t, open(t, testvalkey.New(t)).Ping(t.Context()))
}

func TestPingFailsWhileValkeyIsUnreachable(t *testing.T) {
	t.Parallel()

	assert.Error(t, open(t, testvalkey.Unreachable(t)).Ping(t.Context()))
}

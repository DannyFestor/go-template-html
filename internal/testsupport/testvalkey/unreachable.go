package testvalkey

import (
	"testing"

	"github.com/DannyFestor/go-template-html/internal/testsupport/testnet"
)

func Unreachable(t *testing.T) Instance {
	t.Helper()

	return Instance{Host: testnet.Loopback, Port: testnet.ClosedPort(t), KeyPrefix: "unreachable:"}
}

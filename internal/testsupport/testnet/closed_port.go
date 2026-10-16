package testnet

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

const Loopback = "127.0.0.1"

func ClosedPort(t *testing.T) int {
	t.Helper()

	var config net.ListenConfig

	listener, err := config.Listen(t.Context(), "tcp", net.JoinHostPort(Loopback, "0"))
	require.NoError(t, err)
	require.NoError(t, listener.Close())

	address, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok)

	return address.Port
}

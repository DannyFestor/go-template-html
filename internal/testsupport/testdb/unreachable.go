package testdb

import (
	"net"
	"testing"

	"github.com/stretchr/testify/require"
)

const (
	loopback        = "127.0.0.1"
	unreachableName = "unreachable"
)

func Unreachable(t *testing.T) Database {
	t.Helper()

	return Database{
		Host:     loopback,
		Port:     closedPort(t),
		User:     unreachableName,
		Password: unreachableName,
		Name:     unreachableName,
		Schema:   "public",
	}
}

func closedPort(t *testing.T) int {
	t.Helper()

	var config net.ListenConfig

	listener, err := config.Listen(t.Context(), "tcp", net.JoinHostPort(loopback, "0"))
	require.NoError(t, err)
	require.NoError(t, listener.Close())

	address, ok := listener.Addr().(*net.TCPAddr)
	require.True(t, ok)

	return address.Port
}

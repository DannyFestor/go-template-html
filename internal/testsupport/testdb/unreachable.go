package testdb

import (
	"testing"

	"github.com/DannyFestor/go-template-html/internal/testsupport/testnet"
)

const unreachableName = "unreachable"

func Unreachable(t *testing.T) Database {
	t.Helper()

	return Database{
		Host:     testnet.Loopback,
		Port:     testnet.ClosedPort(t),
		User:     unreachableName,
		Password: unreachableName,
		Name:     unreachableName,
		Schema:   "public",
	}
}

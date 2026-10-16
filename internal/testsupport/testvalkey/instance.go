package testvalkey

import (
	"crypto/rand"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/valkey"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testenv"
)

type Instance struct {
	Host      string
	Port      int
	KeyPrefix string
}

func New(t *testing.T) Instance {
	t.Helper()

	values, err := testenv.Required("VALKEY_HOST", "VALKEY_PORT")
	require.NoError(t, err)

	port, err := values.Port("VALKEY_PORT")
	require.NoError(t, err)

	return Instance{Host: values["VALKEY_HOST"], Port: port, KeyPrefix: "test:" + strings.ToLower(rand.Text()) + ":"}
}

func (i Instance) Config() valkey.Config {
	return valkey.Config{Host: i.Host, Port: i.Port, Password: "", KeyPrefix: i.KeyPrefix}
}

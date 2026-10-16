package testapp

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
)

func New(t *testing.T, config bootstrap.Config, logs io.Writer) *bootstrap.WebApp {
	t.Helper()

	webApp, err := bootstrap.NewWebApp(t.Context(), config, bootstrap.NewLogger(logs, config.App.Env, config.Log))
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, webApp.Close()) })

	return webApp
}

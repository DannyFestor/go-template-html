package testapp

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
)

type App struct {
	*bootstrap.WebApp

	Worker *bootstrap.WorkerApp
}

func New(t *testing.T, config bootstrap.Config, logs io.Writer) *App {
	t.Helper()

	logger := bootstrap.NewLogger(logs, config.App.Env, config.Log)

	webApp, err := bootstrap.NewWebApp(t.Context(), config, logger)
	require.NoError(t, err)
	t.Cleanup(func() { assert.NoError(t, webApp.Close()) })

	workerApp, err := bootstrap.NewWorkerApp(t.Context(), config, logger)
	require.NoError(t, err)
	t.Cleanup(workerApp.Close)

	return &App{WebApp: webApp, Worker: workerApp}
}

func (a *App) RunQueuedJobs(t *testing.T) {
	t.Helper()

	require.NoError(t, a.Worker.RunDue(t.Context()))
}

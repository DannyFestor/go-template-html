package bootstrap_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
)

func TestLoadConfigAppliesDefaults(t *testing.T) {
	t.Parallel()

	config, err := bootstrap.LoadConfig(missingDotenv(t), []string{"APP_ENV=local"})

	require.NoError(t, err)
	assert.Equal(t, bootstrap.Config{
		App: bootstrap.App{
			Name:            "Go Template",
			Env:             bootstrap.EnvironmentLocal,
			ShutdownTimeout: 15 * time.Second,
		},
		HTTP: bootstrap.HTTP{
			Addr:              ":8080",
			ReadHeaderTimeout: 5 * time.Second,
		},
		Log: bootstrap.Log{Level: slog.LevelInfo},
	}, config)
}

func TestLoadConfigReadsDotenv(t *testing.T) {
	t.Parallel()

	dotenv := writeDotenv(t, "APP_ENV=production\nAPP_NAME=Acme Shop\nLOG_LEVEL=debug\n")

	config, err := bootstrap.LoadConfig(dotenv, nil)

	require.NoError(t, err)
	assert.Equal(t, bootstrap.EnvironmentProduction, config.App.Env)
	assert.Equal(t, "Acme Shop", config.App.Name)
	assert.Equal(t, slog.LevelDebug, config.Log.Level)
}

func TestLoadConfigPrefersProcessEnvironmentOverDotenv(t *testing.T) {
	t.Parallel()

	dotenv := writeDotenv(t, "APP_ENV=local\nHTTP_ADDR=:9000\n")

	config, err := bootstrap.LoadConfig(dotenv, []string{"HTTP_ADDR=127.0.0.1:7000"})

	require.NoError(t, err)
	assert.Equal(t, "127.0.0.1:7000", config.HTTP.Addr)
}

func TestLoadConfigReportsEveryInvalidValueAtOnce(t *testing.T) {
	t.Parallel()

	_, err := bootstrap.LoadConfig(missingDotenv(t), []string{
		"LOG_LEVEL=loud",
		"SHUTDOWN_TIMEOUT=soon",
	})

	require.ErrorContains(t, err, "APP_ENV")
	require.ErrorContains(t, err, `"loud"`)
	require.ErrorContains(t, err, `"soon"`)
}

func TestLoadConfigRejectsUnknownEnvironment(t *testing.T) {
	t.Parallel()

	_, err := bootstrap.LoadConfig(missingDotenv(t), []string{"APP_ENV=staging"})

	assert.ErrorContains(t, err, `"staging"`)
}

func TestLoadConfigReportsUnreadableDotenv(t *testing.T) {
	t.Parallel()

	directory := t.TempDir()

	_, err := bootstrap.LoadConfig(directory, []string{"APP_ENV=local"})

	assert.ErrorContains(t, err, directory)
}

func missingDotenv(t *testing.T) string {
	t.Helper()

	return filepath.Join(t.TempDir(), ".env")
}

func writeDotenv(t *testing.T, content string) string {
	t.Helper()

	path := missingDotenv(t)
	require.NoError(t, os.WriteFile(path, []byte(content), 0o600))

	return path
}

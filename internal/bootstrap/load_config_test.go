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

func databaseValues() []string {
	return []string{"DB_HOST=db.internal", "DB_USER=app", "DB_PASSWORD=hunter2", "DB_NAME=shop"}
}

func requiredValues() []string {
	return append(databaseValues(), "APP_ENV=local")
}

func TestLoadConfigAppliesDefaults(t *testing.T) {
	t.Parallel()

	config, err := bootstrap.LoadConfig(missingDotenv(t), requiredValues())

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
			ReadTimeout:       15 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
			MaxBodyBytes:      1 << 20,
			ReadinessCacheTTL: 5 * time.Second,
		},
		DB: bootstrap.DB{
			Host:                     "db.internal",
			Port:                     5432,
			User:                     "app",
			Password:                 bootstrap.Secret("hunter2"),
			Name:                     "shop",
			Schema:                   "public",
			IdleInTransactionTimeout: time.Minute,
			MaxConnLifetime:          time.Hour,
			MaxConnLifetimeJitter:    5 * time.Minute,
			MaxConnIdleTime:          30 * time.Minute,
			ConnectTimeout:           5 * time.Second,
			Web:                      bootstrap.WebDB{MaxConns: 10, MinConns: 0, StatementTimeout: 10 * time.Second},
		},
		Log: bootstrap.Log{Level: slog.LevelInfo},
	}, config)
}

func TestLoadConfigRequiresTheDatabaseConnection(t *testing.T) {
	t.Parallel()

	_, err := bootstrap.LoadConfig(missingDotenv(t), []string{"APP_ENV=local"})

	for _, name := range []string{"DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME"} {
		assert.ErrorContains(t, err, name)
	}
}

func TestLoadConfigReadsDotenv(t *testing.T) {
	t.Parallel()

	dotenv := writeDotenv(t, "APP_ENV=production\nAPP_NAME=Acme Shop\nLOG_LEVEL=debug\n")

	config, err := bootstrap.LoadConfig(dotenv, databaseValues())

	require.NoError(t, err)
	assert.Equal(t, bootstrap.EnvironmentProduction, config.App.Env)
	assert.Equal(t, "Acme Shop", config.App.Name)
	assert.Equal(t, slog.LevelDebug, config.Log.Level)
}

func TestLoadConfigPrefersProcessEnvironmentOverDotenv(t *testing.T) {
	t.Parallel()

	dotenv := writeDotenv(t, "APP_ENV=local\nHTTP_ADDR=:9000\n")

	config, err := bootstrap.LoadConfig(dotenv, append(databaseValues(), "HTTP_ADDR=127.0.0.1:7000"))

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

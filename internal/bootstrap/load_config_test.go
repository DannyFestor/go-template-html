package bootstrap_test

import (
	"fmt"
	"log/slog"
	"net/netip"
	"net/url"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/appkey"
	"github.com/DannyFestor/go-template-html/internal/bootstrap"
)

const appKey = "AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8="

func connectionValues() []string {
	return []string{
		"APP_URL=http://localhost:8080",
		"APP_KEY=" + appKey,
		"DB_HOST=db.internal",
		"DB_USER=app",
		"DB_PASSWORD=hunter2",
		"DB_NAME=shop",
		"VALKEY_HOST=valkey.internal",
	}
}

func requiredValues() []string {
	return append(connectionValues(), "APP_ENV=local")
}

func TestLoadConfigAppliesDefaults(t *testing.T) {
	t.Parallel()

	appURL, err := url.Parse("http://localhost:8080")
	require.NoError(t, err)

	config, err := bootstrap.LoadConfig(missingDotenv(t), requiredValues())

	require.NoError(t, err)
	assert.Equal(t, bootstrap.Config{
		App: bootstrap.App{
			Name:            "Go Template",
			Env:             bootstrap.EnvironmentLocal,
			URL:             *appURL,
			Key:             bootstrap.AppKey{Secret: bootstrap.Secret(appKey)},
			ShutdownTimeout: 15 * time.Second,
		},
		HTTP: bootstrap.HTTP{
			Addr:               ":8080",
			ReadHeaderTimeout:  5 * time.Second,
			ReadTimeout:        15 * time.Second,
			WriteTimeout:       30 * time.Second,
			IdleTimeout:        120 * time.Second,
			MaxBodyBytes:       1 << 20,
			ReadinessCacheTTL:  5 * time.Second,
			CSRFTrustedOrigins: nil,
			TrustedProxies:     nil,
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
			Worker:                   bootstrap.WorkerDB{MaxConns: 12},
		},
		Queue: bootstrap.Queue{Concurrency: 10, PollInterval: time.Second},
		Valkey: bootstrap.Valkey{
			Host:      "valkey.internal",
			Port:      6379,
			Password:  bootstrap.Secret(""),
			KeyPrefix: "go-template-html:",
		},
		Session:         bootstrap.Session{IdleLifetime: 2 * time.Hour, AbsoluteLifetime: 24 * time.Hour},
		PasswordHashing: bootstrap.PasswordHashing{MemoryKiB: 19 * 1024, Iterations: 2, Parallelism: 1},
		SignedURL:       bootstrap.SignedURL{TTL: 2 * time.Hour},
		Log:             bootstrap.Log{Level: slog.LevelInfo},
	}, config)
}

func TestLoadConfigRequiresTheAppURLTheAppKeyAndTheConnections(t *testing.T) {
	t.Parallel()

	_, err := bootstrap.LoadConfig(missingDotenv(t), []string{"APP_ENV=local"})

	for _, name := range []string{"APP_URL", "APP_KEY", "DB_HOST", "DB_USER", "DB_PASSWORD", "DB_NAME", "VALKEY_HOST"} {
		assert.ErrorContains(t, err, name)
	}
}

func TestLoadConfigRejectsAMalformedAppKeyWithoutEchoingIt(t *testing.T) {
	t.Parallel()

	_, err := bootstrap.LoadConfig(missingDotenv(t), append(requiredValues(), "APP_KEY=c2hvcnQta2V5", "LOG_LEVEL=loud"))

	require.ErrorContains(t, err, "APP_KEY: "+appkey.ErrMalformedKey.Error())
	require.ErrorContains(t, err, `"loud"`)
	assert.NotContains(t, err.Error(), "c2hvcnQta2V5")
}

func TestLoadConfigKeepsTheAppKeyOutOfFormattedOutput(t *testing.T) {
	t.Parallel()

	config, err := bootstrap.LoadConfig(missingDotenv(t), requiredValues())

	require.NoError(t, err)
	assert.NotContains(t, fmt.Sprintf("%+v", config), appKey)
	assert.Equal(t, appKey, config.App.Key.Reveal())
}

func TestLoadConfigReadsTheTrustedOriginsAndProxiesAsLists(t *testing.T) {
	t.Parallel()

	config, err := bootstrap.LoadConfig(missingDotenv(t), append(requiredValues(),
		"CSRF_TRUSTED_ORIGINS=https://example.com,https://www.example.com",
		"TRUSTED_PROXIES=10.0.0.0/8,fd00::/8",
	))

	require.NoError(t, err)
	assert.Equal(t, []string{"https://example.com", "https://www.example.com"}, config.HTTP.CSRFTrustedOrigins)
	assert.Equal(t,
		[]netip.Prefix{netip.MustParsePrefix("10.0.0.0/8"), netip.MustParsePrefix("fd00::/8")},
		config.HTTP.TrustedProxies,
	)
}

func TestLoadConfigReadsDotenv(t *testing.T) {
	t.Parallel()

	dotenv := writeDotenv(t, "APP_ENV=production\nAPP_NAME=Acme Shop\nLOG_LEVEL=debug\n")

	config, err := bootstrap.LoadConfig(dotenv, connectionValues())

	require.NoError(t, err)
	assert.Equal(t, bootstrap.EnvironmentProduction, config.App.Env)
	assert.Equal(t, "Acme Shop", config.App.Name)
	assert.Equal(t, slog.LevelDebug, config.Log.Level)
}

func TestLoadConfigPrefersProcessEnvironmentOverDotenv(t *testing.T) {
	t.Parallel()

	dotenv := writeDotenv(t, "APP_ENV=local\nHTTP_ADDR=:9000\n")

	config, err := bootstrap.LoadConfig(dotenv, append(connectionValues(), "HTTP_ADDR=127.0.0.1:7000"))

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

package testapp

import (
	"log/slog"
	"net/url"
	"time"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testdb"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testvalkey"
)

const (
	MaxBodyBytes      = 1 << 20
	maxConnsPerServer = 4
	statementTimeout  = 5 * time.Second
	sessionLifetime   = time.Hour
)

func Config(name string, database testdb.Database, valkey testvalkey.Instance) bootstrap.Config {
	return bootstrap.Config{
		App: bootstrap.App{
			Name:            name,
			Env:             bootstrap.EnvironmentProduction,
			URL:             appURL(),
			ShutdownTimeout: time.Second,
		},
		HTTP: bootstrap.HTTP{
			Addr:               "",
			ReadHeaderTimeout:  time.Second,
			ReadTimeout:        time.Second,
			WriteTimeout:       time.Second,
			IdleTimeout:        time.Second,
			MaxBodyBytes:       MaxBodyBytes,
			ReadinessCacheTTL:  time.Second,
			CSRFTrustedOrigins: nil,
			TrustedProxies:     nil,
		},
		DB:      dbConfig(database),
		Valkey:  valkeyConfig(valkey),
		Session: bootstrap.Session{IdleLifetime: sessionLifetime, AbsoluteLifetime: sessionLifetime},
		Log:     bootstrap.Log{Level: slog.LevelInfo},
	}
}

func appURL() url.URL {
	return url.URL{
		Scheme:      "http",
		Opaque:      "",
		User:        nil,
		Host:        "127.0.0.1",
		Path:        "",
		RawPath:     "",
		OmitHost:    false,
		ForceQuery:  false,
		RawQuery:    "",
		Fragment:    "",
		RawFragment: "",
	}
}

func dbConfig(database testdb.Database) bootstrap.DB {
	return bootstrap.DB{
		Host:                     database.Host,
		Port:                     database.Port,
		User:                     database.User,
		Password:                 bootstrap.Secret(database.Password),
		Name:                     database.Name,
		Schema:                   database.Schema,
		IdleInTransactionTimeout: time.Minute,
		MaxConnLifetime:          time.Hour,
		MaxConnLifetimeJitter:    time.Minute,
		MaxConnIdleTime:          time.Minute,
		ConnectTimeout:           time.Second,
		Web: bootstrap.WebDB{
			MaxConns:         maxConnsPerServer,
			MinConns:         0,
			StatementTimeout: statementTimeout,
		},
	}
}

func valkeyConfig(valkey testvalkey.Instance) bootstrap.Valkey {
	return bootstrap.Valkey{
		Host:      valkey.Host,
		Port:      valkey.Port,
		Password:  bootstrap.Secret(""),
		KeyPrefix: valkey.KeyPrefix,
	}
}

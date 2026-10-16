package testapp

import (
	"log/slog"
	"time"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testdb"
)

const (
	MaxBodyBytes      = 1 << 20
	maxConnsPerServer = 4
	statementTimeout  = 5 * time.Second
)

func Config(name string, database testdb.Database) bootstrap.Config {
	return bootstrap.Config{
		App: bootstrap.App{Name: name, Env: bootstrap.EnvironmentProduction, ShutdownTimeout: time.Second},
		HTTP: bootstrap.HTTP{
			Addr:              "",
			ReadHeaderTimeout: time.Second,
			ReadTimeout:       time.Second,
			WriteTimeout:      time.Second,
			IdleTimeout:       time.Second,
			MaxBodyBytes:      MaxBodyBytes,
			ReadinessCacheTTL: time.Second,
		},
		DB:  dbConfig(database),
		Log: bootstrap.Log{Level: slog.LevelInfo},
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

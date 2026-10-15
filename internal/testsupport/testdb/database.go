package testdb

import (
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
)

type Database struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	Schema   string
}

func (d Database) PostgresConfig() postgres.Config {
	return postgres.Config{
		Host:                     d.Host,
		Port:                     d.Port,
		User:                     d.User,
		Password:                 d.Password,
		Database:                 d.Name,
		Schema:                   d.Schema,
		MaxConns:                 0,
		MinConns:                 0,
		StatementTimeout:         0,
		IdleInTransactionTimeout: 0,
		MaxConnLifetime:          0,
		MaxConnLifetimeJitter:    0,
		MaxConnIdleTime:          0,
		ConnectTimeout:           0,
	}
}

func (d Database) ConnString() string {
	return d.PostgresConfig().ConnString()
}

package bootstrap

import "github.com/DannyFestor/go-template-html/internal/adapter/postgres"

func webPostgresConfig(config DB) postgres.Config {
	pool := postgresConfig(config)
	pool.MaxConns = config.Web.MaxConns
	pool.MinConns = config.Web.MinConns
	pool.StatementTimeout = config.Web.StatementTimeout
	pool.IdleInTransactionTimeout = config.IdleInTransactionTimeout

	return pool
}

func workerPostgresConfig(config DB) postgres.Config {
	pool := postgresConfig(config)
	pool.MaxConns = config.Worker.MaxConns
	pool.IdleInTransactionTimeout = config.IdleInTransactionTimeout

	return pool
}

func postgresConfig(config DB) postgres.Config {
	return postgres.Config{
		Host:                     config.Host,
		Port:                     config.Port,
		User:                     config.User,
		Password:                 config.Password.Reveal(),
		Database:                 config.Name,
		Schema:                   config.Schema,
		MaxConns:                 0,
		MinConns:                 0,
		StatementTimeout:         0,
		IdleInTransactionTimeout: 0,
		MaxConnLifetime:          config.MaxConnLifetime,
		MaxConnLifetimeJitter:    config.MaxConnLifetimeJitter,
		MaxConnIdleTime:          config.MaxConnIdleTime,
		ConnectTimeout:           config.ConnectTimeout,
	}
}

package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"maps"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
)

type Pool struct {
	pool   *pgxpool.Pool
	schema string
}

func Open(ctx context.Context, config Config) (*Pool, error) {
	poolConfig, err := newPoolConfig(config)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("open the postgres pool: %w", err)
	}

	return &Pool{pool: pool, schema: config.Schema}, nil
}

func newPoolConfig(config Config) (*pgxpool.Config, error) {
	poolConfig, err := pgxpool.ParseConfig(config.ConnString())
	if err != nil {
		return nil, fmt.Errorf("configure the postgres pool: %w", err)
	}

	maps.Copy(poolConfig.ConnConfig.RuntimeParams, runtimeParams(config))
	poolConfig.ConnConfig.ConnectTimeout = config.ConnectTimeout
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnLifetimeJitter = config.MaxConnLifetimeJitter
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime

	if config.MaxConns > 0 {
		poolConfig.MaxConns = config.MaxConns
	}

	return poolConfig, nil
}

func (p *Pool) Ping(ctx context.Context) error {
	err := p.pool.Ping(ctx)
	if err != nil {
		return fmt.Errorf("ping postgres: %w", err)
	}

	return nil
}

func (p *Pool) EnsureSchema(ctx context.Context) error {
	_, err := p.pool.Exec(ctx, "CREATE SCHEMA IF NOT EXISTS "+pgx.Identifier{p.schema}.Sanitize())
	if err != nil {
		return fmt.Errorf("create schema %s: %w", p.schema, err)
	}

	return nil
}

func (p *Pool) OpenSQLDB() *sql.DB {
	return stdlib.OpenDBFromPool(p.pool)
}

func (p *Pool) Close() {
	p.pool.Close()
}

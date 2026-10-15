package bootstrap

import "time"

type DB struct {
	Host                     string        `env:"DB_HOST,required"`
	Port                     int           `env:"DB_PORT"                        envDefault:"5432"`
	User                     string        `env:"DB_USER,required"`
	Password                 Secret        `env:"DB_PASSWORD,required"`
	Name                     string        `env:"DB_NAME,required"`
	Schema                   string        `env:"DB_SCHEMA"                      envDefault:"public"`
	IdleInTransactionTimeout time.Duration `env:"DB_IDLE_IN_TRANSACTION_TIMEOUT" envDefault:"60s"`
	MaxConnLifetime          time.Duration `env:"DB_MAX_CONN_LIFETIME"           envDefault:"1h"`
	MaxConnLifetimeJitter    time.Duration `env:"DB_MAX_CONN_LIFETIME_JITTER"    envDefault:"5m"`
	MaxConnIdleTime          time.Duration `env:"DB_MAX_CONN_IDLE_TIME"          envDefault:"30m"`
	ConnectTimeout           time.Duration `env:"DB_CONNECT_TIMEOUT"             envDefault:"5s"`
	Web                      WebDB
}

type WebDB struct {
	MaxConns         int32         `env:"WEB_DB_MAX_CONNS"         envDefault:"10"`
	MinConns         int32         `env:"WEB_DB_MIN_CONNS"         envDefault:"0"`
	StatementTimeout time.Duration `env:"WEB_DB_STATEMENT_TIMEOUT" envDefault:"10s"`
}

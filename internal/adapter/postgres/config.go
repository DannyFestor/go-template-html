package postgres

import (
	"net"
	"net/url"
	"strconv"
	"time"
)

type Config struct {
	Host                     string
	Port                     int
	User                     string
	Password                 string
	Database                 string
	Schema                   string
	MaxConns                 int32
	MinConns                 int32
	StatementTimeout         time.Duration
	IdleInTransactionTimeout time.Duration
	MaxConnLifetime          time.Duration
	MaxConnLifetimeJitter    time.Duration
	MaxConnIdleTime          time.Duration
	ConnectTimeout           time.Duration
}

func (c Config) ConnString() string {
	address := url.URL{
		Scheme:      "postgres",
		Opaque:      "",
		User:        url.UserPassword(c.User, c.Password),
		Host:        net.JoinHostPort(c.Host, strconv.Itoa(c.Port)),
		Path:        "/" + c.Database,
		RawPath:     "",
		OmitHost:    false,
		ForceQuery:  false,
		RawQuery:    "",
		Fragment:    "",
		RawFragment: "",
	}

	return address.String()
}

package valkey

import (
	"fmt"
	"net"
	"strconv"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Host      string
	Port      int
	Password  string
	KeyPrefix string
}

func (c Config) options() (*redis.Options, error) {
	options, err := redis.ParseURL("redis://" + net.JoinHostPort(c.Host, strconv.Itoa(c.Port)))
	if err != nil {
		return nil, fmt.Errorf("configure the valkey client: %w", err)
	}

	options.Password = c.Password

	return options, nil
}

package bootstrap

import "time"

type HTTP struct {
	Addr              string        `env:"HTTP_ADDR"                envDefault:":8080"`
	ReadHeaderTimeout time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" envDefault:"5s"`
}

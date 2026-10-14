package bootstrap

import "time"

type HTTP struct {
	Addr              string        `env:"HTTP_ADDR"                envDefault:":8080"`
	ReadHeaderTimeout time.Duration `env:"HTTP_READ_HEADER_TIMEOUT" envDefault:"5s"`
	ReadTimeout       time.Duration `env:"HTTP_READ_TIMEOUT"        envDefault:"15s"`
	WriteTimeout      time.Duration `env:"HTTP_WRITE_TIMEOUT"       envDefault:"30s"`
	IdleTimeout       time.Duration `env:"HTTP_IDLE_TIMEOUT"        envDefault:"120s"`
	MaxBodyBytes      int64         `env:"HTTP_MAX_BODY_BYTES"      envDefault:"1048576"`
}

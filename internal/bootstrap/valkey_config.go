package bootstrap

import "github.com/DannyFestor/go-template-html/internal/adapter/valkey"

type Valkey struct {
	Host      string `env:"VALKEY_HOST,required"`
	Port      int    `env:"VALKEY_PORT"          envDefault:"6379"`
	Password  Secret `env:"VALKEY_PASSWORD"`
	KeyPrefix string `env:"VALKEY_KEY_PREFIX"    envDefault:"go-template-html:"`
}

func valkeyConfig(config Valkey) valkey.Config {
	return valkey.Config{
		Host:      config.Host,
		Port:      config.Port,
		Password:  config.Password.Reveal(),
		KeyPrefix: config.KeyPrefix,
	}
}

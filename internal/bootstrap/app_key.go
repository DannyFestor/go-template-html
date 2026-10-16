package bootstrap

import (
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/adapter/appkey"
)

type AppKey struct {
	Secret
}

func (k *AppKey) UnmarshalText(text []byte) error {
	_, err := appkey.Parse(string(text))
	if err != nil {
		return fmt.Errorf("APP_KEY: %w", err)
	}

	k.Secret = Secret(text)

	return nil
}

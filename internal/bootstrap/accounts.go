package bootstrap

import (
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/adapter/appkey"
	"github.com/DannyFestor/go-template-html/internal/adapter/argon2id"
	"github.com/DannyFestor/go-template-html/internal/adapter/clock"
	"github.com/DannyFestor/go-template-html/internal/adapter/fingerprint"
	"github.com/DannyFestor/go-template-html/internal/adapter/hmacdigest"
	"github.com/DannyFestor/go-template-html/internal/adapter/postgres"
	"github.com/DannyFestor/go-template-html/internal/adapter/uuidv7"
)

func newUserStore(pool *postgres.Pool) *postgres.UserStore {
	return postgres.NewUserStore(pool, uuidv7.NewGenerator(clock.System{}), clock.System{})
}

func newPasswordHasher(config PasswordHashing) (*argon2id.Hasher, error) {
	hasher, err := argon2id.NewHasher(argon2id.Params{
		MemoryKiB:   config.MemoryKiB,
		Iterations:  config.Iterations,
		Parallelism: config.Parallelism,
	})
	if err != nil {
		return nil, fmt.Errorf("hash passwords: %w", err)
	}

	return hasher, nil
}

func newCredentialFingerprints(key AppKey) (*fingerprint.Credentials, error) {
	subkey, err := subkeyFor(key, appkey.PurposeCredentialFingerprint)
	if err != nil {
		return nil, err
	}

	return fingerprint.NewCredentials(hmacdigest.NewDigester(subkey)), nil
}

func subkeyFor(key AppKey, purpose appkey.Purpose) ([]byte, error) {
	parsed, err := appkey.Parse(key.Reveal())
	if err != nil {
		return nil, fmt.Errorf("APP_KEY: %w", err)
	}

	subkey, err := parsed.Subkey(purpose)
	if err != nil {
		return nil, fmt.Errorf("APP_KEY: %w", err)
	}

	return subkey, nil
}

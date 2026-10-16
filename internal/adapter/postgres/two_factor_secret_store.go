package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres/internal/sqlcgen"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

type TwoFactorSecretStore struct {
	pool      *Pool
	encrypter encrypter
}

func NewTwoFactorSecretStore(pool *Pool, encrypter encrypter) *TwoFactorSecretStore {
	return &TwoFactorSecretStore{pool: pool, encrypter: encrypter}
}

func (s *TwoFactorSecretStore) Save(ctx context.Context, userID domain.UserID, secret domain.TwoFactorSecret) error {
	id, err := parseUserID(userID)
	if err != nil {
		return err
	}

	err = s.pool.queries(ctx).SaveTwoFactorSecret(ctx, sqlcgen.SaveTwoFactorSecretParams{
		UserID:          id,
		EncryptedSecret: s.encrypter.Encrypt(secret),
	})
	if err != nil {
		return fmt.Errorf("save the two-factor secret of user %s: %w", userID, err)
	}

	return nil
}

func (s *TwoFactorSecretStore) Find(ctx context.Context, userID domain.UserID) (domain.TwoFactorSecret, bool, error) {
	id, err := parseUserID(userID)
	if err != nil {
		return nil, false, err
	}

	encrypted, err := s.pool.queries(ctx).FindTwoFactorSecret(ctx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, false, nil
	}

	if err != nil {
		return nil, false, fmt.Errorf("find the two-factor secret of user %s: %w", userID, err)
	}

	secret, err := s.encrypter.Decrypt(encrypted)
	if err != nil {
		return nil, false, fmt.Errorf("decrypt the two-factor secret of user %s: %w", userID, err)
	}

	return secret, true, nil
}

func (s *TwoFactorSecretStore) UseStep(ctx context.Context, userID domain.UserID, step domain.TimeStep) (bool, error) {
	id, err := parseUserID(userID)
	if err != nil {
		return false, err
	}

	updated, err := s.pool.queries(ctx).UseTwoFactorStep(ctx, sqlcgen.UseTwoFactorStepParams{
		Step:   int64(step),
		UserID: id,
	})
	if err != nil {
		return false, fmt.Errorf("use the two-factor step %d of user %s: %w", step, userID, err)
	}

	return updated == 1, nil
}

func (s *TwoFactorSecretStore) Delete(ctx context.Context, userID domain.UserID) error {
	id, err := parseUserID(userID)
	if err != nil {
		return err
	}

	err = s.pool.queries(ctx).DeleteTwoFactorSecret(ctx, id)
	if err != nil {
		return fmt.Errorf("delete the two-factor secret of user %s: %w", userID, err)
	}

	return nil
}

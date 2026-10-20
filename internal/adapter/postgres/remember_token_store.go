package postgres

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres/internal/sqlcgen"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

type RememberTokenStore struct {
	pool *Pool
}

func NewRememberTokenStore(pool *Pool) *RememberTokenStore {
	return &RememberTokenStore{pool: pool}
}

func (s *RememberTokenStore) Issue(
	ctx context.Context,
	userID domain.UserID,
	token domain.RememberToken,
	expiresAt time.Time,
) error {
	id, err := parseUserID(userID)
	if err != nil {
		return err
	}

	err = s.pool.queries(ctx).IssueRememberToken(ctx, sqlcgen.IssueRememberTokenParams{
		Selector:        token.Selector,
		ValidatorDigest: validatorDigest(token),
		UserID:          id,
		ExpiresAt:       expiresAt,
	})
	if err != nil {
		return fmt.Errorf("issue a remember token to user %s: %w", userID, err)
	}

	return nil
}

func (s *RememberTokenStore) Consume(
	ctx context.Context,
	token domain.RememberToken,
	now time.Time,
) (domain.UserID, bool, error) {
	userID, err := s.pool.queries(ctx).ConsumeRememberToken(ctx, sqlcgen.ConsumeRememberTokenParams{
		Selector:        token.Selector,
		ValidatorDigest: validatorDigest(token),
		Now:             now,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}

	if err != nil {
		return "", false, fmt.Errorf("consume a remember token: %w", err)
	}

	return domain.UserID(userID.String()), true, nil
}

func (s *RememberTokenStore) Forget(ctx context.Context, token domain.RememberToken) error {
	err := s.pool.queries(ctx).ForgetRememberToken(ctx, token.Selector)
	if err != nil {
		return fmt.Errorf("forget a remember token: %w", err)
	}

	return nil
}

func (s *RememberTokenStore) DeleteExpired(ctx context.Context, now time.Time) error {
	err := s.pool.queries(ctx).DeleteExpiredRememberTokens(ctx, now)
	if err != nil {
		return fmt.Errorf("delete the remember tokens expired by %s: %w", now, err)
	}

	return nil
}

func validatorDigest(token domain.RememberToken) []byte {
	digest := sha256.Sum256([]byte(token.Validator))

	return digest[:]
}

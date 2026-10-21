package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres/internal/sqlcgen"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

type PasswordResetTokenStore struct {
	pool *Pool
}

func NewPasswordResetTokenStore(pool *Pool) *PasswordResetTokenStore {
	return &PasswordResetTokenStore{pool: pool}
}

func (s *PasswordResetTokenStore) Replace(
	ctx context.Context,
	userID domain.UserID,
	token domain.PasswordResetToken,
	expiresAt time.Time,
) error {
	id, err := parseUserID(userID)
	if err != nil {
		return err
	}

	err = s.pool.queries(ctx).ReplacePasswordResetToken(ctx, sqlcgen.ReplacePasswordResetTokenParams{
		UserID:      id,
		TokenDigest: tokenDigest(string(token)),
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		return fmt.Errorf("replace the password reset token of user %s: %w", userID, err)
	}

	return nil
}

func (s *PasswordResetTokenStore) Consume(
	ctx context.Context,
	token domain.PasswordResetToken,
	now time.Time,
) (domain.UserID, bool, error) {
	id, err := s.pool.queries(ctx).ConsumePasswordResetToken(ctx, sqlcgen.ConsumePasswordResetTokenParams{
		TokenDigest: tokenDigest(string(token)),
		Now:         now,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", false, nil
	}

	if err != nil {
		return "", false, fmt.Errorf("consume a password reset token: %w", err)
	}

	return domain.UserID(id.String()), true, nil
}

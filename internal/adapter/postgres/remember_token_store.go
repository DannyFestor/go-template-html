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

type RememberTokenStore struct {
	pool         *Pool
	replayWindow time.Duration
}

func NewRememberTokenStore(pool *Pool, replayWindow time.Duration) *RememberTokenStore {
	return &RememberTokenStore{pool: pool, replayWindow: replayWindow}
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
) (domain.RememberTokenUse, bool, error) {
	var unused domain.RememberTokenUse

	row, err := s.pool.queries(ctx).ConsumeRememberToken(ctx, sqlcgen.ConsumeRememberTokenParams{
		Now:             now,
		Selector:        token.Selector,
		ValidatorDigest: validatorDigest(token),
		ReplayableAfter: s.replayableAfter(now),
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return unused, false, nil
	}

	if err != nil {
		return unused, false, fmt.Errorf("consume a remember token: %w", err)
	}

	return domain.RememberTokenUse{UserID: domain.UserID(row.UserID.String()), Replay: row.Replay}, true, nil
}

func (s *RememberTokenStore) Forget(ctx context.Context, token domain.RememberToken) error {
	err := s.pool.queries(ctx).ForgetRememberToken(ctx, token.Selector)
	if err != nil {
		return fmt.Errorf("forget a remember token: %w", err)
	}

	return nil
}

func (s *RememberTokenStore) ForgetDevicesOf(ctx context.Context, userID domain.UserID) error {
	id, err := parseUserID(userID)
	if err != nil {
		return err
	}

	err = s.pool.queries(ctx).ForgetRememberTokensOfUser(ctx, id)
	if err != nil {
		return fmt.Errorf("forget the remember tokens of user %s: %w", userID, err)
	}

	return nil
}

func (s *RememberTokenStore) DeleteExpired(ctx context.Context, now time.Time) error {
	err := s.pool.queries(ctx).DeleteExpiredRememberTokens(ctx, sqlcgen.DeleteExpiredRememberTokensParams{
		Now:             now,
		ReplayableAfter: s.replayableAfter(now),
	})
	if err != nil {
		return fmt.Errorf("delete the remember tokens expired or replaced by %s: %w", now, err)
	}

	return nil
}

func (s *RememberTokenStore) replayableAfter(now time.Time) time.Time {
	return now.Add(-s.replayWindow)
}

func validatorDigest(token domain.RememberToken) []byte {
	return tokenDigest(token.Validator)
}

package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/DannyFestor/go-template-html/internal/adapter/postgres/internal/sqlcgen"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

const (
	uniqueViolation     = "23505"
	usersEmailUniqueKey = "users_email_key"
)

type UserStore struct {
	pool  *Pool
	ids   idGenerator
	clock clock
}

func NewUserStore(pool *Pool, ids idGenerator, clock clock) *UserStore {
	return &UserStore{pool: pool, ids: ids, clock: clock}
}

func (s *UserStore) Create(ctx context.Context, user domain.NewUser) (domain.User, error) {
	now := s.clock.Now()
	created := domain.User{
		ID:              domain.UserID(s.ids.NewID()),
		Name:            user.Name,
		Email:           user.Email,
		PasswordHash:    user.PasswordHash,
		EmailVerifiedAt: verifiedAt(user.EmailVerified, now),
		CreatedAt:       now,
	}

	id, err := parseUserID(created.ID)
	if err != nil {
		return domain.User{}, err
	}

	err = s.pool.queries(ctx).CreateUser(ctx, sqlcgen.CreateUserParams{
		ID:              id,
		Name:            created.Name,
		Email:           created.Email,
		Password:        string(created.PasswordHash),
		EmailVerifiedAt: nullableTime(created.EmailVerifiedAt),
		CreatedAt:       created.CreatedAt,
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("create the user %s: %w", created.Email, asEmailTaken(err))
	}

	return created, nil
}

func (s *UserStore) Find(ctx context.Context, userID domain.UserID) (domain.User, bool, error) {
	id, err := parseUserID(userID)
	if err != nil {
		return domain.User{}, false, err
	}

	row, err := s.pool.queries(ctx).FindUser(ctx, id)

	return foundUser(row, err, string(userID))
}

func (s *UserStore) FindByEmail(ctx context.Context, email string) (domain.User, bool, error) {
	row, err := s.pool.queries(ctx).FindUserByEmail(ctx, email)

	return foundUser(row, err, email)
}

func (s *UserStore) UpdatePasswordHash(ctx context.Context, userID domain.UserID, hash domain.PasswordHash) error {
	id, err := parseUserID(userID)
	if err != nil {
		return err
	}

	err = s.pool.queries(ctx).UpdateUserPassword(ctx, sqlcgen.UpdateUserPasswordParams{
		Password: string(hash),
		ID:       id,
	})
	if err != nil {
		return fmt.Errorf("update the password hash of user %s: %w", userID, err)
	}

	return nil
}

func (s *UserStore) MarkEmailVerified(ctx context.Context, userID domain.UserID, at time.Time) error {
	id, err := parseUserID(userID)
	if err != nil {
		return err
	}

	err = s.pool.queries(ctx).MarkUserEmailVerified(ctx, sqlcgen.MarkUserEmailVerifiedParams{VerifiedAt: at, ID: id})
	if err != nil {
		return fmt.Errorf("mark the email of user %s verified: %w", userID, err)
	}

	return nil
}

func verifiedAt(verified bool, now time.Time) time.Time {
	if !verified {
		return time.Time{}
	}

	return now
}

func nullableTime(at time.Time) *time.Time {
	if at.IsZero() {
		return nil
	}

	return &at
}

func timeOrZero(at *time.Time) time.Time {
	if at == nil {
		return time.Time{}
	}

	return *at
}

func foundUser(row sqlcgen.User, err error, lookup string) (domain.User, bool, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		var unknown domain.User

		return unknown, false, nil
	}

	if err != nil {
		return domain.User{}, false, fmt.Errorf("find the user %s: %w", lookup, err)
	}

	return userFromRow(row), true, nil
}

func userFromRow(row sqlcgen.User) domain.User {
	return domain.User{
		ID:              domain.UserID(row.ID.String()),
		Name:            row.Name,
		Email:           row.Email,
		PasswordHash:    domain.PasswordHash(row.Password),
		EmailVerifiedAt: timeOrZero(row.EmailVerifiedAt),
		CreatedAt:       row.CreatedAt,
	}
}

func asEmailTaken(err error) error {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok && pgErr.Code == uniqueViolation && pgErr.ConstraintName == usersEmailUniqueKey {
		return domain.ErrEmailTaken
	}

	return err
}

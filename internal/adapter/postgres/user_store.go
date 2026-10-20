package postgres

import (
	"context"
	"errors"
	"fmt"

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
	created := domain.User{
		ID:           domain.UserID(s.ids.NewID()),
		Name:         user.Name,
		Email:        user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt:    s.clock.Now(),
	}

	id, err := parseUserID(created.ID)
	if err != nil {
		return domain.User{}, err
	}

	err = s.pool.queries(ctx).CreateUser(ctx, sqlcgen.CreateUserParams{
		ID:        id,
		Name:      created.Name,
		Email:     created.Email,
		Password:  string(created.PasswordHash),
		CreatedAt: created.CreatedAt,
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
	if errors.Is(err, pgx.ErrNoRows) {
		var unknown domain.User

		return unknown, false, nil
	}

	if err != nil {
		return domain.User{}, false, fmt.Errorf("find the user %s: %w", userID, err)
	}

	return userFromRow(row), true, nil
}

func userFromRow(row sqlcgen.User) domain.User {
	return domain.User{
		ID:           domain.UserID(row.ID.String()),
		Name:         row.Name,
		Email:        row.Email,
		PasswordHash: domain.PasswordHash(row.Password),
		CreatedAt:    row.CreatedAt,
	}
}

func asEmailTaken(err error) error {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if ok && pgErr.Code == uniqueViolation && pgErr.ConstraintName == usersEmailUniqueKey {
		return domain.ErrEmailTaken
	}

	return err
}

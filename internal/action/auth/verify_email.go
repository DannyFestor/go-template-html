package auth

import (
	"context"
	"errors"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

var ErrVerificationLinkMismatch = errors.New("the verification link was issued for another user or email address")

type VerifyEmail struct {
	links verificationLinkMatcher
	users emailVerifier
	clock clock
}

type VerifyEmailIn struct {
	User       domain.User
	LinkUserID domain.UserID
	LinkHash   string
}

type VerifyEmailOut struct{}

func NewVerifyEmail(links verificationLinkMatcher, users emailVerifier, clock clock) *VerifyEmail {
	return &VerifyEmail{links: links, users: users, clock: clock}
}

func (v *VerifyEmail) Run(ctx context.Context, in VerifyEmailIn) (VerifyEmailOut, error) {
	if in.LinkUserID != in.User.ID || !v.links.Matches(in.User.Email, in.LinkHash) {
		return VerifyEmailOut{}, fmt.Errorf("verify the email of user %s: %w", in.User.ID, ErrVerificationLinkMismatch)
	}

	err := v.users.MarkEmailVerified(ctx, in.User.ID, v.clock.Now())
	if err != nil {
		return VerifyEmailOut{}, fmt.Errorf("verify the email of user %s: %w", in.User.ID, err)
	}

	return VerifyEmailOut{}, nil
}

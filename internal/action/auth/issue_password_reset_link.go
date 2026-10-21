package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type IssuePasswordResetLink struct {
	generator passwordResetTokenGenerator
	tokens    passwordResetTokenReplacer
	links     passwordResetLinker
	clock     clock
	ttl       time.Duration
}

type IssuePasswordResetLinkIn struct {
	UserID domain.UserID
}

type IssuePasswordResetLinkOut struct {
	URL string
}

func NewIssuePasswordResetLink(
	generator passwordResetTokenGenerator,
	tokens passwordResetTokenReplacer,
	links passwordResetLinker,
	clock clock,
	ttl time.Duration,
) *IssuePasswordResetLink {
	return &IssuePasswordResetLink{generator: generator, tokens: tokens, links: links, clock: clock, ttl: ttl}
}

func (i *IssuePasswordResetLink) Run(
	ctx context.Context,
	in IssuePasswordResetLinkIn,
) (IssuePasswordResetLinkOut, error) {
	token := i.generator.New()

	err := i.tokens.Replace(ctx, in.UserID, token, i.clock.Now().Add(i.ttl))
	if err != nil {
		return IssuePasswordResetLinkOut{}, fmt.Errorf("issue a password reset link to user %s: %w", in.UserID, err)
	}

	return IssuePasswordResetLinkOut{URL: i.links.PasswordResetURL(token)}, nil
}

package auth

import (
	"context"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type RequestPasswordReset struct {
	users      userByEmailFinder
	transactor transactor
	issue      passwordResetLinkIssuer
	mailer     resetPasswordMailer
}

type RequestPasswordResetIn struct {
	Email string
}

type RequestPasswordResetOut struct{}

func NewRequestPasswordReset(
	users userByEmailFinder,
	transactor transactor,
	issue passwordResetLinkIssuer,
	mailer resetPasswordMailer,
) *RequestPasswordReset {
	return &RequestPasswordReset{users: users, transactor: transactor, issue: issue, mailer: mailer}
}

func (r *RequestPasswordReset) Run(ctx context.Context, in RequestPasswordResetIn) (RequestPasswordResetOut, error) {
	err := r.request(ctx, in.Email)
	if err != nil {
		return RequestPasswordResetOut{}, fmt.Errorf("request a password reset for %s: %w", in.Email, err)
	}

	return RequestPasswordResetOut{}, nil
}

func (r *RequestPasswordReset) request(ctx context.Context, email string) error {
	user, found, err := r.users.FindByEmail(ctx, email)
	if err != nil {
		return fmt.Errorf("find the user: %w", err)
	}

	if !found {
		return nil
	}

	err = r.transactor.WithinTransaction(ctx, func(ctx context.Context) error {
		return r.mailLink(ctx, user)
	})
	if err != nil {
		return fmt.Errorf("mail a password reset link: %w", err)
	}

	return nil
}

func (r *RequestPasswordReset) mailLink(ctx context.Context, user domain.User) error {
	issued, err := r.issue.Run(ctx, IssuePasswordResetLinkIn{UserID: user.ID})
	if err != nil {
		return fmt.Errorf("issue the link: %w", err)
	}

	err = r.mailer.Send(ctx, domain.ResetPasswordMail{To: user.Email, URL: issued.URL})
	if err != nil {
		return fmt.Errorf("queue the mail: %w", err)
	}

	return nil
}

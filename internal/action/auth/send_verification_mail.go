package auth

import (
	"context"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type SendVerificationMail struct {
	links  verificationLinkIssuer
	mailer verificationMailer
}

type SendVerificationMailIn struct {
	User domain.User
}

type SendVerificationMailOut struct{}

func NewSendVerificationMail(links verificationLinkIssuer, mailer verificationMailer) *SendVerificationMail {
	return &SendVerificationMail{links: links, mailer: mailer}
}

func (s *SendVerificationMail) Run(ctx context.Context, in SendVerificationMailIn) (SendVerificationMailOut, error) {
	err := s.mailer.Send(ctx, domain.VerifyEmailMail{To: in.User.Email, URL: s.links.For(in.User)})
	if err != nil {
		return SendVerificationMailOut{}, fmt.Errorf("send the verification mail to user %s: %w", in.User.ID, err)
	}

	return SendVerificationMailOut{}, nil
}

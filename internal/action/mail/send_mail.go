package mail

import (
	"context"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type SendMail struct {
	sender sender
}

type SendMailIn struct {
	Message domain.MailMessage
}

type SendMailOut struct{}

func NewSendMail(sender sender) *SendMail {
	return &SendMail{sender: sender}
}

func (s *SendMail) Run(ctx context.Context, in SendMailIn) (SendMailOut, error) {
	err := s.sender.Send(ctx, in.Message)
	if err != nil {
		return SendMailOut{}, fmt.Errorf("send the mail %q: %w", in.Message.Subject, err)
	}

	return SendMailOut{}, nil
}

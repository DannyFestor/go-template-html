package mail

import (
	"context"
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/action/mail"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

type SendMail struct {
	Message domain.MailMessage `json:"message"`
}

func (SendMail) Kind() string {
	return "send_mail"
}

type SendMailHandler struct {
	send mailSender
}

func NewSendMailHandler(send mailSender) *SendMailHandler {
	return &SendMailHandler{send: send}
}

func (h *SendMailHandler) Handle(ctx context.Context, job SendMail) error {
	_, err := h.send.Run(ctx, mail.SendMailIn{Message: job.Message})
	if err != nil {
		return fmt.Errorf("handle the %s job: %w", job.Kind(), err)
	}

	return nil
}

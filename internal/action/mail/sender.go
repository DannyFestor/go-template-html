package mail

import (
	"context"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

//go:generate go tool mockgen -source=sender.go -destination=mock_sender_test.go -package=mail -mock_names=sender=MockSender
type sender interface {
	Send(ctx context.Context, message domain.MailMessage) error
}

package testapp

import (
	"context"
	"slices"
	"sync"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type Outbox struct {
	mutex sync.Mutex
	sent  []domain.MailMessage
}

func (o *Outbox) Send(_ context.Context, message domain.MailMessage) error {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	o.sent = append(o.sent, message)

	return nil
}

func (o *Outbox) Sent() []domain.MailMessage {
	o.mutex.Lock()
	defer o.mutex.Unlock()

	return slices.Clone(o.sent)
}

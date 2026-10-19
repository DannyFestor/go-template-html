package mail

import (
	"context"
	"fmt"
)

type Mailer[M any] struct {
	compose func(mail M) Content
	jobs    jobEnqueuer
}

func NewMailer[M any](compose func(mail M) Content, jobs jobEnqueuer) *Mailer[M] {
	return &Mailer[M]{compose: compose, jobs: jobs}
}

func (m *Mailer[M]) Send(ctx context.Context, mail M) error {
	message, err := m.compose(mail).render(ctx)
	if err != nil {
		return err
	}

	err = m.jobs.Enqueue(ctx, SendMail{Message: message})
	if err != nil {
		return fmt.Errorf("queue the mail %q: %w", message.Subject, err)
	}

	return nil
}

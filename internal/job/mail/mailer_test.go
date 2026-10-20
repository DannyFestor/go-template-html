package mail_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/domain"
	mailjob "github.com/DannyFestor/go-template-html/internal/job/mail"
)

var (
	errRender  = errors.New("render failed")
	errEnqueue = errors.New("enqueue failed")
)

type greeting struct {
	To   string
	Name string
}

type writtenPart string

func (p writtenPart) Render(_ context.Context, w io.Writer) error {
	_, err := io.WriteString(w, string(p))
	if err != nil {
		return fmt.Errorf("write the part: %w", err)
	}

	return nil
}

type failingPart struct{}

func (failingPart) Render(context.Context, io.Writer) error {
	return errRender
}

func greetingContent(mail greeting) mailjob.Content {
	return mailjob.Content{
		To:      mail.To,
		Subject: "Hello " + mail.Name,
		HTML:    writtenPart("<p>Hello " + mail.Name + "</p>"),
		Text:    writtenPart("Hello " + mail.Name),
	}
}

func TestSendQueuesTheMailRenderedIntoItsHTMLAndTextParts(t *testing.T) {
	t.Parallel()

	jobs := mailjob.NewMockJobEnqueuer(gomock.NewController(t))
	jobs.EXPECT().Enqueue(gomock.Any(), mailjob.SendMail{Message: domain.MailMessage{
		To:      "ada@example.test",
		Subject: "Hello Ada",
		HTML:    "<p>Hello Ada</p>",
		Text:    "Hello Ada",
	}})

	err := mailjob.NewMailer(greetingContent, jobs).Send(t.Context(), greeting{To: "ada@example.test", Name: "Ada"})

	require.NoError(t, err)
}

func TestSendQueuesNothingWhenAPartFailsToRender(t *testing.T) {
	t.Parallel()

	for name, content := range map[string]mailjob.Content{
		"html": {To: "ada@example.test", Subject: "Hello", HTML: failingPart{}, Text: writtenPart("Hello")},
		"text": {To: "ada@example.test", Subject: "Hello", HTML: writtenPart("<p>Hello</p>"), Text: failingPart{}},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			jobs := mailjob.NewMockJobEnqueuer(gomock.NewController(t))
			compose := func(greeting) mailjob.Content { return content }

			err := mailjob.NewMailer(compose, jobs).Send(t.Context(), greeting{To: "ada@example.test", Name: "Ada"})

			require.ErrorIs(t, err, errRender)
		})
	}
}

func TestSendFailsWhenTheMailCannotBeQueued(t *testing.T) {
	t.Parallel()

	jobs := mailjob.NewMockJobEnqueuer(gomock.NewController(t))
	jobs.EXPECT().Enqueue(gomock.Any(), gomock.Any()).Return(errEnqueue)

	err := mailjob.NewMailer(greetingContent, jobs).Send(t.Context(), greeting{To: "ada@example.test", Name: "Ada"})

	require.ErrorIs(t, err, errEnqueue)
}

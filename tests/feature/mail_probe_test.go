//go:build feature

package feature_test

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"text/template"

	mailjob "github.com/DannyFestor/go-template-html/internal/job/mail"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testapp"
	"github.com/DannyFestor/go-template-html/internal/view/mail"
)

const (
	probeMail          = "/probe/mail"
	probeMailRecipient = "ada@example.test"
	probeMailSubject   = "Probe mail"
	probeMailName      = "Ada <Lovelace>"
	probeMailText      = "Hello {{.Name}},\nthis mail was rendered when it was queued.\n"
)

type greetingMail struct {
	To   string
	Name string
}

type greetingBody struct {
	name string
}

func (b greetingBody) Render(_ context.Context, w io.Writer) error {
	_, err := io.WriteString(w, "<p>Hello "+html.EscapeString(b.name)+"</p>")
	if err != nil {
		return fmt.Errorf("render the greeting: %w", err)
	}

	return nil
}

func composeGreeting(greeting greetingMail) mailjob.Content {
	return mailjob.Content{
		To:      greeting.To,
		Subject: probeMailSubject,
		HTML:    mail.Layout(probeMailSubject, greetingBody{name: greeting.Name}),
		Text:    mail.Text(template.Must(template.New("greeting").Parse(probeMailText)), greeting),
	}
}

func mountMailProbe(testApp *testapp.App) {
	mailer := mailjob.NewMailer(composeGreeting, testApp.Jobs)

	testApp.Mux.Handle("POST "+probeMail, testApp.Responder.Handle(
		func(w http.ResponseWriter, r *http.Request) error {
			err := mailer.Send(r.Context(), greetingMail{To: probeMailRecipient, Name: probeMailName})
			if err != nil {
				return fmt.Errorf("send the probe mail: %w", err)
			}

			w.WriteHeader(http.StatusNoContent)

			return nil
		},
	))
}

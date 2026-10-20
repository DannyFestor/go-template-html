package mail

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type component interface {
	Render(ctx context.Context, w io.Writer) error
}

type Content struct {
	To      string
	Subject string
	HTML    component
	Text    component
}

func (c Content) render(ctx context.Context) (domain.MailMessage, error) {
	html, err := renderPart(ctx, c.HTML)
	if err != nil {
		return domain.MailMessage{}, fmt.Errorf("render the HTML part of the mail %q: %w", c.Subject, err)
	}

	text, err := renderPart(ctx, c.Text)
	if err != nil {
		return domain.MailMessage{}, fmt.Errorf("render the text part of the mail %q: %w", c.Subject, err)
	}

	return domain.MailMessage{To: c.To, Subject: c.Subject, HTML: html, Text: text}, nil
}

func renderPart(ctx context.Context, part component) (string, error) {
	var rendered strings.Builder

	err := part.Render(ctx, &rendered)

	return rendered.String(), err
}

package smtp

import (
	"bytes"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"mime/quotedprintable"
	"net/mail"
	"net/textproto"
	"strings"
	"time"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

const crlf = "\r\n"

func compose(from mail.Address, message domain.MailMessage, date time.Time) ([]byte, error) {
	var body bytes.Buffer

	parts := multipart.NewWriter(&body)

	err := errors.Join(
		writePart(parts, "text/plain; charset=utf-8", message.Text),
		writePart(parts, "text/html; charset=utf-8", message.HTML),
		parts.Close(),
	)
	if err != nil {
		return nil, fmt.Errorf("compose the multipart/alternative body: %w", err)
	}

	header := strings.Join(headerLines(from, message, date, parts.Boundary()), crlf)

	return append([]byte(header+crlf+crlf), body.Bytes()...), nil
}

func headerLines(from mail.Address, message domain.MailMessage, date time.Time, boundary string) []string {
	to := mail.Address{Name: "", Address: message.To}

	return []string{
		"From: " + from.String(),
		"To: " + to.String(),
		"Subject: " + mime.QEncoding.Encode("utf-8", message.Subject),
		"Date: " + date.Format(time.RFC1123Z),
		"MIME-Version: 1.0",
		"Content-Type: " + mime.FormatMediaType("multipart/alternative", map[string]string{"boundary": boundary}),
	}
}

func writePart(parts *multipart.Writer, contentType, content string) error {
	part, err := parts.CreatePart(textproto.MIMEHeader{
		"Content-Type":              {contentType},
		"Content-Transfer-Encoding": {"quoted-printable"},
	})
	if err != nil {
		return fmt.Errorf("start the %s part: %w", contentType, err)
	}

	encoder := quotedprintable.NewWriter(part)

	_, err = encoder.Write([]byte(content))
	if err != nil {
		return fmt.Errorf("write the %s part: %w", contentType, err)
	}

	err = encoder.Close()
	if err != nil {
		return fmt.Errorf("finish the %s part: %w", contentType, err)
	}

	return nil
}

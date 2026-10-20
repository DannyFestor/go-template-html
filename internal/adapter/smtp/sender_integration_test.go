//go:build integration

package smtp_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/smtp"
	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testmailpit"
)

func TestSendDeliversTheHTMLAndTextPartsFromTheConfiguredSender(t *testing.T) {
	t.Parallel()

	mailpit := testmailpit.New(t)
	recipient := testmailpit.NewRecipient()
	sample := domain.MailMessage{
		To:      recipient,
		Subject: "Bestätige deine Adresse",
		HTML:    `<p>Hello <a href="https://example.test/verify?signature=abc">there</a></p>`,
		Text:    "Hello there, follow " + strings.Repeat("this very long link ", 8) + "to verify.\nThanks",
	}

	err := smtp.NewSender(mailpit.Config(fromName, fromAddress), fixedClock{}).Send(t.Context(), sample)
	require.NoError(t, err)

	assert.Equal(t, testmailpit.Message{
		From:    testmailpit.Address{Name: fromName, Address: fromAddress},
		To:      []testmailpit.Address{{Name: "", Address: recipient}},
		Subject: sample.Subject,
		Date:    fixedClock{}.Now(),
		Text:    strings.ReplaceAll(sample.Text, "\n", "\r\n"),
		HTML:    sample.HTML,
	}, mailpit.MessageTo(t, recipient))
}

func TestSendWithStartTLSRefusesAServerThatOffersNoTLS(t *testing.T) {
	t.Parallel()

	config := testmailpit.New(t).Config(fromName, fromAddress)
	config.Encryption = smtp.EncryptionStartTLS

	err := smtp.NewSender(config, fixedClock{}).Send(t.Context(), sampleMessage())

	assert.ErrorIs(t, err, smtp.ErrNoStartTLS)
}

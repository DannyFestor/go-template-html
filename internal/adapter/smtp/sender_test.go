package smtp_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/smtp"
	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testnet"
)

const (
	fromName    = "Grüne Wiese"
	fromAddress = "hello@example.test"
)

type fixedClock struct{}

func (fixedClock) Now() time.Time {
	return time.Date(2026, 10, 20, 9, 30, 0, 0, time.UTC)
}

func sampleMessage() domain.MailMessage {
	return domain.MailMessage{To: "someone@example.test", Subject: "Sample", HTML: "<p>Sample</p>", Text: "Sample"}
}

func TestSendFailsWhileTheMailServerIsUnreachable(t *testing.T) {
	t.Parallel()

	config := smtp.Config{
		Host:        testnet.Loopback,
		Port:        testnet.ClosedPort(t),
		Username:    "",
		Password:    "",
		Encryption:  smtp.EncryptionNone,
		FromName:    fromName,
		FromAddress: fromAddress,
	}

	err := smtp.NewSender(config, fixedClock{}).Send(t.Context(), sampleMessage())

	assert.ErrorContains(t, err, "deliver over SMTP: connect to ")
}

func TestEncryptionAcceptsNoneStartTLSAndTLS(t *testing.T) {
	t.Parallel()

	for _, want := range []smtp.Encryption{smtp.EncryptionNone, smtp.EncryptionStartTLS, smtp.EncryptionTLS} {
		var parsed smtp.Encryption

		require.NoError(t, parsed.UnmarshalText([]byte(want)))
		assert.Equal(t, want, parsed)
	}
}

func TestEncryptionRejectsAnUnknownMode(t *testing.T) {
	t.Parallel()

	var parsed smtp.Encryption

	assert.ErrorContains(t, parsed.UnmarshalText([]byte("ssl")), `unknown mail encryption "ssl"`)
}

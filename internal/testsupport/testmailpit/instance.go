package testmailpit

import (
	"crypto/rand"
	"net"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/smtp"
	"github.com/DannyFestor/go-template-html/internal/testsupport/testenv"
)

type Instance struct {
	Host     string
	SMTPPort int
	apiURL   string
}

func New(t *testing.T) Instance {
	t.Helper()

	values, err := testenv.Required("MAIL_HOST", "MAIL_PORT", "MAILPIT_UI_PORT")
	require.NoError(t, err)

	smtpPort, err := values.Port("MAIL_PORT")
	require.NoError(t, err)

	apiPort, err := values.Port("MAILPIT_UI_PORT")
	require.NoError(t, err)

	return Instance{
		Host:     values["MAIL_HOST"],
		SMTPPort: smtpPort,
		apiURL:   "http://" + net.JoinHostPort(values["MAIL_HOST"], strconv.Itoa(apiPort)) + "/api/v1",
	}
}

func NewRecipient() string {
	return strings.ToLower(rand.Text()) + "@example.test"
}

func (i Instance) Config(fromName, fromAddress string) smtp.Config {
	return smtp.Config{
		Host:        i.Host,
		Port:        i.SMTPPort,
		Username:    "",
		Password:    "",
		Encryption:  smtp.EncryptionNone,
		FromName:    fromName,
		FromAddress: fromAddress,
	}
}

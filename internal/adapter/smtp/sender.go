package smtp

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/smtp"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

var ErrNoStartTLS = errors.New("the mail server offers no STARTTLS")

type Sender struct {
	config Config
	clock  clock
}

func NewSender(config Config, clock clock) *Sender {
	return &Sender{config: config, clock: clock}
}

func (s *Sender) Send(ctx context.Context, message domain.MailMessage) error {
	content, err := compose(s.config.from(), message, s.clock.Now())
	if err != nil {
		return err
	}

	err = s.deliver(ctx, message.To, content)
	if err != nil {
		return fmt.Errorf("deliver over SMTP: %w", err)
	}

	return nil
}

func (s *Sender) deliver(ctx context.Context, to string, content []byte) error {
	connection, err := s.dial(ctx)
	if err != nil {
		return err
	}

	defer func() { _ = connection.Close() }()

	stopClosingOnCancel := closeOnCancel(ctx, connection)
	defer stopClosingOnCancel()

	client, err := smtp.NewClient(connection, s.config.Host)
	if err != nil {
		return fmt.Errorf("greet %s: %w", s.config.address(), err)
	}

	err = s.converse(client, to, content)
	if err != nil {
		return errors.Join(err, context.Cause(ctx))
	}

	return nil
}

func closeOnCancel(ctx context.Context, connection net.Conn) func() bool {
	return context.AfterFunc(ctx, func() { _ = connection.Close() })
}

func (s *Sender) converse(client *smtp.Client, to string, content []byte) error {
	err := s.startTLS(client)
	if err != nil {
		return err
	}

	err = s.authenticate(client)
	if err != nil {
		return err
	}

	err = s.transmit(client, to, content)
	if err != nil {
		return err
	}

	return wrap("quit", client.Quit())
}

func (s *Sender) startTLS(client *smtp.Client) error {
	if s.config.Encryption != EncryptionStartTLS {
		return nil
	}

	if supported, _ := client.Extension("STARTTLS"); !supported {
		return ErrNoStartTLS
	}

	return wrap("start TLS", client.StartTLS(s.config.tlsConfig()))
}

func (s *Sender) authenticate(client *smtp.Client) error {
	if s.config.Username == "" {
		return nil
	}

	return wrap("authenticate", client.Auth(smtp.PlainAuth("", s.config.Username, s.config.Password, s.config.Host)))
}

func (s *Sender) transmit(client *smtp.Client, to string, content []byte) error {
	err := client.Mail(s.config.FromAddress)
	if err != nil {
		return wrap("set the sender", err)
	}

	err = client.Rcpt(to)
	if err != nil {
		return wrap("set the recipient", err)
	}

	data, err := client.Data()
	if err != nil {
		return wrap("start the data", err)
	}

	return wrap("write the data", writeAndClose(data, content))
}

func writeAndClose(data io.WriteCloser, content []byte) error {
	_, err := data.Write(content)

	return errors.Join(err, data.Close())
}

func wrap(step string, err error) error {
	if err != nil {
		return fmt.Errorf("%s: %w", step, err)
	}

	return nil
}

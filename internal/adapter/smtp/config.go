package smtp

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/mail"
	"slices"
	"strconv"
)

type Encryption string

const (
	EncryptionNone     Encryption = "none"
	EncryptionStartTLS Encryption = "starttls"
	EncryptionTLS      Encryption = "tls"
)

var errUnknownEncryption = errors.New("unknown mail encryption")

type Config struct {
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  Encryption
	FromName    string
	FromAddress string
}

func encryptions() []Encryption {
	return []Encryption{EncryptionNone, EncryptionStartTLS, EncryptionTLS}
}

func (e *Encryption) UnmarshalText(text []byte) error {
	candidate := Encryption(text)
	if !slices.Contains(encryptions(), candidate) {
		return fmt.Errorf("%w %q, want one of %q", errUnknownEncryption, text, encryptions())
	}

	*e = candidate

	return nil
}

func (c Config) address() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

func (c Config) from() mail.Address {
	return mail.Address{Name: c.FromName, Address: c.FromAddress}
}

func (c Config) tlsConfig() *tls.Config {
	config := new(tls.Config)
	config.ServerName = c.Host
	config.MinVersion = tls.VersionTLS12

	return config
}

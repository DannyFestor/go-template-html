package smtp

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
)

func (s *Sender) dial(ctx context.Context) (net.Conn, error) {
	plain := new(net.Dialer)
	dialContext := plain.DialContext

	if s.config.Encryption == EncryptionTLS {
		encrypted := &tls.Dialer{NetDialer: plain, Config: s.config.tlsConfig()}
		dialContext = encrypted.DialContext
	}

	connection, err := dialContext(ctx, "tcp", s.config.address())
	if err != nil {
		return nil, fmt.Errorf("connect to %s: %w", s.config.address(), err)
	}

	return connection, nil
}

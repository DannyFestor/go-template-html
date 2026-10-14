package middleware

import (
	"bufio"
	"fmt"
	"net"
	"net/http"
)

// wrappedWriter keeps Flush and Hijack reachable by type assertion, because /ws hijacks through every wrapper.
type wrappedWriter struct {
	http.ResponseWriter
}

func (w wrappedWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}

func (w wrappedWriter) Write(content []byte) (int, error) {
	written, err := w.ResponseWriter.Write(content)
	if err != nil {
		return written, fmt.Errorf("write the response: %w", err)
	}

	return written, nil
}

func (w wrappedWriter) Flush() {
	_ = http.NewResponseController(w.ResponseWriter).Flush()
}

func (w wrappedWriter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	conn, buffered, err := http.NewResponseController(w.ResponseWriter).Hijack()
	if err != nil {
		return nil, nil, fmt.Errorf("hijack the connection: %w", err)
	}

	return conn, buffered, nil
}

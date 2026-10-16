package middleware

import (
	"bufio"
	"errors"
	"net"
	"net/http"
)

var errSessionNotSaved = errors.New("the session could not be saved, so the connection was not hijacked")

// sessionCommitter saves the session on the response's first byte, because the cookie can't be set once headers flush.
type sessionCommitter struct {
	wrappedWriter

	request   *http.Request
	keeper    *sessionKeeper
	state     *sessionState
	renderer  errorRenderer
	committed bool
	failed    bool
}

func (c *sessionCommitter) WriteHeader(status int) {
	if c.commit() {
		c.ResponseWriter.WriteHeader(status)
	}
}

func (c *sessionCommitter) Write(content []byte) (int, error) {
	if !c.commit() {
		return len(content), nil
	}

	return c.wrappedWriter.Write(content)
}

func (c *sessionCommitter) Flush() {
	if c.commit() {
		c.wrappedWriter.Flush()
	}
}

func (c *sessionCommitter) Hijack() (net.Conn, *bufio.ReadWriter, error) {
	if !c.commit() {
		return nil, nil, errSessionNotSaved
	}

	return c.wrappedWriter.Hijack()
}

func (c *sessionCommitter) commit() bool {
	if !c.committed {
		c.committed = true

		err := c.keeper.commit(c.request.Context(), c.ResponseWriter, c.state)
		if err != nil {
			c.failed = true
			c.renderer.Error(c.ResponseWriter, c.request, err)
		}
	}

	return !c.failed
}

package middleware

import (
	"context"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

type sessionPeeker interface {
	Peek(ctx context.Context, id string) ([]byte, bool, error)
}

type SessionReader struct {
	states        sessionStateReader
	authenticator *Authenticator
}

func NewSessionReader(
	store sessionPeeker,
	clock clock,
	settings SessionSettings,
	authenticator *Authenticator,
) *SessionReader {
	return &SessionReader{
		states:        sessionStateReader{cookie: sessionCookie(settings), clock: clock, fetch: store.Peek},
		authenticator: authenticator,
	}
}

func (s *SessionReader) SignedInUser(r *http.Request) (domain.User, bool, error) {
	state, err := s.states.read(r)
	if err != nil {
		var guest domain.User

		return guest, false, err
	}

	return s.authenticator.authenticate(withSessionState(r.Context(), state))
}

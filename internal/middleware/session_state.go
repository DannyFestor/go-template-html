package middleware

import "context"

type sessionStateKey struct{}

type sessionState struct {
	token     sessionToken
	hadCookie bool
	retired   []sessionToken
	data      sessionData
	changed   bool
	renew     bool
}

func guestSessionState() *sessionState {
	var empty sessionData

	return &sessionState{token: "", hadCookie: false, retired: nil, data: empty, changed: false, renew: false}
}

// staleSessionState starts out changed, so the commit deletes what is left of the session and clears its cookie.
func staleSessionState(retired []sessionToken) *sessionState {
	var empty sessionData

	return &sessionState{token: "", hadCookie: true, retired: retired, data: empty, changed: true, renew: false}
}

func loadedSessionState(token sessionToken, data sessionData) *sessionState {
	return &sessionState{token: token, hadCookie: true, retired: nil, data: data, changed: false, renew: false}
}

func withSessionState(ctx context.Context, state *sessionState) context.Context {
	return context.WithValue(ctx, sessionStateKey{}, state)
}

func sessionStateFrom(ctx context.Context) (*sessionState, bool) {
	state, ok := ctx.Value(sessionStateKey{}).(*sessionState)

	return state, ok
}

func (s *sessionState) signIn(userID, credentialFingerprint string) {
	s.data.UserID = userID
	s.data.CredentialFingerprint = credentialFingerprint
	s.renew = true
	s.changed = true
}

func (s *sessionState) clear() {
	var empty sessionData

	s.retireToken()
	s.data = empty
	s.changed = true
}

func (s *sessionState) setFlash(message string) {
	s.data.Flash = message
	s.changed = true
}

func (s *sessionState) takeFlash() string {
	message := s.data.Flash
	if message != "" {
		s.data.Flash = ""
		s.changed = true
	}

	return message
}

func (s *sessionState) retireToken() {
	if s.token != "" {
		s.retired = append(s.retired, s.token)
		s.token = ""
	}
}

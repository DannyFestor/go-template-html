package middleware

import (
	"context"
	"errors"
)

var errNoSession = errors.New("no session in the request context, LoadAndSaveSession must wrap the handler")

type Session struct{}

func (Session) SignIn(ctx context.Context, userID, credentialFingerprint string) {
	mustSessionState(ctx).signIn(userID, credentialFingerprint)
}

func (Session) SignOut(ctx context.Context) {
	mustSessionState(ctx).clear()
}

func (Session) CurrentUserID(ctx context.Context) (string, bool) {
	userID := currentSessionData(ctx).UserID

	return userID, userID != ""
}

func (Session) CredentialFingerprint(ctx context.Context) string {
	return currentSessionData(ctx).CredentialFingerprint
}

func (Session) SetFlash(ctx context.Context, message string) {
	mustSessionState(ctx).setFlash(message)
}

// TakeFlash tolerates a missing session, because error pages rendered outside the session middleware read it too.
func (Session) TakeFlash(ctx context.Context) string {
	state, ok := sessionStateFrom(ctx)
	if !ok {
		return ""
	}

	return state.takeFlash()
}

func currentSessionData(ctx context.Context) sessionData {
	var empty sessionData

	state, ok := sessionStateFrom(ctx)
	if !ok {
		return empty
	}

	return state.data
}

func mustSessionState(ctx context.Context) *sessionState {
	state, ok := sessionStateFrom(ctx)
	if !ok {
		panic(errNoSession)
	}

	return state
}

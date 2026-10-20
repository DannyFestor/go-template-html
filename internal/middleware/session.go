package middleware

import (
	"context"
	"errors"
	"time"
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
	state := mustSessionState(ctx)
	state.set(&state.data.Flash, message)
}

// TakeFlash tolerates a missing session, because error pages rendered outside the session middleware read it too.
func (Session) TakeFlash(ctx context.Context) string {
	state, ok := sessionStateFrom(ctx)
	if !ok {
		return ""
	}

	return state.take(&state.data.Flash)
}

func (Session) SetIntendedURL(ctx context.Context, url string) {
	state := mustSessionState(ctx)
	state.set(&state.data.IntendedURL, url)
}

func (Session) TakeIntendedURL(ctx context.Context) string {
	state := mustSessionState(ctx)

	return state.take(&state.data.IntendedURL)
}

func (Session) ConfirmPassword(ctx context.Context, at time.Time) {
	mustSessionState(ctx).confirmPassword(at)
}

func (Session) PasswordConfirmedAt(ctx context.Context) time.Time {
	return currentSessionData(ctx).PasswordConfirmedAt
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

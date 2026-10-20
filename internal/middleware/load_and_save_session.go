package middleware

import (
	"net/http"
	"time"
)

const sessionCookieName = "session"

type SessionSettings struct {
	Secure           bool
	AbsoluteLifetime time.Duration
}

func LoadAndSaveSession(
	store sessionStore,
	clock clock,
	settings SessionSettings,
	renderer errorRenderer,
) func(http.Handler) http.Handler {
	keeper := &sessionKeeper{
		store:            store,
		clock:            clock,
		cookie:           hostOnlyCookie{baseName: sessionCookieName, secure: settings.Secure},
		absoluteLifetime: settings.AbsoluteLifetime,
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			state, err := keeper.load(r)
			if err != nil {
				renderer.Error(w, r, err)

				return
			}

			r = r.WithContext(withSessionState(r.Context(), state))
			committer := &sessionCommitter{
				ResponseWriter: w,
				request:        r,
				keeper:         keeper,
				state:          state,
				renderer:       renderer,
				committed:      false,
				failed:         false,
			}

			next.ServeHTTP(committer, r)
			committer.commit() //nolint:contextcheck // The committer uses the context of the request it holds.
		})
	}
}

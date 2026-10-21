package middleware

import "net/http"

func NoReferrer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(referrerPolicyHeader, "no-referrer")

		next.ServeHTTP(w, r)
	})
}

package middleware

import (
	"net/http"
	"net/url"

	"github.com/DannyFestor/go-template-html/internal/web"
)

type signatureVerifier interface {
	Verify(target *url.URL) error
}

func RequireValidSignature(verifier signatureVerifier, renderer errorRenderer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if verifier.Verify(r.URL) != nil {
				renderer.Error(w, r, web.ErrForbidden)

				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

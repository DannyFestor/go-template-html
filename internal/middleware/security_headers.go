package middleware

import "net/http"

const (
	contentSecurityPolicy = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; " +
		"font-src 'self'; connect-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'; " +
		"object-src 'none'"
	permissionsPolicy       = "camera=(), microphone=(), geolocation=(), payment=(), usb=()"
	strictTransportSecurity = "max-age=31536000"
	referrerPolicyHeader    = "Referrer-Policy"
)

func SecurityHeaders(strictTransport bool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			setSecurityHeaders(w.Header(), strictTransport)

			next.ServeHTTP(w, r)
		})
	}
}

func setSecurityHeaders(header http.Header, strictTransport bool) {
	header.Set("Content-Security-Policy", contentSecurityPolicy)
	header.Set("X-Content-Type-Options", "nosniff")
	header.Set(referrerPolicyHeader, "strict-origin-when-cross-origin")
	header.Set("Cross-Origin-Opener-Policy", "same-origin")
	header.Set("Permissions-Policy", permissionsPolicy)

	if strictTransport {
		header.Set("Strict-Transport-Security", strictTransportSecurity)
	}
}

//go:build feature

package feature_test

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/route"
)

const strictContentSecurityPolicy = "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; " +
	"font-src 'self'; connect-src 'self'; form-action 'self'; frame-ancestors 'none'; base-uri 'none'; " +
	"object-src 'none'"

func TestEveryResponseCarriesTheStrictContentSecurityPolicy(t *testing.T) {
	t.Parallel()

	running := startApp(t)

	requests := map[string]outgoing{
		"page":         {method: http.MethodGet, path: route.Home, headers: nil, body: nil},
		"not found":    {method: http.MethodGet, path: "/no-such-page", headers: nil, body: nil},
		"server error": {method: http.MethodGet, path: probeFailure, headers: nil, body: nil},
		"asset":        {method: http.MethodGet, path: "/" + builtAppEntry(t).File, headers: nil, body: nil},
		"liveness":     {method: http.MethodGet, path: route.Healthz, headers: nil, body: nil},
		"readiness":    {method: http.MethodGet, path: route.Readyz, headers: nil, body: nil},
		"cross-origin": postWith(map[string]string{"Sec-Fetch-Site": "cross-site"}),
	}

	for name, request := range requests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			fetched := running.send(t, request)

			assert.Equal(t, strictContentSecurityPolicy, fetched.header.Get("Content-Security-Policy"))
			assert.Equal(t, "nosniff", fetched.header.Get("X-Content-Type-Options"))
			assert.Empty(t, fetched.header.Get("Strict-Transport-Security"), "HSTS on a plain-http app")
		})
	}
}

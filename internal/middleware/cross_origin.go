package middleware

import (
	"fmt"
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/web"
)

func CrossOriginProtection(trustedOrigins []string, renderer errorRenderer) (func(http.Handler) http.Handler, error) {
	protection := http.NewCrossOriginProtection()

	for _, origin := range trustedOrigins {
		err := protection.AddTrustedOrigin(origin)
		if err != nil {
			return nil, fmt.Errorf("trust the origin %q: %w", origin, err)
		}
	}

	protection.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		renderer.Error(w, r, web.ErrForbidden)
	}))

	return protection.Handler, nil
}

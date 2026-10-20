package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/DannyFestor/go-template-html/internal/route"
)

type intendedURL struct {
	session Session
}

func (i intendedURL) store(r *http.Request) {
	if r.Method == http.MethodGet {
		i.session.SetIntendedURL(r.Context(), r.URL.RequestURI())
	}
}

func (i intendedURL) take(ctx context.Context) string {
	intended := i.session.TakeIntendedURL(ctx)
	if !isLocalPath(intended) {
		return route.Dashboard
	}

	return intended
}

func isLocalPath(path string) bool {
	return strings.HasPrefix(path, "/") && !strings.HasPrefix(path, "//")
}

package health

import (
	"context"
	"io"
	"net/http"
)

type pinger interface {
	Ping(ctx context.Context) error
}

type Readiness struct {
	dependency pinger
}

func NewReadiness(dependency pinger) *Readiness {
	return &Readiness{dependency: dependency}
}

func (r *Readiness) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	if r.dependency.Ping(req.Context()) != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		_, _ = io.WriteString(w, "unavailable\n")

		return
	}

	_, _ = io.WriteString(w, "ok\n")
}

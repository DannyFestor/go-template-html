package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/DannyFestor/go-template-html/internal/route"
)

func AccessLog(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			started := time.Now()
			recorder := &statusRecorder{ResponseWriter: w, status: 0, bytes: 0}

			next.ServeHTTP(recorder, r)

			logger.LogAttrs(r.Context(), accessLogLevel(r), "request",
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
				slog.Int("status", recorder.statusOrOK()),
				slog.Duration("duration", time.Since(started)),
				slog.Int("bytes", recorder.bytes),
			)
		})
	}
}

func accessLogLevel(r *http.Request) slog.Level {
	if r.URL.Path == route.Healthz || r.URL.Path == route.Readyz {
		return slog.LevelDebug
	}

	return slog.LevelInfo
}

type statusRecorder struct {
	wrappedWriter

	status int
	bytes  int
}

func (r *statusRecorder) WriteHeader(status int) {
	if r.status == 0 {
		r.status = status
	}

	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Write(content []byte) (int, error) {
	written, err := r.wrappedWriter.Write(content)
	r.bytes += written

	return written, err
}

func (r *statusRecorder) statusOrOK() int {
	if r.status == 0 {
		return http.StatusOK
	}

	return r.status
}

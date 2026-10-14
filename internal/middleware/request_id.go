package middleware

import (
	"context"
	"net/http"
	"uuid"
)

const requestIDHeader = "X-Request-ID"

type requestIDKey struct{}

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.NewV4().String()
		w.Header().Set(requestIDHeader, requestID)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), requestIDKey{}, requestID)))
	})
}

func requestIDFrom(ctx context.Context) (string, bool) {
	requestID, ok := ctx.Value(requestIDKey{}).(string)

	return requestID, ok
}

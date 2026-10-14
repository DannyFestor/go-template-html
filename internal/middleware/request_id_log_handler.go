package middleware

import (
	"context"
	"fmt"
	"log/slog"
)

type RequestIDLogHandler struct {
	inner slog.Handler
}

func NewRequestIDLogHandler(inner slog.Handler) *RequestIDLogHandler {
	return &RequestIDLogHandler{inner: inner}
}

func (h *RequestIDLogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.inner.Enabled(ctx, level)
}

func (h *RequestIDLogHandler) Handle(ctx context.Context, record slog.Record) error {
	if requestID, ok := requestIDFrom(ctx); ok {
		record.AddAttrs(slog.String("request_id", requestID))
	}

	err := h.inner.Handle(ctx, record)
	if err != nil {
		return fmt.Errorf("handle the log record: %w", err)
	}

	return nil
}

func (h *RequestIDLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return NewRequestIDLogHandler(h.inner.WithAttrs(attrs))
}

func (h *RequestIDLogHandler) WithGroup(name string) slog.Handler {
	return NewRequestIDLogHandler(h.inner.WithGroup(name))
}

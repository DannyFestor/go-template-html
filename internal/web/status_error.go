package web

import (
	"net/http"
	"time"
)

type StatusError struct {
	Status     int
	RetryAfter time.Duration
}

var (
	ErrForbidden        = &StatusError{Status: http.StatusForbidden, RetryAfter: 0}
	ErrNotFound         = &StatusError{Status: http.StatusNotFound, RetryAfter: 0}
	ErrMethodNotAllowed = &StatusError{Status: http.StatusMethodNotAllowed, RetryAfter: 0}
)

func ErrTooManyRequests(retryAfter time.Duration) *StatusError {
	return &StatusError{Status: http.StatusTooManyRequests, RetryAfter: retryAfter}
}

func (e *StatusError) Error() string {
	return http.StatusText(e.Status)
}

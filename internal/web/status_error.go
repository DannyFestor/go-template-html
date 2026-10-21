package web

import (
	"errors"
	"net/http"
	"time"
)

type StatusError struct {
	Status     int
	RetryAfter time.Duration
	Cause      error
}

var (
	ErrForbidden        = &StatusError{Status: http.StatusForbidden, RetryAfter: 0, Cause: nil}
	ErrNotFound         = &StatusError{Status: http.StatusNotFound, RetryAfter: 0, Cause: nil}
	ErrMethodNotAllowed = &StatusError{Status: http.StatusMethodNotAllowed, RetryAfter: 0, Cause: nil}
)

func ErrTooManyRequests(retryAfter time.Duration) *StatusError {
	return &StatusError{Status: http.StatusTooManyRequests, RetryAfter: retryAfter, Cause: nil}
}

func ErrServiceUnavailable(cause error) *StatusError {
	return &StatusError{Status: http.StatusServiceUnavailable, RetryAfter: 0, Cause: cause}
}

func IsTooManyRequests(err error) bool {
	refusal, ok := errors.AsType[*StatusError](err)

	return ok && refusal.Status == http.StatusTooManyRequests
}

func (e *StatusError) Error() string {
	if e.hasCause() {
		return http.StatusText(e.Status) + ": " + e.Cause.Error()
	}

	return http.StatusText(e.Status)
}

func (e *StatusError) Unwrap() error {
	return e.Cause
}

func (e *StatusError) hasCause() bool {
	return e.Cause != nil
}

package web

import (
	"errors"
	"net/http"
)

var (
	errContentTooLarge = &StatusError{Status: http.StatusRequestEntityTooLarge, RetryAfter: 0, Cause: nil}
	errInternal        = &StatusError{Status: http.StatusInternalServerError, RetryAfter: 0, Cause: nil}
)

func classify(err error) (*StatusError, bool) {
	if statusErr, ok := errors.AsType[*StatusError](err); ok {
		return statusErr, !statusErr.hasCause()
	}

	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		return errContentTooLarge, true
	}

	return errInternal, false
}

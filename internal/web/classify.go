package web

import (
	"errors"
	"net/http"
)

var (
	errContentTooLarge = &StatusError{Status: http.StatusRequestEntityTooLarge, RetryAfter: 0}
	errInternal        = &StatusError{Status: http.StatusInternalServerError, RetryAfter: 0}
)

func classify(err error) (*StatusError, bool) {
	if statusErr, ok := errors.AsType[*StatusError](err); ok {
		return statusErr, true
	}

	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		return errContentTooLarge, true
	}

	return errInternal, false
}

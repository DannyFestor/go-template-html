package middleware

import (
	"errors"
	"fmt"
	"net/http"
)

var errPanicked = errors.New("panic")

func Recover(renderer errorRenderer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer renderPanic(w, r, renderer)

			next.ServeHTTP(w, r)
		})
	}
}

func renderPanic(w http.ResponseWriter, r *http.Request, renderer errorRenderer) {
	recovered := recover()
	if recovered == nil {
		return
	}

	if isAbort(recovered) {
		panic(recovered)
	}

	renderer.Error(w, r, fmt.Errorf("%w: %v", errPanicked, recovered))
}

func isAbort(recovered any) bool {
	err, isError := recovered.(error)

	return isError && errors.Is(err, http.ErrAbortHandler)
}

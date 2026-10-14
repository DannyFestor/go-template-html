package middleware

import (
	"net/http"

	"github.com/DannyFestor/go-template-html/internal/web"
)

type errorRenderer interface {
	Error(w http.ResponseWriter, r *http.Request, err error)
}

func InterceptMuxErrors(renderer errorRenderer) func(http.Handler) http.Handler {
	return func(mux http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			interceptor := &muxErrorInterceptor{ResponseWriter: w, request: r, interceptedStatus: 0}

			mux.ServeHTTP(interceptor, r)

			if interceptor.intercepting() {
				renderer.Error(w, r, muxErrors()[interceptor.interceptedStatus])
			}
		})
	}
}

func muxErrors() map[int]*web.StatusError {
	return map[int]*web.StatusError{
		http.StatusNotFound:         web.ErrNotFound,
		http.StatusMethodNotAllowed: web.ErrMethodNotAllowed,
	}
}

type muxErrorInterceptor struct {
	wrappedWriter

	request           *http.Request
	interceptedStatus int
}

// WriteHeader reads the pattern only now, because the mux sets it on the request before calling any handler.
func (i *muxErrorInterceptor) WriteHeader(status int) {
	if _, isMuxError := muxErrors()[status]; isMuxError && i.request.Pattern == "" {
		i.interceptedStatus = status

		return
	}

	i.ResponseWriter.WriteHeader(status)
}

func (i *muxErrorInterceptor) Write(content []byte) (int, error) {
	if i.intercepting() {
		return len(content), nil
	}

	return i.wrappedWriter.Write(content)
}

func (i *muxErrorInterceptor) intercepting() bool {
	return i.interceptedStatus != 0
}

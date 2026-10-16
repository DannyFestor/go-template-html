package testapp

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func Serve(t *testing.T, handler http.Handler) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)

	return server
}

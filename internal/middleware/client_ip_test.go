package middleware_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/middleware"
)

type forwardedRequest struct {
	remoteAddr   string
	forwardedFor []string
}

func clientIPOf(t *testing.T, trustedProxies []string, request forwardedRequest) string {
	t.Helper()

	prefixes := make([]netip.Prefix, 0, len(trustedProxies))
	for _, proxy := range trustedProxies {
		prefixes = append(prefixes, netip.MustParsePrefix(proxy))
	}

	incoming := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	incoming.RemoteAddr = request.remoteAddr

	for _, value := range request.forwardedFor {
		incoming.Header.Add("X-Forwarded-For", value)
	}

	recorder := httptest.NewRecorder()
	middleware.ClientIP(prefixes)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, middleware.ClientIPFrom(r.Context()).String())
	})).ServeHTTP(recorder, incoming)

	return recorder.Body.String()
}

func TestClientIPIsTheRemoteAddressWithoutTrustedProxies(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "198.51.100.2", clientIPOf(t, nil, forwardedRequest{
		remoteAddr:   "198.51.100.2:51234",
		forwardedFor: []string{"203.0.113.7"},
	}))
}

func TestClientIPIgnoresForwardedForFromAnUntrustedPeer(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "198.51.100.2", clientIPOf(t, []string{"10.0.0.0/8"}, forwardedRequest{
		remoteAddr:   "198.51.100.2:51234",
		forwardedFor: []string{"203.0.113.7"},
	}))
}

func TestClientIPIsTheRightmostUntrustedForwardedForEntryBehindTrustedProxies(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "198.51.100.2", clientIPOf(t, []string{"10.0.0.0/8"}, forwardedRequest{
		remoteAddr:   "10.0.0.1:51234",
		forwardedFor: []string{"203.0.113.7, 198.51.100.2", "10.0.0.5"},
	}))
}

func TestClientIPIsTheLeftmostEntryWhenEveryHopIsTrusted(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "10.0.0.9", clientIPOf(t, []string{"10.0.0.0/8"}, forwardedRequest{
		remoteAddr:   "10.0.0.1:51234",
		forwardedFor: []string{"10.0.0.9, 10.0.0.5"},
	}))
}

func TestClientIPStopsAtAnUnreadableEntry(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "10.0.0.5", clientIPOf(t, []string{"10.0.0.0/8"}, forwardedRequest{
		remoteAddr:   "10.0.0.1:51234",
		forwardedFor: []string{"203.0.113.7, unknown, 10.0.0.5"},
	}))
}

func TestClientIPUnmapsIPv4MappedAddresses(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "198.51.100.2", clientIPOf(t, nil, forwardedRequest{
		remoteAddr:   "[::ffff:198.51.100.2]:51234",
		forwardedFor: nil,
	}))
}

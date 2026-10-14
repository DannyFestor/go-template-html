package bootstrap_test

import (
	"context"
	"net"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
)

type slowHandler struct {
	started chan struct{}
	release chan struct{}
}

func newSlowHandler() *slowHandler {
	return &slowHandler{started: make(chan struct{}), release: make(chan struct{})}
}

func (h *slowHandler) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	close(h.started)
	<-h.release
	w.WriteHeader(http.StatusOK)
}

type runningServer struct {
	addr   string
	cancel context.CancelFunc
	result chan error
	status chan int
}

func TestServeLetsInFlightRequestsFinishAfterCancellation(t *testing.T) {
	t.Parallel()

	handler := newSlowHandler()
	running := serveAndRequest(t, handler, time.Minute)

	<-handler.started
	running.cancel()
	require.Eventually(t, func() bool { return refusesConnections(t, running.addr) }, time.Second, time.Millisecond)
	close(handler.release)

	assert.Equal(t, http.StatusOK, <-running.status)
	assert.NoError(t, <-running.result)
}

func TestServeFailsWhenDrainingOutlastsTheTimeout(t *testing.T) {
	t.Parallel()

	handler := newSlowHandler()
	running := serveAndRequest(t, handler, 10*time.Millisecond)
	t.Cleanup(func() { close(handler.release) })

	<-handler.started
	running.cancel()

	assert.ErrorIs(t, <-running.result, context.DeadlineExceeded)
}

func TestServeClosesConnectionsThatOutlastTheTimeout(t *testing.T) {
	t.Parallel()

	handler := newSlowHandler()
	running := serveAndRequest(t, handler, 10*time.Millisecond)
	t.Cleanup(func() { close(handler.release) })

	<-handler.started
	running.cancel()
	<-running.result

	select {
	case status := <-running.status:
		assert.Zero(t, status)
	case <-time.After(time.Second):
		t.Fatal("the in-flight request is still open after the shutdown timeout")
	}
}

func TestServeFailsWhenTheListenerFails(t *testing.T) {
	t.Parallel()

	listener := listen(t)
	require.NoError(t, listener.Close())

	err := bootstrap.Serve(t.Context(), newServer(http.NotFoundHandler()), listener, time.Minute)

	assert.ErrorIs(t, err, net.ErrClosed)
}

func serveAndRequest(t *testing.T, handler http.Handler, shutdownTimeout time.Duration) runningServer {
	t.Helper()

	ctx, cancel := context.WithCancel(t.Context())
	listener := listen(t)
	running := runningServer{
		addr:   listener.Addr().String(),
		cancel: cancel,
		result: make(chan error, 1),
		status: make(chan int, 1),
	}

	go func() { running.result <- bootstrap.Serve(ctx, newServer(handler), listener, shutdownTimeout) }()
	go func() { running.status <- statusOf(t, "http://"+running.addr) }()

	return running
}

func listen(t *testing.T) net.Listener {
	t.Helper()

	var config net.ListenConfig

	listener, err := config.Listen(t.Context(), "tcp", "127.0.0.1:0")
	require.NoError(t, err)

	return listener
}

func newServer(handler http.Handler) *http.Server {
	return &http.Server{Handler: handler, ReadHeaderTimeout: time.Second} //nolint:exhaustruct_v5 // a bare test server
}

func refusesConnections(t *testing.T, addr string) bool {
	t.Helper()

	var dialer net.Dialer

	conn, err := dialer.DialContext(t.Context(), "tcp", addr)
	if err != nil {
		return true
	}

	_ = conn.Close()

	return false
}

func statusOf(t *testing.T, url string) int {
	t.Helper()

	request, err := http.NewRequestWithContext(context.WithoutCancel(t.Context()), http.MethodGet, url, nil)
	if err != nil {
		return 0
	}

	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return 0
	}

	_ = response.Body.Close()

	return response.StatusCode
}

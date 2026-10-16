//go:build feature

package feature_test

import (
	"bytes"
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/bootstrap"
)

func TestTheWorkerLogsItsVersionAtStartupAndStopsCleanlyOnShutdown(t *testing.T) {
	t.Parallel()

	config := liveConfig(t)
	logs := &lockedBuffer{mutex: sync.Mutex{}, buffer: bytes.Buffer{}}
	ctx, shutDown := context.WithCancel(t.Context())
	stopped := make(chan error, 1)

	go func() {
		stopped <- bootstrap.ServeWorker(ctx, config, bootstrap.NewLogger(logs, config.App.Env, config.Log))
	}()

	require.Eventually(t, func() bool {
		return strings.Contains(logs.String(), `"msg":"worker started","version":"dev"`)
	}, time.Second, time.Millisecond)

	shutDown()

	select {
	case err := <-stopped:
		require.NoError(t, err)
	case <-time.After(config.App.ShutdownTimeout + time.Second):
		t.Fatal("the worker did not stop")
	}
}

//go:build feature

package feature_test

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	"github.com/DannyFestor/go-template-html/internal/testsupport/testapp"
	"github.com/DannyFestor/go-template-html/internal/worker"
)

const (
	probeEnqueue    = "/probe/enqueue"
	probeJobMessage = "Probe job ran."
)

type probeJob struct {
	Message string `json:"message"`
}

func (probeJob) Kind() string {
	return "probe"
}

type probeJobHandler struct {
	mutex    sync.Mutex
	messages []string
}

func (h *probeJobHandler) Handle(_ context.Context, job probeJob) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.messages = append(h.messages, job.Message)

	return nil
}

func (h *probeJobHandler) handled() []string {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	return append([]string(nil), h.messages...)
}

func mountJobProbe(testApp *testapp.App) *probeJobHandler {
	handler := &probeJobHandler{mutex: sync.Mutex{}, messages: nil}
	worker.Register(testApp.Worker.Jobs, handler, worker.DefaultSettings())

	testApp.Mux.Handle("POST "+probeEnqueue, testApp.Responder.Handle(
		func(w http.ResponseWriter, r *http.Request) error {
			err := testApp.Jobs.Enqueue(r.Context(), probeJob{Message: probeJobMessage})
			if err != nil {
				return fmt.Errorf("enqueue the probe job: %w", err)
			}

			w.WriteHeader(http.StatusNoContent)

			return nil
		},
	))

	return handler
}

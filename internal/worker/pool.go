package worker

import (
	"sync"
	"time"
)

type pool struct {
	slots   chan struct{}
	running sync.WaitGroup
}

func newPool(size int) *pool {
	return &pool{slots: make(chan struct{}, size), running: sync.WaitGroup{}}
}

func (p *pool) tryAcquire() bool {
	select {
	case p.slots <- struct{}{}:
		return true
	default:
		return false
	}
}

func (p *pool) release() {
	<-p.slots
}

func (p *pool) start(perform func()) {
	p.running.Go(func() {
		defer p.release()

		perform()
	})
}

func (p *pool) waitFor(timeout time.Duration) bool {
	finished := make(chan struct{})

	go func() {
		p.running.Wait()
		close(finished)
	}()

	select {
	case <-finished:
		return true
	case <-time.After(timeout):
		return false
	}
}

func (p *pool) wait() {
	p.running.Wait()
}

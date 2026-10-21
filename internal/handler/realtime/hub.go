package realtime

import "sync"

type subscribers map[*subscription]struct{}

type Hub struct {
	mutex      sync.Mutex
	bufferSize int
	byChannel  map[string]subscribers
	active     subscribers
	closed     bool
	drained    chan struct{}
}

func NewHub(bufferSize int) *Hub {
	return &Hub{
		mutex:      sync.Mutex{},
		bufferSize: bufferSize,
		byChannel:  map[string]subscribers{},
		active:     subscribers{},
		closed:     false,
		drained:    make(chan struct{}),
	}
}

func (h *Hub) Deliver(channel string, payload []byte) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	for subscribed := range h.byChannel[channel] {
		if !subscribed.offer(payload) {
			h.end(subscribed, errFellBehind)
		}
	}
}

func (h *Hub) Close() {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.closed {
		return
	}

	h.closed = true

	for subscribed := range h.active {
		h.end(subscribed, errHubClosed)
	}

	h.signalIfDrained()
}

func (h *Hub) Drained() <-chan struct{} {
	return h.drained
}

func (h *Hub) subscribe(channels []string) *subscription {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	subscribed := newSubscription(channels, h.bufferSize)
	if h.closed {
		subscribed.end(errHubClosed)

		return subscribed
	}

	h.active[subscribed] = struct{}{}

	for _, channel := range channels {
		if h.byChannel[channel] == nil {
			h.byChannel[channel] = subscribers{}
		}

		h.byChannel[channel][subscribed] = struct{}{}
	}

	return subscribed
}

func (h *Hub) unsubscribe(subscribed *subscription) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, active := h.active[subscribed]; !active {
		return
	}

	h.detach(subscribed)
	delete(h.active, subscribed)
	h.signalIfDrained()
}

func (h *Hub) end(subscribed *subscription, reason error) {
	h.detach(subscribed)
	subscribed.end(reason)
}

func (h *Hub) detach(subscribed *subscription) {
	for _, channel := range subscribed.channels {
		delete(h.byChannel[channel], subscribed)

		if len(h.byChannel[channel]) == 0 {
			delete(h.byChannel, channel)
		}
	}
}

func (h *Hub) signalIfDrained() {
	if h.closed && len(h.active) == 0 {
		close(h.drained)
	}
}

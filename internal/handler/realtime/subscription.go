package realtime

import "errors"

var (
	errFellBehind = errors.New("the socket fell behind its broadcasts")
	errHubClosed  = errors.New("the server is shutting down")
)

type subscription struct {
	channels []string
	messages chan []byte
	ended    chan struct{}
	reason   error
}

func newSubscription(channels []string, bufferSize int) *subscription {
	return &subscription{
		channels: channels,
		messages: make(chan []byte, bufferSize),
		ended:    make(chan struct{}),
		reason:   nil,
	}
}

func (s *subscription) offer(payload []byte) bool {
	select {
	case s.messages <- payload:
		return true
	default:
		return false
	}
}

func (s *subscription) end(reason error) {
	if s.reason == nil {
		s.reason = reason
		close(s.ended)
	}
}

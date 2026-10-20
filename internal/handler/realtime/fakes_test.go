package realtime_test

import (
	"context"
	"errors"
	"net/http"
	"sync"
	"sync/atomic"

	"github.com/DannyFestor/go-template-html/internal/domain"
)

var (
	errValkeyDown     = errors.New("valkey is down")
	errUpgradeRefused = errors.New("the origin is not allowed")
)

type fakeConn struct {
	sent      chan []byte
	closes    chan int
	closed    chan struct{}
	release   chan struct{}
	pings     atomic.Int32
	closeOnce sync.Once
}

func newFakeConn() *fakeConn {
	return &fakeConn{
		sent:      make(chan []byte, 64),
		closes:    make(chan int, 1),
		closed:    make(chan struct{}),
		release:   nil,
		pings:     atomic.Int32{},
		closeOnce: sync.Once{},
	}
}

func newStalledConn() *fakeConn {
	conn := newFakeConn()
	conn.release = make(chan struct{})

	return conn
}

func (c *fakeConn) Send(_ context.Context, payload []byte) error {
	if c.release != nil {
		<-c.release
	}

	c.sent <- payload

	return nil
}

func (c *fakeConn) Ping(context.Context) error {
	c.pings.Add(1)

	return nil
}

func (c *fakeConn) Close(code int, _ string) error {
	c.closes <- code

	c.clientCloses()

	return nil
}

func (c *fakeConn) Closed() <-chan struct{} {
	return c.closed
}

func (c *fakeConn) clientCloses() {
	c.closeOnce.Do(func() { close(c.closed) })
}

type fakeUpgrader struct {
	conn     *fakeConn
	err      error
	accepted chan struct{}
}

func upgradingTo(conn *fakeConn) *fakeUpgrader {
	return &fakeUpgrader{conn: conn, err: nil, accepted: make(chan struct{})}
}

func (u *fakeUpgrader) Accept(http.ResponseWriter, *http.Request) (*fakeConn, error) {
	close(u.accepted)

	return u.conn, u.err
}

type fakeUsers struct {
	mutex    sync.Mutex
	user     domain.User
	signedIn bool
	err      error
	lookups  int
}

func signedInAs(user domain.User) *fakeUsers {
	return &fakeUsers{mutex: sync.Mutex{}, user: user, signedIn: true, err: nil, lookups: 0}
}

func guests() *fakeUsers {
	var nobody domain.User

	return &fakeUsers{mutex: sync.Mutex{}, user: nobody, signedIn: false, err: nil, lookups: 0}
}

func failingUsers() *fakeUsers {
	users := guests()
	users.err = errValkeyDown

	return users
}

func (u *fakeUsers) SignedInUser(*http.Request) (domain.User, bool, error) {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.lookups++

	return u.user, u.signedIn, u.err
}

func (u *fakeUsers) signOut() {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.signedIn = false
}

func (u *fakeUsers) fail() {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	u.err = errValkeyDown
}

func (u *fakeUsers) lookupCount() int {
	u.mutex.Lock()
	defer u.mutex.Unlock()

	return u.lookups
}

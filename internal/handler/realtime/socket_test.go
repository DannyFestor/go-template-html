package realtime_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/domain"
	realtimehandler "github.com/DannyFestor/go-template-html/internal/handler/realtime"
	"github.com/DannyFestor/go-template-html/internal/web"
)

const (
	announcements   = "announcements"
	adaChannel      = "users.ada"
	maxChannels     = 3
	bufferSize      = 2
	pingInterval    = 10 * time.Millisecond
	waitForTheLoop  = 5 * time.Second
	closeGoingAway  = 1001
	closePolicy     = 1008
	closeInternal   = 1011
	shortObservable = 50 * time.Millisecond
)

type serving struct {
	hub      *realtimehandler.Hub
	conn     *fakeConn
	upgrader *fakeUpgrader
	result   chan error
}

func ada() domain.User {
	return domain.User{
		ID:              "ada",
		Name:            "Ada Lovelace",
		Email:           "ada@example.com",
		PasswordHash:    "",
		CreatedAt:       time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC),
		EmailVerifiedAt: time.Date(2026, 10, 21, 9, 5, 0, 0, time.UTC),
	}
}

func channels() *realtimehandler.Channels {
	registry := realtimehandler.NewChannels()
	registry.Public(announcements)
	registry.Private("users.{id}", func(user domain.User, params map[string]string) bool {
		return params["id"] == string(user.ID)
	})

	return registry
}

func settings() realtimehandler.Settings {
	return realtimehandler.Settings{PingInterval: pingInterval, WriteTimeout: time.Second, MaxChannels: maxChannels}
}

func serve(t *testing.T, users *fakeUsers, conn *fakeConn, query string) *serving {
	t.Helper()

	t.Cleanup(conn.clientCloses)

	return serveThrough(t, users, upgradingTo(conn), query)
}

func serveThrough(t *testing.T, users *fakeUsers, upgrader *fakeUpgrader, query string) *serving {
	t.Helper()

	hub := realtimehandler.NewHub(bufferSize)
	socket := realtimehandler.NewSocket[*fakeConn](channels(), hub, users, upgrader, settings())

	result := make(chan error, 1)

	go func() {
		result <- socket.Serve(httptest.NewRecorder(), httptest.NewRequestWithContext(
			t.Context(), http.MethodGet, "/ws?channels="+query, nil,
		))
	}()

	return &serving{hub: hub, conn: upgrader.conn, upgrader: upgrader, result: result}
}

func serveAccepted(t *testing.T, users *fakeUsers, conn *fakeConn, query string) *serving {
	t.Helper()

	running := serve(t, users, conn, query)
	running.awaitAccepted(t)

	return running
}

func (s *serving) awaitAccepted(t *testing.T) {
	t.Helper()

	select {
	case <-s.upgrader.accepted:
	case err := <-s.result:
		require.FailNow(t, "the socket returned before the upgrade", err)
	case <-time.After(waitForTheLoop):
		require.FailNow(t, "the socket never upgraded")
	}
}

func (s *serving) refusal(t *testing.T) error {
	t.Helper()

	select {
	case err := <-s.result:
		return err
	case <-time.After(waitForTheLoop):
		require.FailNow(t, "the socket never returned")

		return nil
	}
}

func (s *serving) nextSent(t *testing.T) string {
	t.Helper()

	select {
	case payload := <-s.conn.sent:
		return string(payload)
	case <-time.After(waitForTheLoop):
		require.FailNow(t, "nothing was sent")

		return ""
	}
}

func (s *serving) closeCode(t *testing.T) int {
	t.Helper()

	select {
	case code := <-s.conn.closes:
		return code
	case <-time.After(waitForTheLoop):
		require.FailNow(t, "the socket was never closed")

		return 0
	}
}

func (s *serving) awaitDrained(t *testing.T, timeout time.Duration) {
	t.Helper()

	select {
	case <-s.hub.Drained():
	case <-time.After(timeout):
		require.FailNow(t, "a socket still holds its subscription")
	}
}

func TestABroadcastOnASubscribedChannelIsSentToTheSocket(t *testing.T) {
	t.Parallel()

	running := serveAccepted(t, guests(), newFakeConn(), announcements)

	running.hub.Deliver(announcements, []byte("<div>hi</div>"))

	assert.Equal(t, "<div>hi</div>", running.nextSent(t))
}

func TestABroadcastOnAnotherChannelIsNotSent(t *testing.T) {
	t.Parallel()

	running := serveAccepted(t, guests(), newFakeConn(), announcements)

	running.hub.Deliver(adaChannel, []byte("private"))
	running.hub.Deliver(announcements, []byte("public"))

	assert.Equal(t, "public", running.nextSent(t))
}

func TestARepeatedChannelDeliversEachBroadcastOnce(t *testing.T) {
	t.Parallel()

	running := serveAccepted(t, guests(), newFakeConn(), announcements+","+announcements)

	running.hub.Deliver(announcements, []byte("first"))
	running.hub.Deliver(announcements, []byte("second"))

	assert.Equal(t, "first", running.nextSent(t))
	assert.Equal(t, "second", running.nextSent(t))
}

func TestASocketNeedsBetweenOneAndTheMaximumNumberOfChannels(t *testing.T) {
	t.Parallel()

	for name, query := range map[string]string{
		"none":         "",
		"an empty one": announcements + ",",
		"too many":     "a,b,c,d",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			err := serve(t, guests(), newFakeConn(), query).refusal(t)

			assert.ErrorIs(t, err, web.ErrBadRequest)
		})
	}
}

func TestAChannelNoRuleMatchesIsForbidden(t *testing.T) {
	t.Parallel()

	for _, query := range []string{"secrets", "users", "users.ada.extra", "users."} {
		t.Run(query, func(t *testing.T) {
			t.Parallel()

			err := serve(t, signedInAs(ada()), newFakeConn(), query).refusal(t)

			assert.ErrorIs(t, err, web.ErrForbidden)
		})
	}
}

func TestAGuestIsForbiddenAPrivateChannel(t *testing.T) {
	t.Parallel()

	err := serve(t, guests(), newFakeConn(), announcements+","+adaChannel).refusal(t)

	assert.ErrorIs(t, err, web.ErrForbidden)
}

func TestAUserIsForbiddenAnotherUsersChannel(t *testing.T) {
	t.Parallel()

	err := serve(t, signedInAs(ada()), newFakeConn(), "users.grace").refusal(t)

	assert.ErrorIs(t, err, web.ErrForbidden)
}

func TestAnUnverifiedUserIsForbiddenTheirOwnChannel(t *testing.T) {
	t.Parallel()

	unverified := ada()
	unverified.EmailVerifiedAt = time.Time{}

	err := serve(t, signedInAs(unverified), newFakeConn(), adaChannel).refusal(t)

	assert.ErrorIs(t, err, web.ErrForbidden)
}

func TestAUserReceivesBroadcastsOnTheirOwnChannel(t *testing.T) {
	t.Parallel()

	running := serveAccepted(t, signedInAs(ada()), newFakeConn(), adaChannel)

	running.hub.Deliver(adaChannel, []byte("notice"))

	assert.Equal(t, "notice", running.nextSent(t))
}

func TestAFailingSessionLookupAtTheUpgradeIsReturned(t *testing.T) {
	t.Parallel()

	err := serve(t, failingUsers(), newFakeConn(), adaChannel).refusal(t)

	assert.ErrorIs(t, err, errValkeyDown)
}

func TestAPublicOnlySocketNeverLooksUpTheSession(t *testing.T) {
	t.Parallel()

	users := failingUsers()
	running := serveAccepted(t, users, newFakeConn(), announcements)

	time.Sleep(5 * pingInterval)

	assert.Zero(t, users.lookupCount())
	assert.Positive(t, running.conn.pings.Load())
}

func TestASocketThatFallsBehindIsClosedAsAPolicyViolation(t *testing.T) {
	t.Parallel()

	conn := newStalledConn()
	running := serveAccepted(t, guests(), conn, announcements)

	for range bufferSize + 2 {
		running.hub.Deliver(announcements, []byte("flood"))
	}

	close(conn.release)

	assert.Equal(t, closePolicy, running.closeCode(t))
}

func TestClosingTheHubClosesSocketsAsGoingAwayAndDrainsOnceTheyReturn(t *testing.T) {
	t.Parallel()

	running := serveAccepted(t, guests(), newFakeConn(), announcements)

	running.hub.Close()

	assert.Equal(t, closeGoingAway, running.closeCode(t))
	require.NoError(t, running.refusal(t))
	running.awaitDrained(t, waitForTheLoop)
}

func TestASocketWhoseSessionEndedIsClosedAsAPolicyViolationAtTheNextPing(t *testing.T) {
	t.Parallel()

	users := signedInAs(ada())
	running := serveAccepted(t, users, newFakeConn(), adaChannel)

	users.signOut()

	assert.Equal(t, closePolicy, running.closeCode(t))
}

func TestASocketWhoseSessionCannotBeCheckedIsClosedAsAnInternalError(t *testing.T) {
	t.Parallel()

	users := signedInAs(ada())
	running := serveAccepted(t, users, newFakeConn(), adaChannel)

	users.fail()

	assert.Equal(t, closeInternal, running.closeCode(t))
}

func TestASocketWithALiveSessionIsPingedAndStaysOpen(t *testing.T) {
	t.Parallel()

	running := serveAccepted(t, signedInAs(ada()), newFakeConn(), adaChannel)

	time.Sleep(5 * pingInterval)

	assert.Positive(t, running.conn.pings.Load())
	assert.Empty(t, running.conn.closes)
}

func TestTheSocketReturnsOnceTheClientCloses(t *testing.T) {
	t.Parallel()

	running := serveAccepted(t, guests(), newFakeConn(), announcements)

	running.conn.clientCloses()

	require.NoError(t, running.refusal(t))
	running.hub.Close()
	running.awaitDrained(t, shortObservable)
}

func TestAFailedUpgradeLeavesTheAnswerToTheUpgrader(t *testing.T) {
	t.Parallel()

	refused := &fakeUpgrader{conn: nil, err: errUpgradeRefused, accepted: make(chan struct{})}

	assert.NoError(t, serveThrough(t, guests(), refused, announcements).refusal(t))
}

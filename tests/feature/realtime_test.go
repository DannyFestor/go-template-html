//go:build feature

package feature_test

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/route"
	"github.com/DannyFestor/go-template-html/internal/testsupport"
)

const (
	announcementsChannel = "announcements"
	announcementText     = "Maintenance starts at noon."
	noticeText           = "Your password was changed on another device."
	maxChannelsPerSocket = 20
	beyondServerTimeouts = 1500 * time.Millisecond
)

func userChannel(user domain.User) string {
	return "users." + string(user.ID)
}

func (a *app) announce(t *testing.T) func() {
	t.Helper()

	return func() { a.Announcements.Broadcast(t.Context(), domain.Announcement{Text: announcementText}) }
}

func (a *app) notify(t *testing.T, user domain.User) func() {
	t.Helper()

	return func() {
		a.UserNotices.Broadcast(t.Context(), domain.UserNotice{UserID: user.ID, Text: noticeText})
	}
}

func (a *app) signInAs(t *testing.T, email string) domain.User {
	t.Helper()

	user := a.createUser(t, email)
	a.loginAs(t, email)

	return user
}

func (a *app) createUser(t *testing.T, email string) domain.User {
	t.Helper()

	return testsupport.Must(a.factory().User().WithEmail(email).Create(t.Context()))(t)
}

func TestAnAnnouncementReachesASocketOnTheAnnouncementsChannel(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	opened := running.mustDial(t, announcementsChannel)

	fragment := opened.receiveWhileBroadcasting(t, running.announce(t))

	banner := fragment.Find("#banner-slot")
	require.Equal(t, 1, banner.Length())
	assert.Equal(t, "true", banner.AttrOr("hx-swap-oob", ""))
	assert.Equal(t, announcementText, strings.TrimSpace(banner.Text()))
}

func TestAUserNoticeReachesTheSocketOnItsUsersPrivateChannel(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	ada := running.signInAs(t, adaEmail)
	opened := running.mustDial(t, userChannel(ada))

	fragment := opened.receiveWhileBroadcasting(t, running.notify(t, ada))

	toast := fragment.Find(`[hx-swap-oob="beforeend:#toast-stack"]`)
	require.Equal(t, 1, toast.Length())
	assert.Equal(t, noticeText, strings.TrimSpace(toast.Text()))
}

func TestAGuestCannotSubscribeToAPrivateUserChannel(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	ada := running.createUser(t, adaEmail)

	_, status := running.dial(t, userChannel(ada))

	assert.Equal(t, http.StatusForbidden, status)
}

func TestAUserCannotSubscribeToAnotherUsersPrivateChannel(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	grace := running.createUser(t, "grace@example.com")
	running.signInAs(t, adaEmail)

	_, status := running.dial(t, announcementsChannel, userChannel(grace))

	assert.Equal(t, http.StatusForbidden, status)
}

func TestAnUnverifiedUserCannotSubscribeToTheirOwnPrivateChannel(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	ada := running.createUnverifiedAda(t)
	running.loginAs(t, adaEmail)

	_, status := running.dial(t, userChannel(ada))

	assert.Equal(t, http.StatusForbidden, status)
}

func TestAnUnknownChannelIsRefused(t *testing.T) {
	t.Parallel()

	_, status := startApp(t).dial(t, "secrets")

	assert.Equal(t, http.StatusForbidden, status)
}

func TestASocketNeedsBetweenOneAndTwentyChannels(t *testing.T) {
	t.Parallel()

	tooMany := make([]string, 0, maxChannelsPerSocket+1)
	for range maxChannelsPerSocket + 1 {
		tooMany = append(tooMany, announcementsChannel+"x")
	}

	for name, channels := range map[string][]string{"none": nil, "too many": tooMany} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, status := startApp(t).dial(t, channels...)

			assert.Equal(t, http.StatusBadRequest, status)
		})
	}
}

func TestACrossOriginUpgradeIsRefused(t *testing.T) {
	t.Parallel()

	_, status := startApp(t).dialFrom(t, http.Header{"Origin": {"https://evil.example"}}, announcementsChannel)

	assert.Equal(t, http.StatusForbidden, status)
}

func TestAnInboundMessageClosesTheSocketAsAPolicyViolation(t *testing.T) {
	t.Parallel()

	opened := startApp(t).mustDial(t, announcementsChannel)

	require.NoError(t, opened.conn.Write(t.Context(), websocket.MessageText, []byte(`{"hello":"server"}`)))

	assert.Equal(t, websocket.StatusPolicyViolation, opened.closeStatus(t))
}

func TestASocketOutlivesTheServersReadAndWriteTimeouts(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	opened := running.mustDial(t, announcementsChannel)

	time.Sleep(beyondServerTimeouts)

	fragment := opened.receiveWhileBroadcasting(t, running.announce(t))

	assert.Equal(t, 1, fragment.Find("#banner-slot").Length())
}

func TestThePublicLayoutConnectsToTheAnnouncementsChannel(t *testing.T) {
	t.Parallel()

	document := startApp(t).page(t, route.Home)

	connect := document.Find(`[hx-ws\:connect]`)
	require.Equal(t, 1, connect.Length())
	assert.Equal(t, route.Realtime+"?channels="+announcementsChannel, connect.AttrOr("hx-ws:connect", ""))
	assert.Equal(t, "none", connect.AttrOr("hx-swap", ""))
	assert.Zero(t, document.Find(`main [hx-ws\:connect]`).Length(), "the socket must survive boosted navigation")
	assert.Equal(t, 1, document.Find("#banner-slot").Length())
}

func TestTheAppLayoutConnectsToTheAnnouncementsAndTheUsersOwnChannel(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	ada := running.signInAs(t, adaEmail)

	document := running.page(t, route.Dashboard)

	connect := document.Find(`[hx-ws\:connect]`)
	require.Equal(t, 1, connect.Length())
	assert.Equal(t,
		route.Realtime+"?channels="+announcementsChannel+","+userChannel(ada),
		connect.AttrOr("hx-ws:connect", ""),
	)
	assert.Zero(t, document.Find(`main [hx-ws\:connect]`).Length(), "the socket must survive boosted navigation")
	assert.Equal(t, 1, document.Find("#banner-slot").Length())
	assert.Equal(t, 1, document.Find("#toast-stack").Length())
}

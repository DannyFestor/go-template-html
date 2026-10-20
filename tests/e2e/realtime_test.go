//go:build e2e

package e2e_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/domain"
	"github.com/DannyFestor/go-template-html/internal/route"
)

const (
	rebroadcastInterval = 100 * time.Millisecond
	announcementText    = "Maintenance starts at noon."
	noticeText          = "Your password was changed on another device."
)

func keepBroadcasting(t *testing.T, broadcast func()) {
	t.Helper()

	stop := make(chan struct{})
	stopped := make(chan struct{})

	go func() {
		defer close(stopped)

		ticker := time.NewTicker(rebroadcastInterval)
		defer ticker.Stop()

		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				broadcast()
			}
		}
	}()

	t.Cleanup(func() {
		close(stop)
		<-stopped
	})
}

func TestAnAnnouncementAppearsAsABannerAfterBoostedNavigation(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	page := running.open(t, "/")
	require.NoError(t, page.Locator(`header a[href="/login"]`).Click())
	require.NoError(t, expect.Page(page.Page).ToHaveURL(regexp.MustCompile(`/login$`)))

	keepBroadcasting(t, func() {
		running.Announcements.Broadcast(t.Context(), domain.Announcement{Text: announcementText})
	})

	require.NoError(t, expect.Locator(page.Locator("#banner-slot")).ToHaveText(announcementText))
	assert.Empty(t, page.errors.list())
}

func TestAUserNoticeAppearsAsAToastInTheAppLayout(t *testing.T) {
	t.Parallel()

	running := startApp(t)
	ada := running.createAda(t)
	page := running.open(t, route.Login)
	page.logIn(t, ada)

	keepBroadcasting(t, func() {
		running.UserNotices.Broadcast(t.Context(), domain.UserNotice{UserID: ada.ID, Text: noticeText})
	})

	require.NoError(t, expect.Locator(page.Locator("#toast-stack [role=status]").First()).ToHaveText(noticeText))
	assert.Empty(t, page.errors.list())
}

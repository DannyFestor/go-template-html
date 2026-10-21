package realtime

import (
	"strings"

	"github.com/DannyFestor/go-template-html/internal/route"
)

const (
	Announcements      = "announcements"
	UserChannelParam   = "id"
	userChannelPrefix  = "users."
	UserChannelPattern = userChannelPrefix + "{" + UserChannelParam + "}"
	channelSeparator   = ","
)

func UserChannel(userID string) string {
	return userChannelPrefix + userID
}

func connectURL(channels []string) string {
	return route.Realtime + "?channels=" + strings.Join(channels, channelSeparator)
}

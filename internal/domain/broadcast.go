package domain

type Announcement struct {
	Text string
}

type UserNotice struct {
	UserID UserID
	Text   string
}

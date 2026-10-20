package domain

type MailMessage struct {
	To      string
	Subject string
	HTML    string
	Text    string
}

type VerifyEmailMail struct {
	To  string
	URL string
}

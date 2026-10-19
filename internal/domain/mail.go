package domain

type MailMessage struct {
	To      string
	Subject string
	HTML    string
	Text    string
}

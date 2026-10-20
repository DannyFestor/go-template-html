package domain

type PasswordResetToken string

type ResetPasswordMail struct {
	To  string
	URL string
}

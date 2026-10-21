package mail

import (
	"text/template"

	"github.com/a-h/templ"
)

const verifyEmailText = "Please open the link below to verify your email address.\n\n" +
	"{{.}}\n\n" +
	"If you did not create an account, no further action is required.\n"

func VerifyEmailText(url string) templ.ComponentFunc {
	return Text(template.Must(template.New("verify_email").Parse(verifyEmailText)), url)
}

package mail

import (
	_ "embed"
	"text/template"

	"github.com/a-h/templ"
)

//go:embed reset_password.txt
var resetPasswordText string

func ResetPasswordText(model ResetPassword) templ.ComponentFunc {
	return Text(template.Must(template.New("reset_password").Parse(resetPasswordText)), model)
}

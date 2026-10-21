package request

import (
	"net/url"

	"github.com/DannyFestor/go-template-html/internal/validate"
)

type PasswordReset struct {
	Token                string
	Password             string
	PasswordConfirmation string
}

func ParsePasswordReset(form url.Values) PasswordReset {
	return PasswordReset{
		Token:                Field(form, FieldToken),
		Password:             SecretField(form, FieldPassword),
		PasswordConfirmation: SecretField(form, FieldPasswordConfirmation),
	}
}

func (p PasswordReset) Validate() validate.Errors {
	errors := validate.Errors{}

	checkNewPassword(errors, p.Password, p.PasswordConfirmation)

	return errors
}

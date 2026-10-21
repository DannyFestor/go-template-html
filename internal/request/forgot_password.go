package request

import (
	"net/url"

	"github.com/DannyFestor/go-template-html/internal/validate"
)

type ForgotPassword struct {
	Email string
}

func ParseForgotPassword(form url.Values) ForgotPassword {
	return ForgotPassword{Email: Field(form, FieldEmail)}
}

func (f ForgotPassword) Validate() validate.Errors {
	errors := validate.Errors{}

	checkEmail(errors, f.Email)

	return errors
}

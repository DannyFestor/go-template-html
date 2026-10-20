package request

import (
	"net/url"

	"github.com/DannyFestor/go-template-html/internal/validate"
)

type Login struct {
	Email    string
	Password string
	Remember bool
}

func ParseLogin(form url.Values) Login {
	return Login{
		Email:    Field(form, FieldEmail),
		Password: SecretField(form, FieldPassword),
		Remember: form.Has(FieldRemember),
	}
}

func (l Login) Validate() validate.Errors {
	errors := validate.Errors{}

	errors.Check(validate.NotBlank(l.Email), FieldEmail, enterYourEmail)
	errors.Check(l.Password != "", FieldPassword, "Enter your password.")

	return errors
}

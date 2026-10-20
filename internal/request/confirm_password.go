package request

import (
	"net/url"

	"github.com/DannyFestor/go-template-html/internal/validate"
)

type ConfirmPassword struct {
	Password string
}

func ParseConfirmPassword(form url.Values) ConfirmPassword {
	return ConfirmPassword{Password: SecretField(form, FieldPassword)}
}

func (c ConfirmPassword) Validate() validate.Errors {
	errors := validate.Errors{}

	errors.Check(c.Password != "", FieldPassword, enterYourPassword)

	return errors
}

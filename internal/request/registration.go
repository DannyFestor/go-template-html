package request

import (
	"net/url"

	"github.com/DannyFestor/go-template-html/internal/validate"
)

type Registration struct {
	Name                 string
	Email                string
	Password             string
	PasswordConfirmation string
}

func ParseRegistration(form url.Values) Registration {
	return Registration{
		Name:                 Field(form, FieldName),
		Email:                Field(form, FieldEmail),
		Password:             SecretField(form, FieldPassword),
		PasswordConfirmation: SecretField(form, FieldPasswordConfirmation),
	}
}

func (r Registration) Validate() validate.Errors {
	errors := validate.Errors{}

	errors.Check(validate.NotBlank(r.Name), FieldName, "Enter your name.")
	checkEmail(errors, r.Email)
	checkNewPassword(errors, r.Password, r.PasswordConfirmation)

	return errors
}

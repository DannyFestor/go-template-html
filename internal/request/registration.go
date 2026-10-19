package request

import (
	"fmt"
	"net/url"

	"github.com/DannyFestor/go-template-html/internal/validate"
)

const minPasswordRunes = 8

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
	errors.Check(validate.NotBlank(r.Email), FieldEmail, "Enter your email address.")
	errors.Check(validate.EmailAddress(r.Email), FieldEmail, "Enter a valid email address.")
	errors.Check(validate.MinRunes(r.Password, minPasswordRunes), FieldPassword,
		fmt.Sprintf("The password must be at least %d characters.", minPasswordRunes))
	errors.Check(r.Password == r.PasswordConfirmation, FieldPasswordConfirmation, "The passwords do not match.")

	return errors
}

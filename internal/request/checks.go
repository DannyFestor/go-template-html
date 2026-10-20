package request

import (
	"fmt"

	"github.com/DannyFestor/go-template-html/internal/validate"
)

const minPasswordRunes = 8

func checkEmail(errors validate.Errors, email string) {
	errors.Check(validate.NotBlank(email), FieldEmail, enterYourEmail)
	errors.Check(validate.EmailAddress(email), FieldEmail, "Enter a valid email address.")
}

func checkNewPassword(errors validate.Errors, password, confirmation string) {
	errors.Check(validate.MinRunes(password, minPasswordRunes), FieldPassword,
		fmt.Sprintf("The password must be at least %d characters.", minPasswordRunes))
	errors.Check(password == confirmation, FieldPasswordConfirmation, "The passwords do not match.")
}

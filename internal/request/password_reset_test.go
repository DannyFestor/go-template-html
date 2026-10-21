package request_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/request"
	"github.com/DannyFestor/go-template-html/internal/validate"
)

func passwordResetForm() url.Values {
	return url.Values{
		"token":                 {"TOKEN"},
		"password":              {" new password "},
		"password_confirmation": {" new password "},
	}
}

func TestParsePasswordResetNeverTrimsThePassword(t *testing.T) {
	t.Parallel()

	assert.Equal(t, request.PasswordReset{
		Token:                "TOKEN",
		Password:             " new password ",
		PasswordConfirmation: " new password ",
	}, request.ParsePasswordReset(passwordResetForm()))
}

func TestAValidPasswordResetHasNoErrors(t *testing.T) {
	t.Parallel()

	assert.Empty(t, request.ParsePasswordReset(passwordResetForm()).Validate())
}

func TestAPasswordResetWithAPasswordUnderEightCharactersIsInvalid(t *testing.T) {
	t.Parallel()

	form := passwordResetForm()
	form.Set("password", "1234567")
	form.Set("password_confirmation", "1234567")

	assert.Equal(t, validate.Errors{"password": "The password must be at least 8 characters."},
		request.ParsePasswordReset(form).Validate())
}

func TestAPasswordResetWhoseConfirmationDiffersIsInvalid(t *testing.T) {
	t.Parallel()

	form := passwordResetForm()
	form.Set("password_confirmation", "another password")

	assert.Equal(t, validate.Errors{"password_confirmation": "The passwords do not match."},
		request.ParsePasswordReset(form).Validate())
}

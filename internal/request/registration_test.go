package request_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/request"
	"github.com/DannyFestor/go-template-html/internal/validate"
)

func registrationForm() url.Values {
	return url.Values{
		"name":                  {" Ada Lovelace "},
		"email":                 {" ada@example.com\n"},
		"password":              {" secret password "},
		"password_confirmation": {" secret password "},
	}
}

func registrationWith(field, value string) request.Registration {
	form := registrationForm()
	form.Set(field, value)

	return request.ParseRegistration(form)
}

func TestParseRegistrationTrimsTheNameAndEmailButNeverThePassword(t *testing.T) {
	t.Parallel()

	assert.Equal(t, request.Registration{
		Name:                 "Ada Lovelace",
		Email:                "ada@example.com",
		Password:             " secret password ",
		PasswordConfirmation: " secret password ",
	}, request.ParseRegistration(registrationForm()))
}

func TestAValidRegistrationHasNoErrors(t *testing.T) {
	t.Parallel()

	assert.Empty(t, request.ParseRegistration(registrationForm()).Validate())
}

func TestARegistrationWithoutNameEmailOrPasswordNamesEachMissingField(t *testing.T) {
	t.Parallel()

	assert.Equal(t, validate.Errors{
		"name":     "Enter your name.",
		"email":    "Enter your email address.",
		"password": "The password must be at least 8 characters.",
	}, request.ParseRegistration(url.Values{}).Validate())
}

func TestARegistrationWithAMalformedEmailIsInvalid(t *testing.T) {
	t.Parallel()

	assert.Equal(t, validate.Errors{"email": "Enter a valid email address."},
		registrationWith("email", "ada@").Validate())
}

func TestARegistrationWithAPasswordUnderEightCharactersIsInvalid(t *testing.T) {
	t.Parallel()

	form := registrationForm()
	form.Set("password", "1234567")
	form.Set("password_confirmation", "1234567")

	assert.Equal(t, validate.Errors{"password": "The password must be at least 8 characters."},
		request.ParseRegistration(form).Validate())
}

func TestARegistrationWhoseConfirmationDiffersIsInvalid(t *testing.T) {
	t.Parallel()

	assert.Equal(t, validate.Errors{"password_confirmation": "The passwords do not match."},
		registrationWith("password_confirmation", "another password").Validate())
}

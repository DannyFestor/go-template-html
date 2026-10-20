package request_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/request"
	"github.com/DannyFestor/go-template-html/internal/validate"
)

func TestParseLoginTrimsTheEmailButNeverThePassword(t *testing.T) {
	t.Parallel()

	form := url.Values{"email": {" ada@example.com\n"}, "password": {" secret "}, "remember": {"on"}}

	assert.Equal(t,
		request.Login{Email: "ada@example.com", Password: " secret ", Remember: true},
		request.ParseLogin(form),
	)
}

func TestALoginWithoutTheRememberBoxIsNotRemembered(t *testing.T) {
	t.Parallel()

	assert.False(t, request.ParseLogin(url.Values{"email": {"ada@example.com"}}).Remember)
}

func TestALoginWithAnEmailAndPasswordHasNoErrors(t *testing.T) {
	t.Parallel()

	assert.Empty(t, request.Login{Email: "ada@example.com", Password: " ", Remember: false}.Validate())
}

func TestALoginWithoutEmailOrPasswordNamesEachMissingField(t *testing.T) {
	t.Parallel()

	assert.Equal(t, validate.Errors{
		"email":    "Enter your email address.",
		"password": "Enter your password.",
	}, request.ParseLogin(url.Values{}).Validate())
}

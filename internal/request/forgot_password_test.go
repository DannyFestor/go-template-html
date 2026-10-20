package request_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/request"
	"github.com/DannyFestor/go-template-html/internal/validate"
)

func TestParseForgotPasswordTrimsTheEmail(t *testing.T) {
	t.Parallel()

	assert.Equal(t,
		request.ForgotPassword{Email: "ada@example.com"},
		request.ParseForgotPassword(url.Values{"email": {" ada@example.com\n"}}),
	)
}

func TestAForgotPasswordRequestWithAnEmailHasNoErrors(t *testing.T) {
	t.Parallel()

	assert.Empty(t, request.ForgotPassword{Email: "ada@example.com"}.Validate())
}

func TestAForgotPasswordRequestWithoutAnEmailAsksForIt(t *testing.T) {
	t.Parallel()

	assert.Equal(t, validate.Errors{"email": "Enter your email address."},
		request.ParseForgotPassword(url.Values{}).Validate())
}

func TestAForgotPasswordRequestWithAMalformedEmailIsInvalid(t *testing.T) {
	t.Parallel()

	assert.Equal(t, validate.Errors{"email": "Enter a valid email address."},
		request.ForgotPassword{Email: "ada@"}.Validate())
}

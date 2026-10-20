package request_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/request"
	"github.com/DannyFestor/go-template-html/internal/validate"
)

func TestParseConfirmPasswordNeverTrimsThePassword(t *testing.T) {
	t.Parallel()

	assert.Equal(t,
		request.ConfirmPassword{Password: " secret "},
		request.ParseConfirmPassword(url.Values{"password": {" secret "}}),
	)
}

func TestAConfirmationWithAPasswordHasNoErrors(t *testing.T) {
	t.Parallel()

	assert.Empty(t, request.ConfirmPassword{Password: " "}.Validate())
}

func TestAConfirmationWithoutAPasswordNamesTheField(t *testing.T) {
	t.Parallel()

	assert.Equal(t,
		validate.Errors{"password": "Enter your password."},
		request.ParseConfirmPassword(url.Values{}).Validate(),
	)
}

package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/validate"
)

func TestCheckRecordsTheMessageOfAFailedRule(t *testing.T) {
	t.Parallel()

	errors := validate.Errors{}

	errors.Check(false, "email", "Enter your email address.")

	assert.Equal(t, validate.Errors{"email": "Enter your email address."}, errors)
}

func TestCheckIgnoresAPassedRule(t *testing.T) {
	t.Parallel()

	errors := validate.Errors{}

	errors.Check(true, "email", "Enter your email address.")

	assert.Empty(t, errors)
}

func TestCheckKeepsTheFirstMessageOfAField(t *testing.T) {
	t.Parallel()

	errors := validate.Errors{}

	errors.Check(false, "email", "Enter your email address.")
	errors.Check(false, "email", "Enter a valid email address.")

	assert.Equal(t, validate.Errors{"email": "Enter your email address."}, errors)
}

package request_test

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/request"
)

func TestFieldReadsTheTrimmedValue(t *testing.T) {
	t.Parallel()

	form := url.Values{"email": {"  ada@example.com \n"}}

	assert.Equal(t, "ada@example.com", request.Field(form, "email"))
}

func TestFieldIsEmptyWhenTheFieldIsMissing(t *testing.T) {
	t.Parallel()

	assert.Empty(t, request.Field(url.Values{}, "email"))
}

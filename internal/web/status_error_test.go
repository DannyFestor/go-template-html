package web_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/web"
)

func TestIsTooManyRequestsSpotsARefusedAttemptThroughWrapping(t *testing.T) {
	t.Parallel()

	assert.True(t, web.IsTooManyRequests(fmt.Errorf("count: %w", web.ErrTooManyRequests(time.Minute))))
}

func TestIsTooManyRequestsIgnoresOtherErrors(t *testing.T) {
	t.Parallel()

	for name, err := range map[string]error{
		"none":                nil,
		"service unavailable": web.ErrServiceUnavailable(errValkeyDown),
		"plain":               errValkeyDown,
	} {
		assert.False(t, web.IsTooManyRequests(err), name)
	}
}

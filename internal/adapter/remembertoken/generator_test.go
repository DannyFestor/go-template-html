package remembertoken_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/adapter/remembertoken"
)

func TestATokenHasA128BitSelectorAndValidator(t *testing.T) {
	t.Parallel()

	token := remembertoken.Generator{}.New()

	assert.Regexp(t, `^[A-Z2-7]{26}$`, token.Selector)
	assert.Regexp(t, `^[A-Z2-7]{26}$`, token.Validator)
}

func TestTwoTokensDiffer(t *testing.T) {
	t.Parallel()

	first, second := remembertoken.Generator{}.New(), remembertoken.Generator{}.New()

	assert.NotEqual(t, first.Selector, second.Selector)
	assert.NotEqual(t, first.Validator, second.Validator)
}

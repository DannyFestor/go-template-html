package resettoken_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/adapter/resettoken"
)

func TestATokenHas128RandomBits(t *testing.T) {
	t.Parallel()

	assert.Regexp(t, `^[A-Z2-7]{26}$`, string(resettoken.Generator{}.New()))
}

func TestTwoTokensDiffer(t *testing.T) {
	t.Parallel()

	first, second := resettoken.Generator{}.New(), resettoken.Generator{}.New()

	assert.NotEqual(t, first, second)
}

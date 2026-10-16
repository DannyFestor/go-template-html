package recoverycode_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/DannyFestor/go-template-html/internal/adapter/recoverycode"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

func TestNewSetHoldsEightDistinctCodesInTwoGroupsOfTen(t *testing.T) {
	t.Parallel()

	distinct := map[domain.RecoveryCode]bool{}

	for _, code := range recoverycode.NewSet() {
		assert.Regexp(t, `^[A-Z2-7]{10}-[A-Z2-7]{10}$`, string(code))

		distinct[code] = true
	}

	assert.Len(t, distinct, 8)
}

func TestNewSetNeverRepeatsASet(t *testing.T) {
	t.Parallel()

	first, second := recoverycode.NewSet(), recoverycode.NewSet()

	assert.NotEqual(t, first, second)
}

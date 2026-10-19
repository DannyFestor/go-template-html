package totp_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/adapter/totp"
	"github.com/DannyFestor/go-template-html/internal/domain"
)

const (
	codeOfStep37037036 = "081804"
	codeOfStep37037037 = "050471"
)

func TestMatchAcceptsTheCodeOfTheCurrentStep(t *testing.T) {
	t.Parallel()

	step, ok := authenticatorAt(t, 1111111111).Match(seed(), codeOfStep37037037)

	assert.True(t, ok)
	assert.Equal(t, domain.TimeStep(37037037), step)
}

func TestMatchAcceptsTheCodeOfThePreviousStep(t *testing.T) {
	t.Parallel()

	step, ok := authenticatorAt(t, 1111111111).Match(seed(), codeOfStep37037036)

	assert.True(t, ok)
	assert.Equal(t, domain.TimeStep(37037036), step)
}

func TestMatchAcceptsTheCodeOfTheNextStep(t *testing.T) {
	t.Parallel()

	step, ok := authenticatorAt(t, 1111111109).Match(seed(), codeOfStep37037037)

	assert.True(t, ok)
	assert.Equal(t, domain.TimeStep(37037037), step)
}

func TestMatchRejectsACodeTwoStepsOld(t *testing.T) {
	t.Parallel()

	_, ok := authenticatorAt(t, 1111111111+60).Match(seed(), codeOfStep37037037)

	assert.False(t, ok)
}

func TestMatchRejectsACodeTwoStepsAhead(t *testing.T) {
	t.Parallel()

	_, ok := authenticatorAt(t, 1111111109-30).Match(seed(), codeOfStep37037037)

	assert.False(t, ok)
}

func TestMatchRejectsAWrongCode(t *testing.T) {
	t.Parallel()

	for _, code := range []string{"", "05047", "0504710", "050472", " 050471"} {
		_, ok := authenticatorAt(t, 1111111111).Match(seed(), code)

		assert.False(t, ok, "code %q", code)
	}
}

func TestMatchRejectsTheCodeOfAnotherSecret(t *testing.T) {
	t.Parallel()

	_, ok := authenticatorAt(t, 1111111111).Match(domain.TwoFactorSecret("09876543210987654321"), codeOfStep37037037)

	assert.False(t, ok)
}

func seed() domain.TwoFactorSecret {
	return domain.TwoFactorSecret(rfc6238SHA1Seed)
}

func authenticatorAt(t *testing.T, unix int64) *totp.Authenticator {
	t.Helper()

	clock := totp.NewMockClock(gomock.NewController(t))
	clock.EXPECT().Now().Return(time.Unix(unix, 0)).AnyTimes()

	return totp.NewAuthenticator(clock)
}

package uuidv7_test

import (
	"testing"
	"time"
	"uuid"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"

	"github.com/DannyFestor/go-template-html/internal/adapter/uuidv7"
)

func TestAnIDIsACanonicalUUIDv7(t *testing.T) {
	t.Parallel()

	id := generatorAt(t, rfc9562AppendixA6Time()).NewID()

	parsed, err := uuid.Parse(id)
	require.NoError(t, err)
	assert.Equal(t, parsed.String(), id)
	assert.Equal(t, byte(7), parsed[6]>>4, "version")
	assert.Equal(t, byte(0b10), parsed[8]>>6, "variant")
}

func TestAnIDStartsWithTheClocksTimeInMilliseconds(t *testing.T) {
	t.Parallel()

	id := generatorAt(t, rfc9562AppendixA6Time()).NewID()

	assert.Equal(t, "017f22e2-79b0-7", id[:15])
}

func TestIDsFromTheSameInstantDiffer(t *testing.T) {
	t.Parallel()

	generator := generatorAt(t, rfc9562AppendixA6Time())
	first, second := generator.NewID(), generator.NewID()

	assert.NotEqual(t, first, second)
}

func rfc9562AppendixA6Time() time.Time {
	return time.Date(2022, time.February, 22, 19, 22, 22, 0, time.UTC)
}

func generatorAt(t *testing.T, now time.Time) *uuidv7.Generator {
	t.Helper()

	clock := uuidv7.NewMockClock(gomock.NewController(t))
	clock.EXPECT().Now().Return(now).AnyTimes()

	return uuidv7.NewGenerator(clock)
}

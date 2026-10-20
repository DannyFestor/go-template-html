package argon2id_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/DannyFestor/go-template-html/internal/adapter/argon2id"
)

const argon2ReferenceCLIHash = "$argon2id$v=19$m=19456,t=2,p=1$cmVmZXJlbmNlLXNhbHQtMTY$XTZR93vezw24YlMX78dfRg0R8xH1lJXzBZUHfQG3QgE"

func TestAHashVerifiesAgainstItsPassword(t *testing.T) {
	t.Parallel()

	hasher := owaspHasher(t)

	matches, err := hasher.Verify("correct horse battery staple", hasher.Hash("correct horse battery staple"))

	require.NoError(t, err)
	assert.True(t, matches)
}

func TestAHashRejectsAnotherPassword(t *testing.T) {
	t.Parallel()

	hasher := owaspHasher(t)

	matches, err := hasher.Verify("Correct horse battery staple", hasher.Hash("correct horse battery staple"))

	require.NoError(t, err)
	assert.False(t, matches)
}

func TestTwoHashesOfOnePasswordDiffer(t *testing.T) {
	t.Parallel()

	hasher := owaspHasher(t)

	first, second := hasher.Hash("correct horse battery staple"), hasher.Hash("correct horse battery staple")

	assert.NotEqual(t, first, second)
}

func TestAHashFromTheReferenceImplementationVerifiesUnderItsOwnParameters(t *testing.T) {
	t.Parallel()

	hasher := hasherWith(t, argon2id.Params{MemoryKiB: 8, Iterations: 1, Parallelism: 1})

	matches, err := hasher.Verify("correct horse battery staple", argon2ReferenceCLIHash)

	require.NoError(t, err)
	assert.True(t, matches)
}

func TestAHashIsAPHCStringCarryingTheConfiguredParameters(t *testing.T) {
	t.Parallel()

	hash := owaspHasher(t).Hash("correct horse battery staple")

	assert.Regexp(t, `^\$argon2id\$v=19\$m=19456,t=2,p=1\$[A-Za-z0-9+/]{22}\$[A-Za-z0-9+/]{43}$`, hash)
}

func TestVerifyingAgainstAnUnusableHashFails(t *testing.T) {
	t.Parallel()

	unusableHashes := map[string]string{
		"bcrypt":           "$2y$10$abcdefghijklmnopqrstuuWcK4Z0bHnS8Xx9z0e6M1o2p3q4r5s6t",
		"argon2i":          "$argon2i$v=19$m=19456,t=2,p=1$cmVmZXJlbmNlLXNhbHQtMTY$XTZR93vezw24YlMX78dfRg0R8xH1lJXzBZUHfQG3QgE",
		"argon2 v1.0":      "$argon2id$v=16$m=19456,t=2,p=1$cmVmZXJlbmNlLXNhbHQtMTY$XTZR93vezw24YlMX78dfRg0R8xH1lJXzBZUHfQG3QgE",
		"missing field":    "$argon2id$v=19$m=19456,t=2,p=1$XTZR93vezw24YlMX78dfRg0R8xH1lJXzBZUHfQG3QgE",
		"garbled version":  "$argon2id$version$m=19456,t=2,p=1$cmVmZXJlbmNlLXNhbHQtMTY$XTZR93vezw24YlMX78dfRg0R8xH1lJXzBZUHfQG3QgE",
		"garbled params":   "$argon2id$v=19$m=lots,t=2,p=1$cmVmZXJlbmNlLXNhbHQtMTY$XTZR93vezw24YlMX78dfRg0R8xH1lJXzBZUHfQG3QgE",
		"zero iterations":  "$argon2id$v=19$m=19456,t=0,p=1$cmVmZXJlbmNlLXNhbHQtMTY$XTZR93vezw24YlMX78dfRg0R8xH1lJXzBZUHfQG3QgE",
		"zero parallelism": "$argon2id$v=19$m=19456,t=2,p=0$cmVmZXJlbmNlLXNhbHQtMTY$XTZR93vezw24YlMX78dfRg0R8xH1lJXzBZUHfQG3QgE",
		"garbled salt":     "$argon2id$v=19$m=19456,t=2,p=1$not*base64$XTZR93vezw24YlMX78dfRg0R8xH1lJXzBZUHfQG3QgE",
		"garbled key":      "$argon2id$v=19$m=19456,t=2,p=1$cmVmZXJlbmNlLXNhbHQtMTY$not*base64",
		"16-byte key":      "$argon2id$v=19$m=19456,t=2,p=1$cmVmZXJlbmNlLXNhbHQtMTY$ll8H5zQ6UIJZBwOfAN48wg",
	}

	for name, unusableHash := range unusableHashes {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := owaspHasher(t).Verify("correct horse battery staple", unusableHash)

			require.Error(t, err)
		})
	}
}

func TestAHashMadeWithTheConfiguredParametersNeedsNoRehash(t *testing.T) {
	t.Parallel()

	hasher := owaspHasher(t)

	assert.False(t, hasher.NeedsRehash(hasher.Hash("correct horse battery staple")))
}

func TestAHashMadeWithOtherParametersNeedsARehash(t *testing.T) {
	t.Parallel()

	hasher := hasherWith(t, argon2id.Params{MemoryKiB: 8, Iterations: 1, Parallelism: 1})

	assert.True(t, hasher.NeedsRehash(argon2ReferenceCLIHash))
}

func TestAnUnusableHashNeedsARehash(t *testing.T) {
	t.Parallel()

	assert.True(t, owaspHasher(t).NeedsRehash("$2y$10$abcdefghijklmnopqrstuuWcK4Z0bHnS8Xx9z0e6M1o2p3q4r5s6t"))
}

func TestAHasherRefusesParametersArgon2CannotRunWith(t *testing.T) {
	t.Parallel()

	for name, params := range map[string]argon2id.Params{
		"zero iterations":  {MemoryKiB: 19 * 1024, Iterations: 0, Parallelism: 1},
		"zero parallelism": {MemoryKiB: 19 * 1024, Iterations: 2, Parallelism: 0},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			_, err := argon2id.NewHasher(params)

			require.Error(t, err)
		})
	}
}

func owaspHasher(t *testing.T) *argon2id.Hasher {
	t.Helper()

	return hasherWith(t, argon2id.Params{MemoryKiB: 19 * 1024, Iterations: 2, Parallelism: 1})
}

func hasherWith(t *testing.T, params argon2id.Params) *argon2id.Hasher {
	t.Helper()

	hasher, err := argon2id.NewHasher(params)
	require.NoError(t, err)

	return hasher
}

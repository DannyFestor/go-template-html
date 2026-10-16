package worker

import (
	"math/rand/v2"
	"time"
)

const (
	backoffFloor     = 15 * time.Second
	backoffExponent  = 4
	jitterPerAttempt = 10 * time.Second
)

func PolynomialBackoff(attempt int) time.Duration {
	polynomial := time.Duration(power(attempt, backoffExponent)) * time.Second

	return polynomial + backoffFloor + jitter(attempt)
}

func jitter(attempt int) time.Duration {
	ceiling := jitterPerAttempt * time.Duration(attempt+1)

	return rand.N(ceiling) //nolint:gosec // Jitter needs no cryptographic randomness.
}

func power(base, exponent int) int {
	result := 1
	for range exponent {
		result *= base
	}

	return result
}

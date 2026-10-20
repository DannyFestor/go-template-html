package bootstrap

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var errMalformedRateLimit = errors.New("want a positive number of attempts per positive window, such as 5/1m")

type RateLimits struct {
	Registration RateLimit `env:"RATE_LIMIT_REGISTRATION" envDefault:"5/1m"`
	Login        RateLimit `env:"RATE_LIMIT_LOGIN"        envDefault:"5/1m"`
}

type RateLimit struct {
	MaxAttempts int
	Window      time.Duration
}

func (l *RateLimit) UnmarshalText(text []byte) error {
	parsed, ok := parseRateLimit(string(text))
	if !ok {
		return fmt.Errorf("rate limit %q: %w", text, errMalformedRateLimit)
	}

	*l = parsed

	return nil
}

func parseRateLimit(text string) (RateLimit, bool) {
	attempts, window, found := strings.Cut(text, "/")
	if !found {
		return RateLimit{MaxAttempts: 0, Window: 0}, false
	}

	maxAttempts, attemptsErr := strconv.Atoi(attempts)
	duration, windowErr := time.ParseDuration(window)
	limit := RateLimit{MaxAttempts: maxAttempts, Window: duration}

	return limit, attemptsErr == nil && windowErr == nil && maxAttempts > 0 && duration > 0
}

package auth

import "context"

const rateLimitedMessage = "Too many attempts. Please try again later."

type rateLimit interface {
	Hit(ctx context.Context, key string) error
}

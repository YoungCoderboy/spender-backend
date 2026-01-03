package domain

import "context"

// Define what Ai client should do.
type AIClient interface {
	ProcessText(ctx context.Context, text string) (*Transaction, error)
}

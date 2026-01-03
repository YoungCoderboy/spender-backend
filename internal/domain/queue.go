package domain

import "context"

// Defines what redis must do.
type ExpenseQueue interface {
	Push(ctx context.Context, userId, msg string) error
	Pop(ctx context.Context) (string, string, error)
}

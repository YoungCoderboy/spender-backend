package redis

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type ExpenseQueue struct {
	rdb *redis.Client
	key string
}

func NewExpenseQueue(client *redis.Client) *ExpenseQueue {
	return &ExpenseQueue{
		rdb: client,
		key: "expense_processing_queue",
	}
}

// Push Data to Redis.
func (eq *ExpenseQueue) Push(ctx context.Context, userID, rawMessage string) error {
	payload := map[string]string{
		"user_id":  userID,
		"raw_text": rawMessage,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal queue payload: %w", err)
	}

	// RPUSH adds to the tail of the list
	return eq.rdb.RPush(ctx, eq.key, data).Err()
}

// Pop removes and returns the first element from the list
func (eq *ExpenseQueue) Pop(ctx context.Context) (string, string, error) {
	// LPop removes from the head of the list
	result, err := eq.rdb.LPop(ctx, eq.key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", "", nil // Queue is empty
		}
		return "", "", err
	}

	var payload map[string]string
	if err := json.Unmarshal([]byte(result), &payload); err != nil {
		return "", "", err
	}

	return payload["user_id"], payload["raw_text"], nil
}

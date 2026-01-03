package domain

import (
	"context"
	"time"
)

type Transaction struct {
	Id            string    `json:"id"`
	Type          string    `json:"type"`
	UserId        string    `json:"userId"`
	Amount        float64   `json:"amount"`
	Date          time.Time `json:"date"`
	IsTransaction int       `json:"isTransaction"`
}

// Defines what postgres must do for transactionsTable.
type TransactionRepository interface {
	CreateTransaction(ctx context.Context, transaction *Transaction) error
	GetTransactionByUserId(ctx context.Context, userId string) ([]*Transaction, error)
}

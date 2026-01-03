package postgres

import (
	"context"
	"database/sql"
	"spender-backend/v2/internal/domain"
	"time"
)

type TransRepo struct {
	db *sql.DB
}

func NewTransRepo(db *sql.DB) *TransRepo {
	return &TransRepo{db: db}
}

func (r *TransRepo) CreateTransaction(ctx context.Context, t *domain.Transaction) error {
	query := `
		INSERT INTO transactions (user_id, type, amount, transaction_date)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	// If Date is zero, we use time.Now()
	if t.Date.IsZero() {
		t.Date = time.Now()
	}

	return r.db.QueryRowContext(ctx, query,
		t.UserId, t.Type, t.Amount, t.Date,
	).Scan(&t.Id)
}

func (r *TransRepo) GetTransactionByUserId(ctx context.Context, userId string) ([]*domain.Transaction, error) {
	query := `
		SELECT id, transaction_date, type, user_id,
		FROM transactions 
		WHERE user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []*domain.Transaction
	for rows.Next() {
		t := &domain.Transaction{}
		err := rows.Scan(&t.Id, &t.Date, &t.Type, &t.UserId, &t.Amount)
		if err != nil {
			return nil, err
		}
		txs = append(txs, t)
	}
	return txs, nil
}

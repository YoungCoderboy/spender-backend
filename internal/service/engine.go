package service

import (
	"context"
	"spender-backend/v2/internal/domain"
	"time"

	"go.uber.org/zap"
)

type EngineService struct {
	TransactionSvc  domain.TransactionRepository
	ExpenseQueueSvc domain.ExpenseQueue
	AiClientSvc     domain.AIClient
	Logger          *zap.Logger
}

func InitExpenseEngine(transRepo domain.TransactionRepository, queueRepo domain.ExpenseQueue, aiClient domain.AIClient, logger *zap.Logger) *EngineService {
	return &EngineService{
		TransactionSvc:  transRepo,
		ExpenseQueueSvc: queueRepo,
		AiClientSvc:     aiClient,
		Logger:          logger,
	}
}

func (es *EngineService) Run(ctx context.Context) error {
	es.Logger.Info("Expense Engine started and polling for jobs...")

	for {
		select {
		case <-ctx.Done():
			es.Logger.Info("Engine shutting down gracefully...")
			return ctx.Err()
		default:
			// 1. Fetch from Redis (Blocking Pop or Polling)
			userID, rawText, err := es.ExpenseQueueSvc.Pop(ctx)
			if err != nil {
				es.Logger.Error("Queue Pop failed", zap.Error(err))
				time.Sleep(2 * time.Second) // Prevent tight-looping on error
				continue
			}

			// If queue is empty, userID will be empty
			if userID == "" {
				time.Sleep(1 * time.Second)
				continue
			}

			// 2. Process with AI
			es.Logger.Info("Processing message for user", zap.String("user_id", userID))
			tx, err := es.AiClientSvc.ProcessText(ctx, rawText)
			if err != nil {
				es.Logger.Error("AI Processing failed", zap.Error(err), zap.String("raw_text", rawText))
				// Consider: Push to a 'dead_letter_queue' in Redis for manual review
				continue
			}

			// 3. Attach Metadata & Save to Postgres
			tx.UserId = userID
			if tx.Date.IsZero() {
				tx.Date = time.Now()
			}

			if err := es.TransactionSvc.CreateTransaction(ctx, tx); err != nil {
				es.Logger.Error("Failed to save transaction", zap.Error(err), zap.Any("tx", tx))
				continue
			}

			es.Logger.Info("Transaction successfully processed", zap.String("type", tx.Type), zap.Float64("amount", tx.Amount))
		}
	}

}

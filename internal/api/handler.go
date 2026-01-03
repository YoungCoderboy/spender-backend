package api

import (
	"net/http"
	"spender-backend/v2/internal/domain"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type TransactionHandler struct {
	Queue  domain.ExpenseQueue
	Logger *zap.Logger
}

// NotifyRequest matches what the mobile app sends
type NotifyRequest struct {
	RawText string `json:"rawText" binding:"required"`
}

func NewTransactionQueueHandler(q domain.ExpenseQueue, l *zap.Logger) *TransactionHandler {
	return &TransactionHandler{
		Queue:  q,
		Logger: l,
	}
}

func (th *TransactionHandler) HandleNotification(ctx *gin.Context) {
	userIdVal, exists := ctx.Get("userID")
	if !exists {
		th.Logger.Error("User ID not found in context")
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized context"})
		return
	}

	userId, ok := userIdVal.(string)
	if !ok {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user identity format"})
		return
	}

	var req NotifyRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request: rawText is required"})
		return
	}

	// 3. Push to Redis Queue
	// We use c.Request.Context() so the operation is canceled if the user disconnects
	err := th.Queue.Push(ctx.Request.Context(), userId, req.RawText)
	if err != nil {
		th.Logger.Error("Failed to queue message", zap.Error(err), zap.String("user_id", userId))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process notification"})
		return
	}

	// 4. Return 202 Accepted
	// We use 202 because the processing (AI & DB) is happening asynchronously in the background
	th.Logger.Info("Notification queued successfully", zap.String("user_id", userId))
	ctx.JSON(http.StatusAccepted, gin.H{
		"message": "Notification received and is being processed",
	})
}

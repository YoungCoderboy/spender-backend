package api

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"spender-backend/v2/internal/domain"
	"spender-backend/v2/internal/service"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func StartServer(ctx context.Context, port int, queue domain.ExpenseQueue, authSvc *service.AuthService, logger *zap.Logger) error {
	// Set Gin to release mode in production
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// 1. Initialize Handlers
	authHandler := &AuthHandler{AuthSvc: authSvc}
	transHandler := NewTransactionQueueHandler(queue, logger)

	// 2. Define Routes
	router.POST("/register", authHandler.Register)
	router.POST("/login", authHandler.Login)

	apiGroup := router.Group("/api")
	apiGroup.Use(AuthMiddleware(authSvc.JwtSecret))
	{
		apiGroup.POST("/notify", transHandler.HandleNotification)
	}

	// 3. Configure the HTTP Server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
		// Good practice: set timeouts to prevent slow-loris attacks
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// 4. Graceful Shutdown Logic
	// We start a goroutine that waits for the parent context to be canceled
	go func() {
		<-ctx.Done()
		logger.Info("Shutting down HTTP server signals received...")

		// Create a short-lived context for the shutdown process (e.g., 5 seconds)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			logger.Error("HTTP server forced to shutdown", zap.Error(err))
		}
	}()

	// 5. Start Listening
	// This will block until the server is closed or an error occurs
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("listen and serve error: %w", err)
	}

	logger.Info("HTTP server stopped")
	return nil
}

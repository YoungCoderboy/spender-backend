package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"spender-backend/v2/internal/ai/gemini"
	"spender-backend/v2/internal/api"
	"spender-backend/v2/internal/repository/postgres"
	"spender-backend/v2/internal/repository/redis"
	"spender-backend/v2/internal/service"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Unable to Initialize Logger {error: %v}", err)
	}
	defer logger.Sync()

	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	ctx, cancel := context.WithCancel(context.Background())

	// 3. Load Environment Variables (Ensure these are set in your .env or shell)
	dsn := os.Getenv("DATABASE_URL")
	redisAddr := os.Getenv("REDIS_ADDR")
	geminiKey := os.Getenv("GEMINI_API_KEY")
	jwtSecret := os.Getenv("JWT_SECRET")

	if dsn == "" || redisAddr == "" || geminiKey == "" || jwtSecret == "" {
		logger.Fatal("Missing required environment variables")
		return
	}

	pgDB, err := postgres.NewPostGresConn(dsn)
	if err != nil {
		log.Fatalf("Unable to Initialize DB {error: %v}", err)
		return
	}
	defer pgDB.Close()

	rdb, err := redis.NewRedisConn(redisAddr)
	if err != nil {
		log.Fatalf("Unable to Initialize redis db {error: %v}", err)
		return
	}
	defer rdb.Close()

	transactionRepo := postgres.NewTransRepo(pgDB)
	userRepo := postgres.NewUserRepo(pgDB)
	rdbRepo := redis.NewExpenseQueue(rdb)

	// LLM Service
	aiClient, err := gemini.NewClient(ctx, geminiKey)
	if err != nil {
		log.Fatalf("Unable to Initialize Ai Client {error: %v}", err)
		return
	}

	authSvc := service.NewAuthService(userRepo, jwtSecret)
	expenseEngine := service.InitExpenseEngine(
		transactionRepo,
		rdbRepo,
		aiClient,
		logger,
	)

	go func() {
		logger.Info("Starting background engine worker...")
		if err := expenseEngine.Run(ctx); err != nil {
			logger.Error("engine stopped with error", zap.Error(err))
		}
	}()

	// 6. Start API Server (Goroutine 2)
	// We run this in a goroutine so main doesn't block here
	go func() {
		logger.Info("Starting API server", zap.String("port", "8080"))
		if err := api.StartServer(ctx, 8080, rdbRepo, authSvc, logger.Named("api")); err != nil {
			// Ignore error if it was a planned shutdown
			if !errors.Is(err, http.ErrServerClosed) {
				logger.Fatal("API server crashed", zap.Error(err))
			}
		}
	}()

	// 7. Wait for Shutdown Signal
	// This blocks main until you press Ctrl+C or the system sends a kill signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	sig := <-quit
	logger.Info("Shutdown signal received", zap.String("signal", sig.String()))

	// 8. Graceful Shutdown Period
	// Give the engine and API 5 seconds to wrap up current work
	cancel() // This triggers ctx.Done() in engine and server

	logger.Info("Waiting for services to exit...")
	time.Sleep(2 * time.Second)

	logger.Info("Spender Backend stopped clean.")
}

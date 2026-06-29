package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"llm-gateway/internal/shared/config"
	"llm-gateway/internal/shared/database"
	"llm-gateway/internal/shared/event"
	"llm-gateway/internal/shared/redis"
)

func main() {
	if err := godotenv.Load(); err != nil {
		slog.Info("no .env file found, using system environment variables")
	}

	cfg := config.Load()
	if err := database.Init(cfg); err != nil {
		slog.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer database.GetDB().Close()

	if err := redis.Init(cfg); err != nil {
		slog.Error("failed to initialize redis", "error", err)
		os.Exit(1)
	}
	defer redis.Close()

	consumerName := os.Getenv("WORKER_CONSUMER_NAME")
	if consumerName == "" {
		hostname, _ := os.Hostname()
		consumerName = fmt.Sprintf("%s-%d", hostname, time.Now().Unix())
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	consumer := event.NewConsumer(redis.GetClient(), database.GetDB(), consumerName)

	slog.Info("worker starting", "consumer", consumerName, "stream", event.RequestCompletedStream, "group", event.RequestCompletedGroup)
	if err := consumer.Run(ctx); err != nil && err != context.Canceled {
		slog.Error("worker stopped with error", "error", err)
		os.Exit(1)
	}
	slog.Info("worker stopped")
}

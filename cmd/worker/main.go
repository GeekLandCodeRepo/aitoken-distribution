package main

import (
	"context"
	"fmt"
	"log"
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
		log.Println("No .env file found, using system environment variables")
	}

	cfg := config.Load()
	if err := database.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.GetDB().Close()

	if err := redis.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize redis: %v", err)
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

	log.Printf("Worker starting: consumer=%s stream=%s group=%s", consumerName, event.RequestCompletedStream, event.RequestCompletedGroup)
	if err := consumer.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("Worker stopped with error: %v", err)
	}
	log.Println("Worker stopped")
}

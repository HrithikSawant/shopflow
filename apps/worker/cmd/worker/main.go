package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/HrithikSawant/shopflow-worker/internal/jobs"
	"github.com/HrithikSawant/shopflow-worker/internal/rabbitmq"
)

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")

	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("starting worker")

	if err := rabbitmq.Consume(ctx, rabbitURL, jobs.HandleOrderCreated); err != nil {
		log.Fatal(err)
	}

	log.Println("worker stopped")
}

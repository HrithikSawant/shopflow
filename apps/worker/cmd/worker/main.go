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
		log.Fatal("RABBITMQ_URL is required (it carries the credentials; there is no default)")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("starting worker")

	if err := rabbitmq.Consume(ctx, rabbitURL, jobs.HandleOrderCreated); err != nil {
		log.Fatal(err)
	}

	log.Println("worker stopped")
}

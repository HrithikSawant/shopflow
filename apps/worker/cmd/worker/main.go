package main

import (
	"log"
	"os"

	"github.com/HrithikSawant/shopflow-worker/internal/jobs"
	"github.com/HrithikSawant/shopflow-worker/internal/rabbitmq"
)

func main() {
	rabbitURL := os.Getenv("RABBITMQ_URL")

	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	log.Println("starting worker")

	if err := rabbitmq.Consume(
		rabbitURL,
		jobs.HandleOrderCreated,
	); err != nil {
		log.Fatal(err)
	}
}

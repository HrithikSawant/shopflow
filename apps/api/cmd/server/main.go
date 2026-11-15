package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/HrithikSawant/shopflow-api/internal/config"
	"github.com/HrithikSawant/shopflow-api/internal/database"
	"github.com/HrithikSawant/shopflow-api/internal/outbox"
	"github.com/HrithikSawant/shopflow-api/internal/rabbitmq"
	"github.com/HrithikSawant/shopflow-api/internal/router"
)

func main() {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	database.DB = pool

	publisher := rabbitmq.NewPublisher(cfg.RabbitMQURL)
	defer publisher.Close()

	go outbox.Run(ctx, pool, publisher, time.Second)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router.New(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_ = srv.Shutdown(shutdownCtx)
	}()

	log.Printf("starting api on :%s\n", cfg.Port)

	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}

	log.Println("api stopped")
}

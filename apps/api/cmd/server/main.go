package main

import (
	"context"
	"log"

	"github.com/HrithikSawant/shopflow-api/internal/config"
	"github.com/HrithikSawant/shopflow-api/internal/database"
	"github.com/HrithikSawant/shopflow-api/internal/router"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close(context.Background())

	log.Println("postgres connected")

	r := router.New()

	log.Printf("starting api on :%s\n", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}
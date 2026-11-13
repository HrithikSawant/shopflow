package main

import (
	"log"

	"github.com/HrithikSawant/shopflow-api/internal/config"
	"github.com/HrithikSawant/shopflow-api/internal/router"
)

func main() {
	cfg := config.Load()

	r := router.New()

	log.Printf("starting api on :%s\n", cfg.Port)

	if err := r.Run(":" + cfg.Port); err != nil {
		log.Fatal(err)
	}
}

package main

import (
	"fmt"
	"log"

	"github.com/kiokyky/MnyMngmnt-Go/internal/config"
	"github.com/kiokyky/MnyMngmnt-Go/internal/handler"
)

func main() {
	cfg := config.LoadConfig()

	db, err := config.ConnectDB(cfg)
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	r := handler.SetupRouter(db, cfg)

	addr := fmt.Sprintf(":%s", cfg.Port)
	log.Printf("Money Management & Forex Tracker API running on http://localhost%s", addr)
	log.Printf("Swagger UI available at http://localhost%s/swagger", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}


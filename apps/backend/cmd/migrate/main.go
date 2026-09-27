// Command migrate applies (or checks) database migrations.
//
// Usage:
//
//	go run ./cmd/migrate up
package main

import (
	"context"
	"log"
	"os"

	"github.com/kovalit/secrets-center/backend/internal/config"
	"github.com/kovalit/secrets-center/backend/internal/db"
)

func main() {
	cmd := "up"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	if cmd != "up" {
		log.Fatalf("unknown command %q (supported: up)", cmd)
	}

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx := context.Background()
	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer pool.Close()

	if err := db.Migrate(ctx, pool); err != nil {
		log.Fatalf("migrate: %v", err)
	}
	log.Println("migrations up to date")
}

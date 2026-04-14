package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/seWy-bit/GO-and-eat/internal/pkg/config"
	"github.com/seWy-bit/GO-and-eat/internal/pkg/database"
	"github.com/seWy-bit/GO-and-eat/internal/pkg/migrate"
)

func main() {
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatal("failed to load config: %w", err)
	}

	dbCfg := database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    1,
		MaxIdleConns:    1,
		ConnMaxLifetime: 5 * time.Minute,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, dbCfg)
	if err != nil {
		log.Fatal("failed to connect to database:", err)
	}

	defer pool.Close()

	runner := migrate.NewRunner(pool.Pool)

	args := os.Args[1:]
	if len(args) > 0 && args[0] == "down" {
		log.Println("Rolling back last migration...")
		if err := runner.Down(ctx, "scripts/migrations/restaurant"); err != nil {
			log.Fatal("failed to rollback migration:", err)
		}
		log.Printf("migration rollback completed")
		return
	}

	log.Println("Applying migrations...")
	if err := runner.Up(ctx, "scripts/migrations/restaurant"); err != nil {
		log.Fatal("failed to apply migrations:", err)
	}
	log.Printf("migrations applied successfully")
}

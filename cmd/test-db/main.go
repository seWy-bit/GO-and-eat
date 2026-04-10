package main

import (
	"context"
	"log"
	"time"

	"github.com/seWy-bit/GO-and-eat/internal/pkg/config"
	"github.com/seWy-bit/GO-and-eat/internal/pkg/database"
)

func main() {
	// Загружаем конфиг
	cfg, err := config.Load("configs/config.yaml")
	if err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// Создаем подключение к БД
	dbCfg := database.Config{
		Host:            cfg.Database.Host,
		Port:            cfg.Database.Port,
		User:            cfg.Database.User,
		Password:        cfg.Database.Password,
		DBName:          cfg.Database.DBName,
		SSLMode:         cfg.Database.SSLMode,
		MaxOpenConns:    cfg.Database.MaxOpenConns,
		MaxIdleConns:    cfg.Database.MaxIdleConns,
		ConnMaxLifetime: cfg.Database.ConnMaxLifetime,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, dbCfg)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer pool.Close()

	log.Println("✅ Successfully connected to PostgreSQL!")

	// Проверяем, что можем выполнить запрос
	var result int
	err = pool.QueryRow(ctx, "SELECT 1").Scan(&result)
	if err != nil {
		log.Fatal("Query failed:", err)
	}

	log.Println("✅ Query executed successfully, result:", result)
}

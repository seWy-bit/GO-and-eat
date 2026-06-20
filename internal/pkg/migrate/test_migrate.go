package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ApplyMigrationsForTest применяет миграции для тестов
// Использует те же миграции, что и основное приложение
func ApplyMigrationsForTest(ctx context.Context, pool *pgxpool.Pool, migrationsDir string) error {
	createTableQuery := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
	`
	if _, err := pool.Exec(ctx, createTableQuery); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	var appliedNames []string
	rows, err := pool.Query(ctx, "SELECT name FROM schema_migrations")
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return fmt.Errorf("failed to scan migration name: %w", err)
		}
		appliedNames = append(appliedNames, name)
	}

	// Создаём map для быстрой проверки
	appliedMap := make(map[string]bool)
	for _, name := range appliedNames {
		appliedMap[name] = true
	}

	files, err := filepath.Glob(filepath.Join(migrationsDir, "*.up.sql"))
	if err != nil {
		return fmt.Errorf("failed to find migration files: %w", err)
	}
	sort.Strings(files)

	for _, file := range files {
		migrationName := filepath.Base(file)

		// Проверяем, не применена ли уже
		if appliedMap[migrationName] {
			continue
		}

		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		// Применяем миграцию в транзакции
		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migrationName, err)
		}

		// Записываем, что миграция применена
		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (name) VALUES ($1)", migrationName); err != nil {
			return fmt.Errorf("failed to record migration %s: %w", migrationName, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migrationName, err)
		}
	}

	return nil
}

// CleanTables очищает все таблицы после каждого теста
func CleanTables(ctx context.Context, pool *pgxpool.Pool) error {
	// Удаляем данные в правильном порядке (с учётом внешних ключей)
	queries := []string{
		"DELETE FROM order_items;",
		"DELETE FROM orders;",
		"DELETE FROM menu_items;",
		"DELETE FROM restaurants;",
		"DELETE FROM schema_migrations;", // Чтобы миграции применялись заново
	}

	for _, query := range queries {
		if _, err := pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to clean table: %w", err)
		}
	}

	return nil
}

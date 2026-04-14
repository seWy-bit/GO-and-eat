package migrate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Runner выполняет миграции
type Runner struct {
	pool *pgxpool.Pool
}

func NewRunner(pool *pgxpool.Pool) *Runner {
	return &Runner{pool: pool}
}

// Up выполняет миграции из директории migrationsDir
func (r *Runner) Up(ctx context.Context, migrationsDir string) error {
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*up.sql"))
	if err != nil {
		return fmt.Errorf("failed to find migration files: %w", err)
	}

	sort.Strings(files)

	if err := r.createMigrationTable(ctx); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	applied, err := r.getAppliedMigrations(ctx)
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	for _, file := range files {
		migrationName := filepath.Base(file)

		if applied[migrationName] {
			continue
		}

		fmt.Printf("Applying migration: %s\n", migrationName)

		sqlBytes, err := os.ReadFile(file)
		if err != nil {
			return fmt.Errorf("failed to read migration file %s: %w", file, err)
		}

		tx, err := r.pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback(ctx)

		if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
			return fmt.Errorf("failed to execute migration %s: %w", migrationName, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (name) VALUES ($1)", migrationName); err != nil {
			return fmt.Errorf("failed to record applied migration %s: %w", migrationName, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("failed to commit migration %s: %w", migrationName, err)
		}

		fmt.Printf("Applied migration: %s\n", migrationName)
	}

	return nil
}

func (r *Runner) createMigrationTable(ctx context.Context) error {
	query := `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id SERIAL PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		);
	`
	_, err := r.pool.Exec(ctx, query)
	return err
}

func (r *Runner) Down(ctx context.Context, migrationsDir string) error {
	files, err := filepath.Glob(filepath.Join(migrationsDir, "*down.sql"))
	if err != nil {
		return fmt.Errorf("failed to find migration files: %w", err)
	}

	sort.Sort(sort.Reverse(sort.StringSlice(files)))

	if len(files) == 0 {
		return fmt.Errorf("no down migration files found")
	}

	lastMigration := files[0]
	migrationName := filepath.Base(lastMigration)

	fmt.Printf("Rolling back migration: %s\n", migrationName)

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer tx.Rollback(ctx)

	sqlBytes, err := os.ReadFile(lastMigration)
	if err != nil {
		return fmt.Errorf("failed to read migration file: %w", err)
	}

	if _, err := tx.Exec(ctx, string(sqlBytes)); err != nil {
		return fmt.Errorf("failed to execute migration down: %w", err)
	}

	baseName := strings.TrimSuffix(migrationName, ".down.sql")
	if _, err := tx.Exec(ctx, "DELETE FROM schema_migrations WHERE name = $1", baseName+"up.sql"); err != nil {
		return fmt.Errorf("failed to remove migration record: %w", err)
	}

	fmt.Printf("Rolled back migration: %s\n", migrationName)

	return nil
}

func (r *Runner) getAppliedMigrations(ctx context.Context) (map[string]bool, error) {
	rows, err := r.pool.Query(ctx, "SELECT name FROM schema_migrations")
	if err != nil {
		return nil, fmt.Errorf("failed to query applied migrations: %w", err)
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, fmt.Errorf("failed to scan applied migration: %w", err)
		}
		applied[name] = true
	}

	return applied, nil
}

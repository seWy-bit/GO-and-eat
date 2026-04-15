package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seWy-bit/GO-and-eat/internal/restaurant/domain"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(pool *pgxpool.Pool) *PostgresStorage {
	return &PostgresStorage{pool: pool}
}

func (s *PostgresStorage) CreateRestaurant(r domain.Restaurant) error {
	ctx := context.Background()

	query := `
		INSERT INTO restaurants (id, name, adress, phone, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`

	_, err := s.pool.Exec(ctx, query, r.ID, r.Name, r.Address, r.Phone, r.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key") {
			return errors.New("restaurant already exists")
		}
		return fmt.Errorf("failed to create restaurant: %w", err)
	}

	return nil
}

func (s *PostgresStorage) GetMenu(restaurantID string) ([]domain.MenuItem, error) {
	ctx := context.Background()

	query := `
		SELECT id, restaurant_id, name, description, price, stock, available, created_at
		FROM menu_items
		WHERE restaurant_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.pool.Query(ctx, query, restaurantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get menu: %w", err)
	}
	defer rows.Close()

	var items []domain.MenuItem

	for rows.Next() {
		var item domain.MenuItem
		err := rows.Scan(
			&item.ID,
			&item.RestaurantID,
			&item.Name,
			&item.Description,
			&item.Price,
			&item.Stock,
			&item.Available,
			&item.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan menu item: %w", err)
		}
		items = append(items, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating menu items: %w", err)
	}

	if items == nil {
		return []domain.MenuItem{}, nil // Возвращаем пустой список, если блюд нет
	}

	return items, nil
}

func (s *PostgresStorage) AddMenuItem(item domain.MenuItem) error {
	ctx := context.Background()

	exists, err := s.restaurantExists(ctx, item.RestaurantID)
	if err != nil {
		return fmt.Errorf("failed to check restaurant: %w", err)
	}
	if !exists {
		return errors.New("restaurant not found")
	}

	query := `
		INSERT INTO menu_items (id, restaurant_id, name, description, price, stock, available, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`

	_, err = s.pool.Exec(ctx, query,
		item.ID,
		item.RestaurantID,
		item.Name,
		item.Description,
		item.Price,
		item.Stock,
		item.Available,
		item.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add menu item: %w", err)
	}

	return nil
}

func (s *PostgresStorage) DecreaseStock(restaurantID, menuItemID string, quantity int) error {
	ctx := context.Background()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var currentStock int
	checkQuery := `SELECT stock FROM menu_items WHERE id = $1 AND restaurant_id = $2`
	err = tx.QueryRow(ctx, checkQuery, menuItemID, restaurantID).Scan(&currentStock)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("menu item not found")
		}
		return fmt.Errorf("failed to check menu item stock: %w", err)
	}

	if currentStock < quantity {
		return errors.New("not enough stock")
	}

	updateQuery := `
		UPDATE menu_items
		SET stock = stock - $1,
			available = (stock - $1) > 0
		WHERE id = $2 AND restaurant_id = $3
	`

	cmdTag, err := tx.Exec(ctx, updateQuery, quantity, menuItemID, restaurantID)
	if err != nil {
		return fmt.Errorf("failed to update stock: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return errors.New("menu item not found")
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (s *PostgresStorage) CheckAvailability(restaurantID string, items []struct {
	ID       string
	Quantity int
},
) (bool, error) {
	ctx := context.Background()

	for _, item := range items {
		var hasStock bool
		query := `
			SELECT stock >= $1
			FROM menu_items
			WHERE id = $2 AND restaurant_id = $3
		`
		err := s.pool.QueryRow(ctx, query, item.Quantity, item.ID, restaurantID).Scan(&hasStock)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return false, nil // Блюдо не найдено, считаем его недоступным
			}
			return false, fmt.Errorf("failed to check availability: %w", err)
		}

		if !hasStock {
			return false, nil // Недостаточно товара
		}
	}

	return true, nil
}

func (s *PostgresStorage) restaurantExists(ctx context.Context, restaurantID string) (bool, error) {
	var exists bool

	query := `SELECT EXISTS(SELECT 1 FROM restaurants WHERE id = $1)`
	err := s.pool.QueryRow(ctx, query, restaurantID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check restaurant existence: %w", err)
	}

	return exists, nil
}

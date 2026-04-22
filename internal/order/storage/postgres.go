package storage

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seWy-bit/GO-and-eat/internal/order/domain"
)

type PostgresOrderStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresOrderStorage(pool *pgxpool.Pool) *PostgresOrderStorage {
	return &PostgresOrderStorage{pool: pool}
}

func (s *PostgresOrderStorage) CreateOrderWithTx(ctx context.Context, tx pgx.Tx, order domain.Order) error {
	orderQuery := `
		INSERT INTO orders (id, user_id, restaurant_id, status, total_amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err := tx.Exec(ctx, orderQuery,
		order.ID,
		order.UserID,
		order.RestaurantID,
		string(order.Status),
		order.TotalAmount,
		order.CreatedAt,
		order.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert order: %w", err)
	}

	itemQuery := `
		INSERT INTO order_items (id, order_id, menu_item_id, quantity, price)
		VALUES ($1, $2, $3, $4, $5)
	`

	for _, item := range order.Items {
		itemID := fmt.Sprintf("%s-%s", order.ID, item.MenuItemID)

		_, err := tx.Exec(ctx, itemQuery,
			itemID,
			order.ID,
			item.MenuItemID,
			item.Quantity,
			item.Price,
		)
		if err != nil {
			return fmt.Errorf("failed to insert order item %s: %w", item.MenuItemID, err)
		}
	}

	return nil
}

func (s *PostgresOrderStorage) GetOrder(ctx context.Context, id string) (domain.Order, error) {
	orderQuery := `
		SELECT id, user_id, restaurant_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE id = $1
	`

	var order domain.Order
	var statusStr string

	err := s.pool.QueryRow(ctx, orderQuery).Scan(
		&order.ID,
		&order.UserID,
		&order.RestaurantID,
		&statusStr,
		&order.TotalAmount,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Order{}, errors.New("order not found")
		}
		return domain.Order{}, fmt.Errorf("failed to get order: %w", err)
	}

	order.Status = domain.OrderStatus(statusStr)

	itemsQuery := `
		SELECT menu_item_id, name, quantity, price
		FROM order_items
		WHERE order_id = $1
	`

	rows, err := s.pool.Query(ctx, itemsQuery, id)
	if err != nil {
		return domain.Order{}, fmt.Errorf("failed to get order items: %w", err)
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		err := rows.Scan(
			&item.MenuItemID,
			&item.Name,
			&item.Quantity,
			&item.Price,
		)
		if err != nil {
			return domain.Order{}, fmt.Errorf("failed to scan order item: %w", err)
		}
		items = append(items, item)
	}

	if err = rows.Err(); err != nil {
		return domain.Order{}, fmt.Errorf("rows iteration error: %w", err)
	}

	order.Items = items
	return order, nil
}

func (s *PostgresOrderStorage) UpdateOrderStatus(ctx context.Context, id string, newStatus domain.OrderStatus) error {
	var currentStatusStr string
	query := `SELECT status FROM orders WHERE id = $1`
	err := s.pool.QueryRow(ctx, query, id).Scan(&currentStatusStr)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return errors.New("order not found")
		}
		return fmt.Errorf("failed to get order status: %w", err)
	}

	currentStatus := domain.OrderStatus(currentStatusStr)
	if !currentStatus.CanTransitionTo(newStatus) {
		return fmt.Errorf("invalid status transition: %s -> %s", currentStatus, newStatus)
	}

	updateQuery := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err = s.pool.Exec(ctx, updateQuery, string(newStatus), id)
	if err != nil {
		return fmt.Errorf("failed to update order status: %w", err)
	}

	return nil
}

func (s *PostgresOrderStorage) GetOrdersByUser(ctx context.Context, userID string) ([]domain.Order, error) {
	ordersQuery := `
		SELECT id, user_id, restaurant_id, status, total_amount, created_at, updated_at
		FROM orders
		WHERE user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := s.pool.Query(ctx, ordersQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user orders: %w", err)
	}
	defer rows.Close()

	var orders []domain.Order
	for rows.Next() {
		var order domain.Order
		var statusStr string

		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.RestaurantID,
			&statusStr,
			&order.TotalAmount,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan order: %w", err)
		}
		order.Status = domain.OrderStatus(statusStr)

		itemsQuery := `
			SELECT menu_item_id, name, quantity, price
			FROM order_items
			WHERE order_id = $1
		`
		itemRows, err := s.pool.Query(ctx, itemsQuery, order.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get order items: %w", err)
		}

		var items []domain.OrderItem
		for itemRows.Next() {
			var item domain.OrderItem
			err := itemRows.Scan(
				&item.MenuItemID,
				&item.Name,
				&item.Quantity,
				&item.Price,
			)
			if err != nil {
				itemRows.Close()
				return nil, fmt.Errorf("failed to scan order item: %w", err)
			}
			items = append(items, item)
		}
		itemRows.Close()
		order.Items = items
		orders = append(orders, order)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %w", err)
	}

	if orders == nil {
		return []domain.Order{}, nil
	}

	return orders, nil
}

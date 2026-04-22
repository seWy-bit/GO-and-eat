package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/seWy-bit/GO-and-eat/internal/order/domain"
	"github.com/seWy-bit/GO-and-eat/internal/order/storage"
	restaurantDomain "github.com/seWy-bit/GO-and-eat/internal/restaurant/domain"
	restaurantStorage "github.com/seWy-bit/GO-and-eat/internal/restaurant/storage"
)

type CreateOrderInput struct {
	ID           string
	UserID       string
	RestaurantID string
	Items        []OrderItemInput
}

type OrderItemInput struct {
	MenuItemID string
	Quantity   int
}

type CreateOrderUseCase struct {
	orderStorage      *storage.PostgresOrderStorage
	restaurantStorage *restaurantStorage.PostgresStorage
	db                *pgxpool.Pool
}

func NewCreateOrderUseCase(
	orderStorage *storage.PostgresOrderStorage,
	restaurantStorage *restaurantStorage.PostgresStorage,
	db *pgxpool.Pool,
) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		orderStorage:      orderStorage,
		restaurantStorage: restaurantStorage,
		db:                db,
	}
}

func (uc *CreateOrderUseCase) Execute(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {
	tx, err := uc.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var checkItems []struct {
		ID       string
		Quantity int
	}
	for _, item := range input.Items {
		checkItems = append(checkItems, struct {
			ID       string
			Quantity int
		}{
			ID:       item.MenuItemID,
			Quantity: item.Quantity,
		})
	}

	available, err := uc.restaurantStorage.CheckAvailabilityWithTx(ctx, tx, input.RestaurantID, checkItems)
	if err != nil {
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}
	if !available {
		return nil, errors.New("not enough stock for one or more items")
	}

	for _, item := range input.Items {
		err = uc.restaurantStorage.DecreaseStockWithTx(ctx, tx, input.RestaurantID, item.MenuItemID, item.Quantity)
		if err != nil {
			return nil, fmt.Errorf("failed to decrease stock for %s: %w", item.MenuItemID, err)
		}
	}

	menu, err := uc.restaurantStorage.GetMenu(input.RestaurantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get menu: %w", err)
	}

	menuMap := make(map[string]restaurantDomain.MenuItem)
	for _, item := range menu {
		menuMap[item.ID] = item
	}

	var orderItems []domain.OrderItem
	var totalAmount int64

	for _, reqItem := range input.Items {
		menuItem, exists := menuMap[reqItem.MenuItemID]
		if !exists {
			return nil, errors.New("menu item not found: " + reqItem.MenuItemID)
		}

		orderItems = append(orderItems, domain.OrderItem{
			MenuItemID: menuItem.ID,
			Name:       menuItem.Name,
			Quantity:   reqItem.Quantity,
			Price:      menuItem.Price,
		})

		totalAmount += int64(reqItem.Quantity) * menuItem.Price
	}

	now := time.Now()
	order := domain.Order{
		ID:           input.ID,
		UserID:       input.UserID,
		RestaurantID: input.RestaurantID,
		Items:        orderItems,
		TotalAmount:  totalAmount,
		Status:       domain.OrderStatusCreated,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.orderStorage.CreateOrderWithTx(ctx, tx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transcation: %w", err)
	}

	return &order, nil
}

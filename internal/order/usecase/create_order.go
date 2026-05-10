package usecase

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/seWy-bit/GO-and-eat/internal/order/domain"
	restaurantDomain "github.com/seWy-bit/GO-and-eat/internal/restaurant/domain"
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
	orderCreator   OrderCreator
	menuGetter     MenuGetter
	stockChecker   StockChecker
	stockDecreaser StockDecreaser
	txManager      TransactionManager
}

func NewCreateOrderUseCase(
	orderCreator OrderCreator,
	menuGetter MenuGetter,
	stockChecker StockChecker,
	stockDecreaser StockDecreaser,
	txManager TransactionManager,
) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		orderCreator:   orderCreator,
		menuGetter:     menuGetter,
		stockChecker:   stockChecker,
		stockDecreaser: stockDecreaser,
		txManager:      txManager,
	}
}

func (uc *CreateOrderUseCase) Execute(ctx context.Context, input CreateOrderInput) (*domain.Order, error) {
	if input.ID == "" {
		return nil, errors.New("id is required")
	}
	if input.UserID == "" {
		return nil, errors.New("user_id is required")
	}
	if input.RestaurantID == "" {
		return nil, errors.New("restaurant_id is required")
	}
	if len(input.Items) == 0 {
		return nil, errors.New("items are required")
	}

	tx, err := uc.txManager.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer uc.txManager.Rollback(tx)

	menu, err := uc.menuGetter.GetMenu(input.RestaurantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get menu: %w", err)
	}

	menuMap := make(map[string]restaurantDomain.MenuItem)
	for _, item := range menu {
		menuMap[item.ID] = item
	}

	checkItems := make([]struct {
		ID       string
		Quantity int
	}, len(input.Items))
	for i, item := range input.Items {
		checkItems[i] = struct {
			ID       string
			Quantity int
		}{
			ID:       item.MenuItemID,
			Quantity: item.Quantity,
		}
	}

	available, err := uc.stockChecker.CheckAvailabilityWithTx(ctx, tx, input.RestaurantID, checkItems)
	if err != nil {
		return nil, fmt.Errorf("failed to check availability: %w", err)
	}
	if !available {
		return nil, errors.New("not enough stock for one or more items")
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

		if err := uc.stockDecreaser.DecreaseStockWithTx(ctx, tx, input.RestaurantID, reqItem.MenuItemID, reqItem.Quantity); err != nil {
			return nil, fmt.Errorf("failed to decrease stock for %s: %w", reqItem.MenuItemID, err)
		}
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

	if err := uc.orderCreator.CreateOrderWithTx(ctx, tx, order); err != nil {
		return nil, fmt.Errorf("failed to save order: %w", err)
	}

	if err = uc.txManager.Commit(tx); err != nil {
		return nil, fmt.Errorf("failed to commit transcation: %w", err)
	}

	return &order, nil
}

package usecase

import (
	"errors"
	"time"

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
	orderStorage      *storage.MemoryOrderStorage
	restaurantStorage *restaurantStorage.MemoryStorage
}

func NewCreateOrderUseCase(
	orderStorage *storage.MemoryOrderStorage,
	restaurantStorage *restaurantStorage.MemoryStorage,
) *CreateOrderUseCase {
	return &CreateOrderUseCase{
		orderStorage:      orderStorage,
		restaurantStorage: restaurantStorage,
	}
}

func (uc *CreateOrderUseCase) Execute(input CreateOrderInput) (*domain.Order, error) {
	menu, err := uc.restaurantStorage.GetMenu(input.RestaurantID)
	if err != nil {
		return nil, errors.New("restaurant not found")
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

		if menuItem.Stock < reqItem.Quantity {
			return nil, errors.New("not enough stock for: " + menuItem.Name)
		}

		orderItems = append(orderItems, domain.OrderItem{
			MenuItemID: menuItem.ID,
			Name:       menuItem.Name,
			Price:      menuItem.Price,
			Quantity:   reqItem.Quantity,
		})

		totalAmount += int64(reqItem.Quantity) * menuItem.Price
	}

	order := domain.Order{
		ID:           input.ID,
		UserID:       input.UserID,
		RestaurantID: input.RestaurantID,
		Items:        orderItems,
		TotalAmount:  totalAmount,
		Status:       domain.OrderStatusCreated,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := uc.orderStorage.CreateOrder(order); err != nil {
		return nil, err
	}

	for _, reqItem := range input.Items {
		if err := uc.restaurantStorage.DecreaseStock(
			input.RestaurantID,
			reqItem.MenuItemID,
			reqItem.Quantity,
		); err != nil {
			return nil, err
		}
	}

	return &order, nil
}

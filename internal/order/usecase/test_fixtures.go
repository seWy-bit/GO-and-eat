package usecase

import (
	"time"

	"github.com/seWy-bit/GO-and-eat/internal/order/domain"
	restaurantDomain "github.com/seWy-bit/GO-and-eat/internal/restaurant/domain"
)

// Тестовый заказ
func testOrder() domain.Order {
	return domain.Order{
		ID:           "order-123",
		UserID:       "user-123",
		RestaurantID: "rest-123",
		Status:       domain.OrderStatusCreated,
		TotalAmount:  119800,
		Items: []domain.OrderItem{
			{
				MenuItemID: "pizza-1",
				Name:       "Маргарита",
				Quantity:   2,
				Price:      59900,
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func testMenu() []restaurantDomain.MenuItem {
	return []restaurantDomain.MenuItem{
		{
			ID:    "pizza-1",
			Name:  "Маргарита",
			Price: 59900,
			Stock: 10,
		},
		{
			ID:    "pizza-2",
			Name:  "Пепперони",
			Price: 69900,
			Stock: 5,
		},
	}
}

func testCreateOrderInput() CreateOrderInput {
	return CreateOrderInput{
		ID:           "order-123",
		UserID:       "user-123",
		RestaurantID: "rest-123",
		Items: []OrderItemInput{
			{MenuItemID: "pizza-1", Quantity: 2},
		},
	}
}

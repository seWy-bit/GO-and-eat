package domain

import "time"

type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "created"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID           string
	UserID       string
	RestaurantID string
	Items        []OrderItem
	Status       OrderStatus
	TotalAmount  int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type OrderItem struct {
	MenuItemID string
	Name       string
	Quantity   int
	Price      int64
}

func (o *Order) CalculateTotal() int64 {
	var total int64
	for _, item := range o.Items {
		total += item.Price * int64(item.Quantity)
	}

	return total
}

func (os OrderStatus) CanTransitionTo(newStatus OrderStatus) bool {
	transitions := map[OrderStatus][]OrderStatus{
		OrderStatusCreated:   {OrderStatusConfirmed, OrderStatusCancelled},
		OrderStatusConfirmed: {OrderStatusCancelled},
		OrderStatusCancelled: {},
	}

	for _, allowed := range transitions[os] {
		if allowed == newStatus {
			return true
		}
	}

	return false
}

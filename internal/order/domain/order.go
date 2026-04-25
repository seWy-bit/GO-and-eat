package domain

import "time"

type OrderStatus string

const (
	OrderStatusCreated   OrderStatus = "created"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusCooking   OrderStatus = "cooking"
	OrderStatusReady     OrderStatus = "ready"
	OrderStatusDelivered OrderStatus = "delivering"
	OrderStatusCompleted OrderStatus = "completed"
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

func (o *Order) CanTransitionTo(newStatus OrderStatus) bool {
	if o.Status == newStatus {
		return true
	}

	if o.IsFinal() {
		return false
	}

	transitions := map[OrderStatus][]OrderStatus{
		OrderStatusCreated:   {OrderStatusConfirmed, OrderStatusCancelled},
		OrderStatusConfirmed: {OrderStatusCancelled, OrderStatusCooking},
		OrderStatusCooking:   {OrderStatusCancelled, OrderStatusReady},
		OrderStatusReady:     {OrderStatusDelivered, OrderStatusCancelled},
		OrderStatusDelivered: {OrderStatusCompleted},
	}

	allowedStatuses, exists := transitions[o.Status]
	if !exists {
		return false
	}

	for _, allowed := range allowedStatuses {
		if allowed == newStatus {
			return true
		}
	}

	return false
}

func (o *Order) IsFinal() bool {
	return o.Status == OrderStatusCompleted || o.Status == OrderStatusCancelled
}

func (o *Order) CanBeCancelled() bool {
	return !o.IsFinal()
}

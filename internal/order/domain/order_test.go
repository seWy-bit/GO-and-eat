package domain

import (
	"testing"
	"time"
)

func createTestOrder(items []OrderItem) Order {
	return Order{
		ID:           "test-order-1",
		UserID:       "test-user-1",
		RestaurantID: "test-rest-1",
		Items:        items,
		Status:       OrderStatusCreated,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// ТЕСТ 1: CalculateTotal
func TestOrder_CalculateTotal(t *testing.T) {
	tests := []struct {
		name     string
		items    []OrderItem
		expected int64
	}{
		{
			name:     "empty order should have zero total",
			items:    []OrderItem{},
			expected: 0,
		},
		{
			name: "single item with quantity 1",
			items: []OrderItem{
				{Price: 10000, Quantity: 1},
			},
			expected: 10000,
		},
		{
			name: "single item with quantity 3",
			items: []OrderItem{
				{Price: 10000, Quantity: 3},
			},
			expected: 30000,
		},
		{
			name: "multiple items with different prices",
			items: []OrderItem{
				{Price: 10000, Quantity: 2},
				{Price: 5000, Quantity: 1},
				{Price: 7500, Quantity: 4},
			},
			expected: 10000*2 + 5000*1 + 7500*4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := createTestOrder(tt.items)
			result := order.CalculateTotal()
			if result != tt.expected {
				t.Errorf("CalculateTotal() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// ТЕСТ 2: CanTransitionTo
func TestOrder_CanTransitionTo(t *testing.T) {
	tests := []struct {
		name     string
		from     OrderStatus
		to       OrderStatus
		expected bool
	}{
		{"created -> confirmed", OrderStatusCreated, OrderStatusConfirmed, true},
		{"created -> cancelled", OrderStatusCreated, OrderStatusCancelled, true},

		{"confirmed -> cooking", OrderStatusConfirmed, OrderStatusCooking, true},
		{"confirmed -> cancelled", OrderStatusConfirmed, OrderStatusCancelled, true},

		{"cooking -> ready", OrderStatusCooking, OrderStatusReady, true},
		{"cooking -> cancelled", OrderStatusCooking, OrderStatusCancelled, true},

		{"ready -> delivered", OrderStatusReady, OrderStatusDelivered, true},
		{"ready -> cancelled", OrderStatusReady, OrderStatusCancelled, true},

		{"delivered -> completed", OrderStatusDelivered, OrderStatusCompleted, true},

		{"created -> created", OrderStatusCreated, OrderStatusCreated, true},
		{"confirmed -> confirmed", OrderStatusConfirmed, OrderStatusConfirmed, true},
		{"completed -> completed", OrderStatusCompleted, OrderStatusCompleted, true},
		{"cancelled -> cancelled", OrderStatusCancelled, OrderStatusCancelled, true},

		{"created -> cooking", OrderStatusCreated, OrderStatusCooking, false},
		{"created -> ready", OrderStatusCreated, OrderStatusReady, false},
		{"created -> delivered", OrderStatusCreated, OrderStatusDelivered, false},
		{"created -> completed", OrderStatusCreated, OrderStatusCompleted, false},
		{"confirmed -> ready", OrderStatusConfirmed, OrderStatusReady, false},
		{"confirmed -> delivered", OrderStatusConfirmed, OrderStatusDelivered, false},
		{"confirmed -> completed", OrderStatusConfirmed, OrderStatusCompleted, false},
		{"cooking -> delivered", OrderStatusCooking, OrderStatusDelivered, false},
		{"cooking -> completed", OrderStatusCooking, OrderStatusCompleted, false},
		{"ready -> completed", OrderStatusReady, OrderStatusCompleted, false},

		{"completed -> created", OrderStatusCompleted, OrderStatusCreated, false},
		{"completed -> confirmed", OrderStatusCompleted, OrderStatusConfirmed, false},
		{"completed -> cancelled", OrderStatusCompleted, OrderStatusCancelled, false},
		{"cancelled -> created", OrderStatusCancelled, OrderStatusCreated, false},
		{"cancelled -> confirmed", OrderStatusCancelled, OrderStatusConfirmed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			order := createTestOrder(nil)
			order.Status = tt.from
			result := order.CanTransitionTo(tt.to)
			if result != tt.expected {
				t.Errorf("CanTransitionTo(%s -> %s) = %v, want %v", tt.from, tt.to, result, tt.expected)
			}
		})
	}
}

// ТЕСТ 3: IsFinal
func TestOrder_IsFinal(t *testing.T) {
	tests := []struct {
		status   OrderStatus
		expected bool
	}{
		{OrderStatusCreated, false},
		{OrderStatusConfirmed, false},
		{OrderStatusCooking, false},
		{OrderStatusReady, false},
		{OrderStatusDelivered, false},
		{OrderStatusCompleted, true},
		{OrderStatusCancelled, true},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			order := createTestOrder(nil)
			order.Status = tt.status
			result := order.IsFinal()
			if result != tt.expected {
				t.Errorf("IsFinal() for status %s = %v, want %v",
					tt.status, result, tt.expected)
			}
		})
	}
}

// ТЕСТ 4: CanBeCancelled
func TestOrder_CanBeCancelled(t *testing.T) {
	tests := []struct {
		status   OrderStatus
		expected bool
	}{
		{OrderStatusCreated, true},
		{OrderStatusConfirmed, true},
		{OrderStatusCooking, true},
		{OrderStatusReady, true},
		{OrderStatusDelivered, true},
		{OrderStatusCompleted, false},
		{OrderStatusCancelled, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			order := createTestOrder(nil)
			order.Status = tt.status
			result := order.CanBeCancelled()
			if result != tt.expected {
				t.Errorf("CanBeCancelled() for status %s = %v, want %v",
					tt.status, result, tt.expected)
			}
		})
	}
}

// ТЕСТ 5: Полная цепочка переходов
func TestOrder_TransitionChain(t *testing.T) {
	order := createTestOrder(nil)
	order.Status = OrderStatusCreated

	// Правильная цепочка переходов
	transitions := []OrderStatus{
		OrderStatusConfirmed,
		OrderStatusCooking,
		OrderStatusReady,
		OrderStatusDelivered,
		OrderStatusCompleted,
	}

	for _, nextStatus := range transitions {
		t.Run("transition to "+string(nextStatus), func(t *testing.T) {
			if !order.CanTransitionTo(nextStatus) {
				t.Errorf("Cannot transition from %s to %s", order.Status, nextStatus)
			}
			order.Status = nextStatus
		})
	}
}

func TestOrder_InvalidTransitionChain(t *testing.T) {
	order := createTestOrder(nil)
	order.Status = OrderStatusCreated

	// Невозможная цепочка (прыжок через статусы)
	if order.CanTransitionTo(OrderStatusDelivered) {
		t.Error("Should not be able to jump from created to delivered")
	}
}

package storage

import (
	"errors"
	"sync"

	"github.com/seWy-bit/GO-and-eat/internal/order/domain"
)

type MemoryOrderStorage struct {
	mu     sync.RWMutex
	orders map[string]domain.Order
}

func NewMemoryOrderStorage() *MemoryOrderStorage {
	return &MemoryOrderStorage{
		orders: make(map[string]domain.Order),
	}
}

func (s *MemoryOrderStorage) CreateOrder(order domain.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.orders[order.ID]; exists {
		return errors.New("order already exists")
	}

	s.orders[order.ID] = order
	return nil
}

func (s *MemoryOrderStorage) GetOrder(id string) (domain.Order, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	order, exists := s.orders[id]
	if !exists {
		return domain.Order{}, errors.New("order not found")
	}
	return order, nil
}

func (s *MemoryOrderStorage) UpdateOrder(order domain.Order) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.orders[order.ID]; !exists {
		return errors.New("order not found")
	}

	s.orders[order.ID] = order
	return nil
}

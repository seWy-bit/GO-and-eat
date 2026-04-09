package storage

import (
	"errors"
	"sync"

	"github.com/seWy-bit/GO-and-eat/internal/restaurant/domain"
)

type MemoryStorage struct {
	mu          sync.RWMutex
	restaurants map[string]domain.Restaurant
	menuItems   map[string][]domain.MenuItem
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		restaurants: make(map[string]domain.Restaurant),
		menuItems:   make(map[string][]domain.MenuItem),
	}
}

func (s *MemoryStorage) CreateRestaurant(r domain.Restaurant) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.restaurants[r.ID]; exists {
		return errors.New("restaurant already exists")
	}

	s.restaurants[r.ID] = r
	return nil
}

func (s *MemoryStorage) GetRestaurant(id string) (domain.Restaurant, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	r, exists := s.restaurants[id]
	if !exists {
		return domain.Restaurant{}, errors.New("restaurant not found")
	}
	return r, nil
}

func (s *MemoryStorage) AddMenuItem(item domain.MenuItem) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.restaurants[item.RestaurantID]; !exists {
		return errors.New("restaurant not found")
	}

	s.menuItems[item.RestaurantID] = append(s.menuItems[item.RestaurantID], item)
	return nil
}

func (s *MemoryStorage) GetMenu(restaurantID string) ([]domain.MenuItem, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.restaurants[restaurantID]; !exists {
		return nil, errors.New("restaurant not found")
	}

	items, exists := s.menuItems[restaurantID]
	if !exists {
		return []domain.MenuItem{}, nil // Возвращаем пустой список, если меню не найдено
	}

	return items, nil
}

func (s *MemoryStorage) DecreaseStock(restaurantID, menuItemID string, quantity int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.restaurants[restaurantID]; !exists {
		return errors.New("restaurant not found")
	}

	items, exists := s.menuItems[restaurantID]
	if !exists {
		return errors.New("menu not found")
	}

	for i := range items {
		if items[i].ID == menuItemID {
			if items[i].Stock < quantity {
				return errors.New("not enough stock")
			}
			items[i].Stock -= quantity
			s.menuItems[restaurantID] = items
			return nil
		}
	}
	return errors.New("menu item not found")
}

func (s *MemoryStorage) CheckAvailability(restaurantID string, items []struct {
	ID       string
	Quantity int
},
) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.restaurants[restaurantID]; !exists {
		return false, errors.New("restaurant not found")
	}

	menu, exists := s.menuItems[restaurantID]
	if !exists {
		return false, errors.New("menu not found")
	}

	for _, reqItem := range items {
		found := false
		for _, menuItem := range menu {
			if menuItem.ID == reqItem.ID {
				if menuItem.Stock < reqItem.Quantity {
					return false, nil // Недостаточно товара
				}
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}
	return true, nil
}

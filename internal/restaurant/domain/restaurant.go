package domain

import "time"

// Представляет ресторан в системе
type Restaurant struct {
	ID        string
	Name      string
	Address   string
	Phone     string
	CreatedAt time.Time
}

// Представляет блюдо в меню ресторана
type MenuItem struct {
	ID           string
	RestaurantID string
	Name         string
	Description  string
	Price        int64
	Stock        int
	Available    bool
	CreatedAt    time.Time
}

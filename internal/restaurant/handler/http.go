package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/seWy-bit/GO-and-eat/internal/restaurant/domain"
	"github.com/seWy-bit/GO-and-eat/internal/restaurant/storage"
)

type RestaurantHandler struct {
	storage *storage.MemoryStorage
}

type CreateRestaurantRequest struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Phone   string `json:"phone"`
}

type AddMenuItemRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Price       int64  `json:"price"`
	Stock       int    `json:"stock"`
	Available   bool   `json:"available"`
}

func NewRestaurantHandler(storage *storage.MemoryStorage) *RestaurantHandler {
	return &RestaurantHandler{
		storage: storage,
	}
}

func (h *RestaurantHandler) CreateRestaurant(w http.ResponseWriter, r *http.Request) {
	var req CreateRestaurantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ID == "" || req.Name == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	restaurant := domain.Restaurant{
		ID:        req.ID,
		Name:      req.Name,
		Address:   req.Address,
		Phone:     req.Phone,
		CreatedAt: time.Now(),
	}

	if err := h.storage.CreateRestaurant(restaurant); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(restaurant)
}

func (h *RestaurantHandler) GetMenu(w http.ResponseWriter, r *http.Request) {
	restaurantID := r.PathValue("id")

	if restaurantID == "" {
		http.Error(w, "Missing restaurant ID", http.StatusBadRequest)
		return
	}

	menu, err := h.storage.GetMenu(restaurantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(menu)
}

func (h *RestaurantHandler) AddMenuItem(w http.ResponseWriter, r *http.Request) {
	restaurantID := r.PathValue("id")

	if restaurantID == "" {
		http.Error(w, "Missing restaurant ID", http.StatusBadRequest)
		return
	}

	var req AddMenuItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.ID == "" || req.Name == "" || req.Price <= 0 {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	item := domain.MenuItem{
		ID:           req.ID,
		RestaurantID: restaurantID,
		Name:         req.Name,
		Description:  req.Description,
		Price:        req.Price,
		Stock:        req.Stock,
		Available:    req.Available,
		CreatedAt:    time.Now(),
	}

	if err := h.storage.AddMenuItem(item); err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(item)
}

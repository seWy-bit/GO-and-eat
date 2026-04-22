package handler

import (
	"encoding/json"
	"net/http"

	"github.com/seWy-bit/GO-and-eat/internal/order/usecase"
)

type OrderHandler struct {
	createOrderUseCase *usecase.CreateOrderUseCase
}

func NewOrderHandler(createOrderUseCase *usecase.CreateOrderUseCase) *OrderHandler {
	return &OrderHandler{
		createOrderUseCase: createOrderUseCase,
	}
}

type CreateOrderRequest struct {
	ID           string `json:"id"`
	UserID       string `json:"user_id"`
	RestaurantID string `json:"restaurant_id"`
	Items        []struct {
		MenuItemID string `json:"menu_item_id"`
		Quantity   int    `json:"quantity"`
	} `json:"items"`
}

type CreateOrderResponse struct {
	OrderID     string `json:"order_id"`
	Status      string `json:"status"`
	TotalAmount int64  `json:"total_amount"`
}

func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if req.ID == "" || req.UserID == "" || req.RestaurantID == "" || len(req.Items) == 0 {
		http.Error(w, "missing required fields", http.StatusBadRequest)
		return
	}

	input := usecase.CreateOrderInput{
		ID:           req.ID,
		UserID:       req.UserID,
		RestaurantID: req.RestaurantID,
		Items:        make([]usecase.OrderItemInput, len(req.Items)),
	}

	for i, item := range req.Items {
		input.Items[i] = usecase.OrderItemInput{
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
		}
	}

	order, err := h.createOrderUseCase.Execute(ctx, input)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	response := CreateOrderResponse{
		OrderID:     order.ID,
		Status:      string(order.Status),
		TotalAmount: order.TotalAmount,
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

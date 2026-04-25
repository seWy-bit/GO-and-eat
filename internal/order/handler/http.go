package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/seWy-bit/GO-and-eat/internal/order/domain"
	"github.com/seWy-bit/GO-and-eat/internal/order/usecase"
)

type OrderHandler struct {
	createOrderUseCase       *usecase.CreateOrderUseCase
	getOrderUseCase          *usecase.GetOrderUseCase
	updateOrderStatusUseCase *usecase.UpdateOrderStatusUseCase
	getUserOrdersUseCase     *usecase.GetUserOrdersUseCase
}

func NewOrderHandler(
	createOrderUseCase *usecase.CreateOrderUseCase,
	getOrderUseCase *usecase.GetOrderUseCase,
	updateOrderStatusUseCase *usecase.UpdateOrderStatusUseCase,
	getUserOrdersUseCase *usecase.GetUserOrdersUseCase,
) *OrderHandler {
	return &OrderHandler{
		createOrderUseCase:       createOrderUseCase,
		getOrderUseCase:          getOrderUseCase,
		updateOrderStatusUseCase: updateOrderStatusUseCase,
		getUserOrdersUseCase:     getUserOrdersUseCase,
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

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "order id is required", http.StatusBadRequest)
		return
	}

	order, err := h.getOrderUseCase.Execute(r.Context(), id)
	if err != nil {
		if err.Error() == "order not found" {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		http.Error(w, "internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(order)
}

func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		http.Error(w, "order id is required", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	newStatus := domain.OrderStatus(req.Status)

	validStatuses := map[domain.OrderStatus]bool{
		domain.OrderStatusCreated:   true,
		domain.OrderStatusConfirmed: true,
		domain.OrderStatusCooking:   true,
		domain.OrderStatusReady:     true,
		domain.OrderStatusDelivered: true,
		domain.OrderStatusCompleted: true,
		domain.OrderStatusCancelled: true,
	}
	if !validStatuses[newStatus] {
		http.Error(w, "invalid status", http.StatusBadRequest)
		return
	}

	err := h.updateOrderStatusUseCase.Execute(r.Context(), id, newStatus)
	if err != nil {
		if err.Error() == "order not found" {
			http.Error(w, "order not found", http.StatusNotFound)
			return
		}
		if strings.Contains(err.Error(), "invalid status transition") {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *OrderHandler) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	userID := r.PathValue("user_id")
	if userID == "" {
		http.Error(w, "user id is required", http.StatusBadRequest)
		return
	}

	orders, err := h.getUserOrdersUseCase.Execute(r.Context(), userID)
	if err != nil {
		http.Error(w, "internal server error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(orders)
}

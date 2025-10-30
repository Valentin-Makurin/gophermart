package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Valentin-Makurin/gophermart/internal/middleware"
	"github.com/Valentin-Makurin/gophermart/internal/models"
	repository "github.com/Valentin-Makurin/gophermart/internal/repo"
	// "github.com/Valentin-Makurin/gophermart/internal/service"
)

type OrdersHandler struct {
	orderService repository.OrderService
}

func NewOrdersHandler(orderService repository.OrderService) *OrdersHandler {
	return &OrdersHandler{
		orderService: orderService,
	}
}

func (h *OrdersHandler) UploadOrder(w http.ResponseWriter, r *http.Request) {

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderNumber := string(body)
	if orderNumber == "" {
		http.Error(w, "Empty order number", http.StatusBadRequest)
		return
	}

	err = h.orderService.UploadOrder(r.Context(), userID, orderNumber)
	if err != nil {
		switch err {
		case models.ErrInvalidOrderNumber:
			http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
		case models.ErrOrderAlreadyUploadedByUser:
			w.WriteHeader(http.StatusOK)
		case models.ErrOrderAlreadyUploadedByOther:
			http.Error(w, "Order already uploaded by other user", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusAccepted)
}

func (h *OrdersHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	orders, err := h.orderService.GetUserOrders(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

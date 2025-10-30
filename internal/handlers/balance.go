package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Valentin-Makurin/gophermart/internal/middleware"
	"github.com/Valentin-Makurin/gophermart/internal/models"
	repository "github.com/Valentin-Makurin/gophermart/internal/repo"
	// "github.com/Valentin-Makurin/gophermart/internal/service"
)

type BalanceHandler struct {
	balanceService repository.BalanceService
}

func NewBalanceHandler(balanceService repository.BalanceService) *BalanceHandler {
	return &BalanceHandler{
		balanceService: balanceService,
	}
}

func (h *BalanceHandler) GetBalance(w http.ResponseWriter, r *http.Request) {

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	balance, err := h.balanceService.GetBalance(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(balance); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

func (h *BalanceHandler) Withdraw(w http.ResponseWriter, r *http.Request) {

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req models.WithdrawalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request format", http.StatusBadRequest)
		return
	}

	if req.Order == "" || req.Sum <= 0 {
		http.Error(w, "Invalid order or sum", http.StatusBadRequest)
		return
	}

	err = h.balanceService.Withdraw(r.Context(), userID, &req)
	if err != nil {
		switch err {
		case models.ErrInvalidOrderNumber:
			http.Error(w, "Invalid order number", http.StatusUnprocessableEntity)
		case models.ErrInsufficientFunds:
			http.Error(w, "Insufficient funds", http.StatusPaymentRequired)
		case models.ErrOrderAlreadyUsedForWithdrawal:
			http.Error(w, "Order already used for withdrawal", http.StatusConflict)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *BalanceHandler) GetWithdrawals(w http.ResponseWriter, r *http.Request) {

	userID, err := middleware.GetUserIDFromContext(r.Context())
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := h.balanceService.GetWithdrawals(r.Context(), userID)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	if len(withdrawals) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(withdrawals); err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}

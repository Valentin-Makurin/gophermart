package models

import "time"

type Balance struct {
	Current   float64 `json:"current"`
	Withdrawn float64 `json:"withdrawn"`
}

type WithdrawalRequest struct {
	Order string  `json:"order"`
	Sum   float64 `json:"sum"`
}

type Withdrawal struct {
	Order       string    `json:"order"`
	Sum         float64   `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

type BalanceOperation struct {
	ID            int64     `json:"-"`
	UserID        int64     `json:"-"`
	OrderNumber   string    `json:"order,omitempty"`
	Sum           float64   `json:"sum"`
	OperationType string    `json:"-"`
	ProcessedAt   time.Time `json:"processed_at"`
}

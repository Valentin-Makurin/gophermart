package repository

import (
	"context"

	"github.com/Valentin-Makurin/gophermart/internal/auth"
	"github.com/Valentin-Makurin/gophermart/internal/models"
)

type UserRepository interface {
	CreateUser(ctx context.Context, user *models.User) error
	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
	// GetUserByID(ctx context.Context, id int64) (*models.User, error)
}

type OrderRepository interface {
	CreateOrder(ctx context.Context, order *models.Order) error
	GetOrderByNumber(ctx context.Context, number string) (*models.Order, error)
	GetUserOrders(ctx context.Context, userID int64) ([]*models.Order, error)
	UpdateOrderStatus(ctx context.Context, orderNumber string, status string, accrual *float64) error
	GetOrdersByStatuses(ctx context.Context, statuses []string) ([]*models.Order, error)
}

type BalanceRepository interface {
	GetUserBalance(ctx context.Context, userID int64) (*models.Balance, error)
	CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error
	GetUserWithdrawals(ctx context.Context, userID int64) ([]*models.Withdrawal, error)
	GetAccrualsSum(ctx context.Context, userID int64) (float64, error)
	GetWithdrawalsSum(ctx context.Context, userID int64) (float64, error)
	GetWithdrawalByOrder(ctx context.Context, orderNumber string) (*models.Withdrawal, error)
	CreateAccrual(ctx context.Context, userID int64, orderNumber string, sum float64) error
	GetAccrualByOrder(ctx context.Context, orderNumber string) (*models.BalanceOperation, error)
}

type AuthService interface {
	Register(ctx context.Context, req *models.UserRegisterRequest) (*models.User, error)
	Login(ctx context.Context, req *models.UserLoginRequest) (*models.User, error)
}

type JWTManager interface {
	GenerateToken(userID int64) (string, error)
	ValidateToken(tokenString string) (*auth.Claims, error)
}

type OrderService interface {
	GetUserOrders(ctx context.Context, userID int64) ([]*models.Order, error)
	UploadOrder(ctx context.Context, userID int64, orderNumber string) error
}

type BalanceService interface {
	GetBalance(ctx context.Context, userID int64) (*models.Balance, error)
	Withdraw(ctx context.Context, userID int64, req *models.WithdrawalRequest) error
	GetWithdrawals(ctx context.Context, userID int64) ([]*models.Withdrawal, error)
}

type Client interface {
	GetOrderInfo(ctx context.Context, orderNumber string) (*models.AccrualOrder, error)
}

package repository

// import (
// 	"context"

// 	"github.com/Valentin-Makurin/gophermart/internal/models"
// )

// type UserRepository interface {
// 	CreateUser(ctx context.Context, user *models.User) error
// 	GetUserByLogin(ctx context.Context, login string) (*models.User, error)
// 	GetUserByID(ctx context.Context, id int64) (*models.User, error)
// }

// type OrderRepository interface {
// 	CreateOrder(ctx context.Context, order *models.Order) error
// 	GetOrderByNumber(ctx context.Context, number string) (*models.Order, error)
// 	GetUserOrders(ctx context.Context, userID int64) ([]*models.Order, error)
// 	UpdateOrderStatus(ctx context.Context, orderNumber string, status string, accrual *float64) error
// }

// type BalanceRepository interface {
// 	GetUserBalance(ctx context.Context, userID int64) (*models.Balance, error)
// 	CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error
// 	GetUserWithdrawals(ctx context.Context, userID int64) ([]*models.Withdrawal, error)
// 	GetAccrualsSum(ctx context.Context, userID int64) (float64, error)
// 	GetWithdrawalsSum(ctx context.Context, userID int64) (float64, error)
// 	GetWithdrawalByOrder(ctx context.Context, orderNumber string) (*models.Withdrawal, error)
// }

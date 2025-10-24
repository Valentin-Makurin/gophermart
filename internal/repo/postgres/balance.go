package postgres

import (
	"context"
	"errors"

	"github.com/Valentin-Makurin/gophermart/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BalanceRepository struct {
	db *pgxpool.Pool
}

func NewBalanceRepository(db *pgxpool.Pool) *BalanceRepository {
	return &BalanceRepository{db: db}
}

func (r *BalanceRepository) GetUserBalance(ctx context.Context, userID int64) (*models.Balance, error) {
	query := `
		SELECT 
			COALESCE(SUM(CASE WHEN ot.code = 'ACCRUAL' THEN bo.sum ELSE 0 END), 0) as current,
			COALESCE(SUM(CASE WHEN ot.code = 'WITHDRAWAL' THEN bo.sum ELSE 0 END), 0) as withdrawn
		FROM mvv.balance_operations bo
		JOIN mvv.operation_types ot 
		  ON bo.operation_type_id = ot.id
		WHERE bo.user_id = $1`

	var balance models.Balance
	err := r.db.QueryRow(ctx, query, userID).Scan(&balance.Current, &balance.Withdrawn)
	if err != nil {
		return nil, err
	}

	return &balance, nil
}

func (r *BalanceRepository) CreateWithdrawal(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	query := `
		INSERT INTO mvv.balance_operations (user_id, order_number, sum, operation_type_id)
		VALUES ($1, $2, $3, (SELECT id FROM mvv.operation_types WHERE code = 'WITHDRAWAL'))`

	_, err := r.db.Exec(ctx, query, userID, orderNumber, sum)
	return err
}

func (r *BalanceRepository) GetUserWithdrawals(ctx context.Context, userID int64) ([]*models.Withdrawal, error) {
	query := `
		SELECT bo.order_number,
		       bo.sum,
		       bo.processed_at
		FROM mvv.balance_operations bo
		JOIN mvv.operation_types ot 
		  ON bo.operation_type_id = ot.id
		WHERE bo.user_id = $1 
		  AND ot.code = 'WITHDRAWAL'
		ORDER BY bo.processed_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var withdrawals []*models.Withdrawal
	for rows.Next() {
		var withdrawal models.Withdrawal
		err := rows.Scan(&withdrawal.Order, &withdrawal.Sum, &withdrawal.ProcessedAt)
		if err != nil {
			return nil, err
		}
		withdrawals = append(withdrawals, &withdrawal)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return withdrawals, nil
}

func (r *BalanceRepository) GetAccrualsSum(ctx context.Context, userID int64) (float64, error) {
	query := `
		SELECT COALESCE(SUM(bo.sum), 0)
		FROM mvv.balance_operations bo
		JOIN mvv.operation_types ot 
		  ON bo.operation_type_id = ot.id
		WHERE bo.user_id = $1 
		  AND ot.code = 'ACCRUAL'`

	var sum float64
	err := r.db.QueryRow(ctx, query, userID).Scan(&sum)
	if err != nil {
		return 0, err
	}

	return sum, nil
}

func (r *BalanceRepository) GetWithdrawalsSum(ctx context.Context, userID int64) (float64, error) {
	query := `
		SELECT COALESCE(SUM(bo.sum), 0)
		FROM mvv.balance_operations bo
		JOIN mvv.operation_types ot 
		  ON bo.operation_type_id = ot.id
		WHERE bo.user_id = $1 
		  AND ot.code = 'WITHDRAWAL'`

	var sum float64
	err := r.db.QueryRow(ctx, query, userID).Scan(&sum)
	if err != nil {
		return 0, err
	}

	return sum, nil
}

func (r *BalanceRepository) GetWithdrawalByOrder(ctx context.Context, orderNumber string) (*models.Withdrawal, error) {
	query := `
		SELECT bo.order_number,
		       bo.sum,
		       bo.processed_at
		FROM mvv.balance_operations bo
		JOIN mvv.operation_types ot 
		  ON bo.operation_type_id = ot.id
		WHERE bo.order_number = $1 
		  AND ot.code = 'WITHDRAWAL'`

	var withdrawal models.Withdrawal
	err := r.db.QueryRow(ctx, query, orderNumber).Scan(
		&withdrawal.Order,
		&withdrawal.Sum,
		&withdrawal.ProcessedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrWithdrawalNotFound
		}
		return nil, err
	}

	return &withdrawal, nil
}

func (r *BalanceRepository) GetAccrualByOrder(ctx context.Context, orderNumber string) (*models.BalanceOperation, error) {
	query := `
        SELECT bo.id,
		       bo.user_id,
		       bo.order_number,
		       bo.sum,
		       bo.processed_at
        FROM mvv.balance_operations bo
        JOIN mvv.operation_types ot 
		  ON bo.operation_type_id = ot.id
        WHERE bo.order_number = $1 
		  AND ot.code = 'ACCRUAL'`

	var operation models.BalanceOperation
	err := r.db.QueryRow(ctx, query, orderNumber).Scan(
		&operation.ID,
		&operation.UserID,
		&operation.OrderNumber,
		&operation.Sum,
		&operation.ProcessedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrAccrualNotFound
		}
		return nil, err
	}

	return &operation, nil
}

func (r *BalanceRepository) CreateAccrual(ctx context.Context, userID int64, orderNumber string, sum float64) error {
	query := `
        INSERT INTO mvv.balance_operations (user_id, order_number, sum, operation_type_id)
        VALUES ($1, $2, $3, (SELECT id FROM mvv.operation_types WHERE code = 'ACCRUAL'))`

	_, err := r.db.Exec(ctx, query, userID, orderNumber, sum)
	return err
}

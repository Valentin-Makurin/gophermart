package postgres

import (
	"context"
	"errors"

	"github.com/Valentin-Makurin/gophermart/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) CreateOrder(ctx context.Context, order *models.Order) error {
	query := `
		INSERT INTO mvv.orders (user_id, number, status_id) 
		VALUES ($1, $2, (SELECT id FROM mvv.order_statuses WHERE code = 'NEW'))
		RETURNING id, uploaded_at`

	err := r.db.QueryRow(
		ctx,
		query,
		order.UserID,
		order.Number,
	).Scan(&order.ID, &order.UploadedAt)

	if err != nil {
		if isUniqueViolation(err) {
			return models.ErrOrderAlreadyExists
		}
		return err
	}

	return nil
}

func (r *OrderRepository) GetOrderByNumber(ctx context.Context, number string) (*models.Order, error) {
	query := `
		SELECT o.id, 
			   o.user_id, 
			   o.number, 
			   os.code as status, 
			   o.accrual, 
			   o.uploaded_at
		FROM mvv.orders o
		JOIN mvv.order_statuses os 
		  ON o.status_id = os.id
		WHERE o.number = $1`

	var order models.Order
	err := r.db.QueryRow(ctx, query, number).Scan(
		&order.ID,
		&order.UserID,
		&order.Number,
		&order.Status,
		&order.Accrual,
		&order.UploadedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, models.ErrOrderNotFound
		}
		return nil, err
	}

	return &order, nil
}

func (r *OrderRepository) GetUserOrders(ctx context.Context, userID int64) ([]*models.Order, error) {
	query := `
		SELECT o.id, 
		       o.user_id, 
		       o.number, 
		       os.code as status, 
		       o.accrual, 
		       o.uploaded_at
		FROM mvv.orders o
		JOIN mvv.order_statuses os 
		  ON o.status_id = os.id
		WHERE o.user_id = $1
		ORDER BY o.uploaded_at DESC`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return orders, nil
}

func (r *OrderRepository) UpdateOrderStatus(ctx context.Context, orderNumber string, status string, accrual *float64) error {
	query := `
		UPDATE mvv.orders 
		SET status_id = (SELECT id FROM mvv.order_statuses WHERE code = $1), 
		    accrual = $2, 
		    updated_at = NOW()
		WHERE number = $3`

	result, err := r.db.Exec(ctx, query, status, accrual, orderNumber)
	if err != nil {
		return err
	}

	if result.RowsAffected() == 0 {
		return models.ErrOrderNotFound
	}

	return nil
}

func (r *OrderRepository) GetOrdersByStatuses(ctx context.Context, statuses []string) ([]*models.Order, error) {
	query := `
		SELECT o.id, 
		       o.user_id, 
		       o.number, 
		       os.code as status, 
		       o.accrual, 
		       o.uploaded_at
		FROM mvv.orders o
		JOIN mvv.order_statuses os 
		  ON o.status_id = os.id
		WHERE os.code = ANY($1)
		ORDER BY o.uploaded_at ASC
		LIMIT 100`

	rows, err := r.db.Query(ctx, query, statuses)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*models.Order
	for rows.Next() {
		var order models.Order
		err := rows.Scan(
			&order.ID,
			&order.UserID,
			&order.Number,
			&order.Status,
			&order.Accrual,
			&order.UploadedAt,
		)
		if err != nil {
			return nil, err
		}
		orders = append(orders, &order)
	}

	return orders, nil
}

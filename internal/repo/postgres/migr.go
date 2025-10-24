package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

const timeout time.Duration = 5 * time.Second

type MigrRepository struct {
	db *pgxpool.Pool
}

func NewMigrRepository(db *pgxpool.Pool) *MigrRepository {
	return &MigrRepository{db: db}
}

func (m *MigrRepository) RunMigrations() error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	Migration := `
	CREATE SCHEMA IF NOT EXISTS mvv;

	CREATE TABLE IF NOT EXISTS mvv.order_statuses (
		id SERIAL PRIMARY KEY,
		code TEXT UNIQUE NOT NULL,
		description TEXT
	);
	
	CREATE TABLE IF NOT EXISTS mvv.operation_types (
		id SERIAL PRIMARY KEY,
		code TEXT UNIQUE NOT null,
		description TEXT
	);
	
	CREATE TABLE IF NOT EXISTS mvv.users (
		id BIGSERIAL PRIMARY KEY,
		login TEXT UNIQUE NOT NULL,
		password_hash TEXT NOT NULL,
		created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
	);
	
	CREATE TABLE IF NOT EXISTS mvv.orders (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		number TEXT UNIQUE NOT NULL,
		status_id INTEGER NOT NULL,
		accrual DECIMAL(10,2) DEFAULT 0,
		uploaded_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		CONSTRAINT fk_orders_user_id FOREIGN KEY (user_id) REFERENCES mvv.users (id) ON DELETE CASCADE,
		CONSTRAINT fk_orders_status_id FOREIGN KEY (status_id) REFERENCES mvv.order_statuses (id)    
	);
	
	
	CREATE TABLE IF NOT EXISTS mvv.balance_operations (
		id BIGSERIAL PRIMARY KEY,
		user_id BIGINT NOT NULL,
		order_number TEXT NOT NULL,
		sum DECIMAL(10,2) NOT NULL,
		operation_type_id INTEGER NOT NULL,
		processed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
		CONSTRAINT fk_balance_operations_user_id FOREIGN KEY (user_id) REFERENCES mvv.users(id) ON DELETE CASCADE,
		CONSTRAINT fk_balance_operations_operation_type_id FOREIGN KEY (operation_type_id) REFERENCES mvv.operation_types(id)    
	);
	
	INSERT INTO mvv.order_statuses (code, description) VALUES
	('NEW', 'Заказ загружен в систему, но не попал в обработку'),
	('PROCESSING', 'Вознаграждение за заказ рассчитывается'),
	('INVALID', 'Система расчёта вознаграждений отказала в расчёте'),
	('PROCESSED',  'Данные по заказу проверены и информация о расчёте успешно получена')
	ON CONFLICT (code) DO NOTHING;
	
	INSERT INTO mvv.operation_types (code, description) VALUES
	('ACCRUAL', 'Начисление баллов лояльности за заказ'),
	('WITHDRAWAL', 'Списание баллов для оплаты заказа')
	ON CONFLICT (code) DO NOTHING;
	`

	_, err := m.db.Exec(ctx, Migration)
	if err != nil {
		return err
	}
	return nil
}

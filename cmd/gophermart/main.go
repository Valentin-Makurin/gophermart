package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"database/sql"

	"github.com/Valentin-Makurin/gophermart/internal/auth"
	"github.com/Valentin-Makurin/gophermart/internal/config"
	"github.com/Valentin-Makurin/gophermart/internal/handlers"
	"github.com/Valentin-Makurin/gophermart/internal/middleware"
	"github.com/Valentin-Makurin/gophermart/internal/repo/postgres"
	"github.com/Valentin-Makurin/gophermart/internal/service"
	"github.com/Valentin-Makurin/gophermart/internal/service/accrual"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	// "github.com/jackc/pgx"
	// "embed"

	"github.com/jackc/pgx/v5/pgxpool"
	// "github.com/pressly/goose/v3"
	"go.uber.org/zap"

	"github.com/golang-migrate/migrate/v4"
	mgPostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Ошибка инициализации logger:", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	cfg := config.ParseFlagsServer(sugar)

	db, err := pgxpool.New(context.Background(), cfg.DatabaseURI)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer db.Close()

	if err := db.Ping(context.Background()); err != nil {
		sugar.Error("Unable to ping database: %v", err)
	}

	if err := runMigrations(cfg.DatabaseURI); err != nil {
		sugar.Error("Failed to run migrations: %v", err)
	}

	// migrRepo := postgres.NewMigrRepository(db)
	// err = migrRepo.RunMigrations()
	// if err != nil {
	// 	sugar.Error("Failed to run migr %v", err)
	// }

	userRepo := postgres.NewUserRepository(db)
	orderRepo := postgres.NewOrderRepository(db)
	balanceRepo := postgres.NewBalanceRepository(db)

	authService := service.NewAuthService(userRepo)
	jwtManager := auth.NewJWTManager(cfg.JWTKey)

	accrualClient := accrual.NewClient(cfg.AccrualSystemAddress)
	orderService := service.NewOrderService(orderRepo, balanceRepo, accrualClient)
	balanceService := service.NewBalanceService(balanceRepo)

	authHandler := handlers.NewAuthHandler(authService, jwtManager)
	ordersHandler := handlers.NewOrdersHandler(orderService)
	balanceHandler := handlers.NewBalanceHandler(balanceService)

	r := chi.NewRouter()

	r.Use(middleware.LoggerMiddleware(sugar))
	r.Use(middleware.Compress)

	r.Group(func(r chi.Router) {
		authHandler.RegisterRoutes(r)
	})

	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(jwtManager))

		r.Post("/api/user/orders", ordersHandler.UploadOrder)
		r.Get("/api/user/orders", ordersHandler.GetOrders)

		r.Get("/api/user/balance", balanceHandler.GetBalance)
		r.Post("/api/user/balance/withdraw", balanceHandler.Withdraw)
		r.Get("/api/user/withdrawals", balanceHandler.GetWithdrawals)
	})

	server := &http.Server{
		Addr:    cfg.RunAddress,
		Handler: r,
	}

	if cfg.AccrualSystemAddress != "" {
		worker := accrual.NewWorker(orderRepo, balanceRepo, accrualClient, 5*time.Second)
		go worker.Start(context.Background())
		sugar.Error("Accrual worker started for: %s", cfg.AccrualSystemAddress)
	} else {
		sugar.Error("Accrual system address not provided")
	}

	sugar.Error("Running server on", cfg.RunAddress)
	err = server.ListenAndServe()
	if err != nil {
		sugar.Error("Filed to start server", err)
	}

}

// func runMigrations(databaseURI string) error {
// 	// var migrationsFS embed.FS

// 	db, err := sql.Open("pgx", databaseURI)
// 	if err != nil {
// 		return err
// 	}

// 	defer db.Close()

// 	// if _, err := os.Stat("migrations"); os.IsNotExist(err) {
// 	// 	return fmt.Errorf("migrations directory does not exist at: %s. Current working directory: %s",
// 	// 		"migrations", getWorkingDir())
// 	// }

// 	// Проверяем соединение
// 	if err := db.Ping(); err != nil {
// 		return fmt.Errorf("failed to ping database: %w", err)
// 	}
// 	// goose.SetBaseFS(migrationsFS)

// 	// Устанавливаем диалект
// 	if err := goose.SetDialect("postgres"); err != nil {
// 		return fmt.Errorf("failed to set dialect: %w", err)
// 	}

// 	// Выполняем миграции
// 	if err := goose.Up(db, "migrations"); err != nil {
// 		return fmt.Errorf("failed to run migrations: %w", err)
// 	}

// 	// log.Println("Database migrations applied successfully")
// 	return nil
// }

// // func getWorkingDir() string {
// // 	dir, _ := os.Getwd()
// // 	return dir
// // }

func runMigrations(databaseURI string) error {
	db, err := sql.Open("pgx", databaseURI)
	if err != nil {
		return fmt.Errorf("failed to connect for migrations: %w", err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	driver, err := mgPostgres.WithInstance(db, &mgPostgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	migrationsPath, err := getMigrationsPath()
	if err != nil {
		return fmt.Errorf("failed to get migrations path: %w", err)
	}

	log.Printf("Migrations path: %s", migrationsPath)

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+migrationsPath,
		"postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migration instance: %w", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Println("Database migrations applied successfully")
	return nil
}

func getMigrationsPath() (string, error) {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("failed to get current file path")
	}

	currentDir := filepath.Dir(filename)

	migrationsPath := filepath.Join(currentDir, "migrations")

	if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
		// Пробуем найти на уровень выше (для структуры cmd/gophermart/)
		parentDir := filepath.Dir(currentDir)
		migrationsPath = filepath.Join(parentDir, "migrations")

		if _, err := os.Stat(migrationsPath); os.IsNotExist(err) {
			return "", fmt.Errorf("migrations directory not found in: %s or %s",
				filepath.Join(currentDir, "migrations"), migrationsPath)
		}
	}

	return migrationsPath, nil
}

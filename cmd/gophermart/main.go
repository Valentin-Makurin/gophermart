package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/Valentin-Makurin/gophermart/internal/auth"
	"github.com/Valentin-Makurin/gophermart/internal/config"
	"github.com/Valentin-Makurin/gophermart/internal/handlers"
	"github.com/Valentin-Makurin/gophermart/internal/middleware"
	"github.com/Valentin-Makurin/gophermart/internal/repo/postgres"
	"github.com/Valentin-Makurin/gophermart/internal/service"
	"github.com/Valentin-Makurin/gophermart/internal/service/accrual"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
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

	migrRepo := postgres.NewMigrRepository(db)
	err = migrRepo.RunMigrations()
	if err != nil {
		sugar.Error("Failed to run migr %v", err)
	}

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

package accrual

import (
	"context"
	"log"
	"time"

	"github.com/Valentin-Makurin/gophermart/internal/common"
	"github.com/Valentin-Makurin/gophermart/internal/models"
	"github.com/Valentin-Makurin/gophermart/internal/repo/postgres"
)

type Worker struct {
	orderRepo     *postgres.OrderRepository
	balanceRepo   *postgres.BalanceRepository
	accrualClient *Client
	interval      time.Duration
}

func NewWorker(orderRepo *postgres.OrderRepository, balanceRepo *postgres.BalanceRepository, accrualClient *Client, interval time.Duration) *Worker {
	return &Worker{
		orderRepo:     orderRepo,
		balanceRepo:   balanceRepo,
		accrualClient: accrualClient,
		interval:      interval,
	}
}

func (w *Worker) Start(ctx context.Context) {
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	log.Println("Accrual worker started")

	for {
		select {
		case <-ctx.Done():
			log.Println("Accrual worker stopped")
			return
		case <-ticker.C:
			w.processPendingOrders(ctx)
		}
	}
}

func (w *Worker) processPendingOrders(ctx context.Context) {
	orders, err := w.orderRepo.GetOrdersByStatuses(ctx, []string{"NEW", "PROCESSING"})
	if err != nil {
		log.Printf("Error getting pending orders: %v", err)
		return
	}

	for _, order := range orders {
		accrualOrder, err := w.accrualClient.GetOrderInfo(ctx, order.Number)
		if err != nil {
			if rateLimitErr, ok := err.(*RateLimitError); ok {
				time.Sleep(rateLimitErr.RetryAfter)
				return
			}
			log.Printf("Error getting order info for %s: %v", order.Number, err)
			continue
		}

		mappedStatus := common.MapAccrualStatus(accrualOrder.Status)

		err = w.orderRepo.UpdateOrderStatus(ctx, order.Number, mappedStatus, accrualOrder.Accrual)
		if err != nil {
			log.Printf("Error updating order %s: %v", order.Number, err)
			continue
		}

		if mappedStatus == "PROCESSED" && accrualOrder.Accrual != nil && *accrualOrder.Accrual > 0 {

			existingAccrual, err := w.balanceRepo.GetAccrualByOrder(ctx, order.Number)
			if err != nil && err != models.ErrAccrualNotFound {
				log.Printf("Error checking existing accrual for order %s: %v", order.Number, err)
				continue
			}

			if existingAccrual == nil {

				err = w.balanceRepo.CreateAccrual(ctx, order.UserID, order.Number, *accrualOrder.Accrual)
				if err != nil {
					log.Printf("Error creating accrual for order %s: %v", order.Number, err)
				} else {
					log.Printf("Accrual created for order %s: %.2f points", order.Number, *accrualOrder.Accrual)
				}
			} else {
				log.Printf("Accrual already exists for order %s: %.2f points", order.Number, existingAccrual.Sum)
			}
		}

		log.Printf("Order %s updated to status: %s", order.Number, mappedStatus)
	}
}

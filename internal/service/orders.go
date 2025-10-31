package service

import (
	"context"
	"log"
	"strconv"

	"github.com/Valentin-Makurin/gophermart/internal/common"
	"github.com/Valentin-Makurin/gophermart/internal/models"
	repository "github.com/Valentin-Makurin/gophermart/internal/repo"
)

type OrderService struct {
	orderRepo     repository.OrderRepository
	balanceRepo   repository.BalanceRepository
	accrualClient repository.Client
}

func NewOrderService(orderRepo repository.OrderRepository, balanceRepo repository.BalanceRepository, accrualClient repository.Client) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		balanceRepo:   balanceRepo,
		accrualClient: accrualClient,
	}
}

func (s *OrderService) GetUserOrders(ctx context.Context, userID int64) ([]*models.Order, error) {
	return s.orderRepo.GetUserOrders(ctx, userID)
}

func isValidLuhn(number string) bool {
	sum := 0
	isSecond := false

	for i := len(number) - 1; i >= 0; i-- {
		digit, err := strconv.Atoi(string(number[i]))
		if err != nil {
			return false
		}

		if isSecond {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isSecond = !isSecond
	}

	return sum%10 == 0
}

func (s *OrderService) UploadOrder(ctx context.Context, userID int64, orderNumber string) error {
	if !isValidLuhn(orderNumber) {
		return models.ErrInvalidOrderNumber
	}

	existingOrder, err := s.orderRepo.GetOrderByNumber(ctx, orderNumber)
	if err != nil && err != models.ErrOrderNotFound {
		return err
	}

	if existingOrder != nil {
		if existingOrder.UserID == userID {
			return models.ErrOrderAlreadyUploadedByUser
		}
		return models.ErrOrderAlreadyUploadedByOther
	}

	accrualOrder, err := s.accrualClient.GetOrderInfo(ctx, orderNumber)
	if err != nil {
		log.Printf("Order %s not found in accrual system: %v", orderNumber, err)
	}

	order := &models.Order{
		UserID: userID,
		Number: orderNumber,
	}

	if accrualOrder != nil {
		order.Status = common.MapAccrualStatus(accrualOrder.Status)
		order.Accrual = accrualOrder.Accrual
	} else {
		order.Status = "NEW"
	}

	err = s.orderRepo.CreateOrder(ctx, order)
	if err != nil {
		return err
	}

	if order.Status == "PROCESSED" && order.Accrual != nil && *order.Accrual > 0 {
		err = s.balanceRepo.CreateAccrual(ctx, userID, orderNumber, *order.Accrual)
		if err != nil {
			log.Printf("Error creating immediate accrual for order %s: %v", orderNumber, err)
		} else {
			log.Printf("Immediate accrual created for order %s: %.2f points", orderNumber, *order.Accrual)
		}
	}

	return nil
}

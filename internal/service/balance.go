package service

import (
	"context"

	"github.com/Valentin-Makurin/gophermart/internal/models"
	repository "github.com/Valentin-Makurin/gophermart/internal/repo"
)

type BalanceService struct {
	balanceRepo repository.BalanceRepository
}

func NewBalanceService(balanceRepo repository.BalanceRepository) *BalanceService {
	return &BalanceService{
		balanceRepo: balanceRepo,
	}
}

func (s *BalanceService) GetBalance(ctx context.Context, userID int64) (*models.Balance, error) {
	balance, err := s.balanceRepo.GetUserBalance(ctx, userID)
	if err == nil {
		balance.Current = balance.Current - balance.Withdrawn
	}
	return balance, err
}

func (s *BalanceService) Withdraw(ctx context.Context, userID int64, req *models.WithdrawalRequest) error {

	if !isValidLuhn(req.Order) {
		return models.ErrInvalidOrderNumber
	}

	balance, err := s.balanceRepo.GetUserBalance(ctx, userID)
	if err != nil {
		return err
	}

	if balance.Current < req.Sum {
		return models.ErrInsufficientFunds
	}

	existingWithdrawal, err := s.balanceRepo.GetWithdrawalByOrder(ctx, req.Order)
	if err != nil && err != models.ErrWithdrawalNotFound {
		return err
	}

	if existingWithdrawal != nil {
		return models.ErrOrderAlreadyUsedForWithdrawal
	}

	return s.balanceRepo.CreateWithdrawal(ctx, userID, req.Order, req.Sum)
}

func (s *BalanceService) GetWithdrawals(ctx context.Context, userID int64) ([]*models.Withdrawal, error) {
	return s.balanceRepo.GetUserWithdrawals(ctx, userID)
}

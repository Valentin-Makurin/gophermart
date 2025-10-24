package service

import (
	"context"
	"github.com/Valentin-Makurin/gophermart/internal/models"
	"github.com/Valentin-Makurin/gophermart/internal/repo/postgres"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	userRepo *postgres.UserRepository
}

func NewAuthService(userRepo *postgres.UserRepository) *AuthService {
	return &AuthService{
		userRepo: userRepo,
	}
}

func (s *AuthService) Register(ctx context.Context, req *models.UserRegisterRequest) (*models.User, error) {

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Login:        req.Login,
		PasswordHash: string(passwordHash),
	}

	err = s.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(ctx context.Context, req *models.UserLoginRequest) (*models.User, error) {

	user, err := s.userRepo.GetUserByLogin(ctx, req.Login)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password))
	if err != nil {
		return nil, models.ErrInvalidPassword
	}

	return user, nil
}

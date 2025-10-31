package models

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrOrderAlreadyExists          = errors.New("order already exists")
	ErrOrderNotFound               = errors.New("order not found")
	ErrInvalidOrderNumber          = errors.New("invalid order number")
	ErrOrderAlreadyUploadedByUser  = errors.New("order already uploaded by user")
	ErrOrderAlreadyUploadedByOther = errors.New("order already uploaded by other user")

	ErrInsufficientFunds             = errors.New("insufficient funds")
	ErrWithdrawalNotFound            = errors.New("withdrawal not found")
	ErrOrderAlreadyUsedForWithdrawal = errors.New("order already used for withdrawal")
	ErrAccrualNotFound               = errors.New("accrual not found")
)

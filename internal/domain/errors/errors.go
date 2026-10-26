package errors

import "errors"

var (
	ErrUserNotFound        = errors.New("user not found")
	ErrUserAlreadyExists   = errors.New("email already registered")
	ErrUsernameTaken       = errors.New("username already taken")
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrUnauthorized        = errors.New("unauthorized")
	ErrInvalidInput        = errors.New("invalid input")
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrPaymentFailed       = errors.New("payment failed")
	ErrServiceUnavailable  = errors.New("service temporarily unavailable")
)

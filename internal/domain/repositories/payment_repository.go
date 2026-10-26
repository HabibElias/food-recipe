package repositories

import (
	"context"

	"go/food-recipe/internal/domain/entities"
)

type PaymentRepository interface {
	CreateTransaction(ctx context.Context, tx *entities.Transaction) error
	GetTransactionByRef(ctx context.Context, txRef string) (*entities.Transaction, error)
	CreatePayment(ctx context.Context, payment *entities.Payment) error
}

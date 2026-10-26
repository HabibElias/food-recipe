package hasura

import (
	"context"

	"github.com/google/uuid"
	"go/food-recipe/internal/domain/entities"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/domain/repositories"
)

type paymentRepository struct {
	client *Client
}

var _ repositories.PaymentRepository = (*paymentRepository)(nil)

func NewPaymentRepository(client *Client) repositories.PaymentRepository {
	return &paymentRepository{client: client}
}

func (r *paymentRepository) CreateTransaction(ctx context.Context, tx *entities.Transaction) error {
	if r.client == nil || r.client.GQL == nil {
		return domainerrors.ErrServiceUnavailable
	}
	if tx == nil {
		return domainerrors.ErrInvalidInput
	}

	var mutation struct {
		InsertTransaction struct {
			Returning []struct {
				ID uuid.UUID `graphql:"id"`
			} `graphql:"returning"`
		} `graphql:"insert_transaction(objects: {trx_ref: $trx_ref, user_id: $user_id, recipe_id: $recipe_id, currency: $currency, amount: $amount, checkout_url: $checkout_url})"`
	}

	variables := map[string]interface{}{
		"user_id":      GQLUUID(tx.UserID),
		"recipe_id":    tx.RecipeID,
		"trx_ref":      tx.TrxRef,
		"checkout_url": tx.CheckoutURL,
		"amount":       tx.Amount,
		"currency":     tx.Currency,
	}

	if err := r.client.GQL.Mutate(ctx, &mutation, variables); err != nil {
		return err
	}

	if len(mutation.InsertTransaction.Returning) > 0 {
		tx.ID = mutation.InsertTransaction.Returning[0].ID
	}

	return nil
}

func (r *paymentRepository) GetTransactionByRef(ctx context.Context, txRef string) (*entities.Transaction, error) {
	if r.client == nil || r.client.GQL == nil {
		return nil, domainerrors.ErrServiceUnavailable
	}

	var query struct {
		Transaction []struct {
			ID          uuid.UUID `graphql:"id"`
			TrxRef      string    `graphql:"trx_ref"`
			RecipeID    int       `graphql:"recipe_id"`
			UserID      uuid.UUID `graphql:"user_id"`
			Currency    string    `graphql:"currency"`
			CheckoutURL string    `graphql:"checkout_url"`
			Amount      float64   `graphql:"amount"`
		} `graphql:"transaction(where: {trx_ref: {_eq: $trx_ref}})"`
	}

	variables := map[string]interface{}{
		"trx_ref": txRef,
	}

	if err := r.client.GQL.Query(ctx, &query, variables); err != nil {
		return nil, err
	}

	if len(query.Transaction) == 0 {
		return nil, domainerrors.ErrTransactionNotFound
	}

	t := query.Transaction[0]
	return &entities.Transaction{
		ID:          t.ID,
		UserID:      t.UserID,
		RecipeID:    t.RecipeID,
		TrxRef:      t.TrxRef,
		Amount:      t.Amount,
		Currency:    t.Currency,
		CheckoutURL: t.CheckoutURL,
	}, nil
}

func (r *paymentRepository) CreatePayment(ctx context.Context, payment *entities.Payment) error {
	if r.client == nil || r.client.GQL == nil {
		return domainerrors.ErrServiceUnavailable
	}
	if payment == nil {
		return domainerrors.ErrInvalidInput
	}

	var mutation struct {
		InsertPayment struct {
			Returning []struct {
				ID string `graphql:"id"`
			} `graphql:"returning"`
		} `graphql:"insert_payment(objects: {user_id: $user_id, recipe_id: $recipe_id, tx_ref: $tx_ref, amount: $amount, chapa_checkout_url: $checkout_url, currency: $currency, status: $status})"`
	}

	variables := map[string]interface{}{
		"user_id":      GQLUUID(payment.UserID),
		"recipe_id":    payment.RecipeID,
		"tx_ref":       payment.TxRef,
		"amount":       payment.Amount,
		"checkout_url": payment.CheckoutURL,
		"currency":     payment.Currency,
		"status":       payment.Status,
	}

	if err := r.client.GQL.Mutate(ctx, &mutation, variables); err != nil {
		return err
	}

	if len(mutation.InsertPayment.Returning) > 0 {
		payment.ID = mutation.InsertPayment.Returning[0].ID
	}

	return nil
}

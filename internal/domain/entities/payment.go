package entities

import "github.com/google/uuid"

type Transaction struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	RecipeID    int
	TrxRef      string
	Amount      float64
	Currency    string
	CheckoutURL string
}

type Payment struct {
	ID          string
	UserID      uuid.UUID
	RecipeID    int
	TxRef       string
	Amount      float64
	CheckoutURL string
	Currency    string
	Status      string
}

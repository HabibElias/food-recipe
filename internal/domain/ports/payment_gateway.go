package ports

import "context"

type PaymentInitParams struct {
	Amount      float64
	Currency    string
	FirstName   string
	LastName    string
	Email       string
	TxRef       string
	Phone       string
	CallbackURL string
}

type PaymentInitResult struct {
	CheckoutURL string
	TxRef       string
	Status      string
}

type PaymentVerifyResult struct {
	Status  string
	Message string
}

type PaymentGateway interface {
	Initialize(ctx context.Context, params PaymentInitParams) (*PaymentInitResult, error)
	Verify(ctx context.Context, txRef string) (*PaymentVerifyResult, error)
}

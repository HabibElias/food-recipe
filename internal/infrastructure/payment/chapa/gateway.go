package chapa

import (
	"context"
	"fmt"

	chapa "github.com/Chapa-Et/chapa-go"
	"github.com/shopspring/decimal"
	"github.com/spf13/viper"
	"go/food-recipe/internal/config"
	"go/food-recipe/internal/domain/ports"
)

type ChapaGateway struct {
	client chapa.API
}

func NewChapaGateway(cfg config.ChapaConfig) *ChapaGateway {
	if cfg.SecretKey != "" {
		viper.Set("API_KEY", cfg.SecretKey)
	}
	return &ChapaGateway{
		client: chapa.New(),
	}
}

var _ ports.PaymentGateway = (*ChapaGateway)(nil)

func (g *ChapaGateway) Initialize(ctx context.Context, params ports.PaymentInitParams) (*ports.PaymentInitResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	req := &chapa.PaymentRequest{
		Amount:         decimal.NewFromFloat(params.Amount),
		Currency:       params.Currency,
		FirstName:      params.FirstName,
		LastName:       params.LastName,
		Email:          params.Email,
		TransactionRef: params.TxRef,
		Phone:          params.Phone,
		CallbackURL:    params.CallbackURL,
	}

	api := g.client
	res, err := api.PaymentRequest(req)
	if err != nil {
		return nil, fmt.Errorf("chapa request error: %w", err)
	}
	if res == nil {
		return nil, fmt.Errorf("empty response from chapa")
	}

	if res.Status != "success" {
		return nil, fmt.Errorf("chapa payment initialization failed: %s", res.Message)
	}

	return &ports.PaymentInitResult{
		CheckoutURL: res.Data.CheckoutURL,
		TxRef:       params.TxRef,
		Status:      res.Status,
	}, nil
}

func (g *ChapaGateway) Verify(ctx context.Context, txRef string) (*ports.PaymentVerifyResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	res, err := g.client.Verify(txRef)
	if err != nil {
		return nil, fmt.Errorf("chapa verify error: %w", err)
	}
	if res == nil {
		return nil, fmt.Errorf("empty response from chapa")
	}

	return &ports.PaymentVerifyResult{
		Status:  res.Status,
		Message: res.Message,
	}, nil
}

package chapa

import (
	"context"
	"errors"
	"testing"

	chapaSDK "github.com/Chapa-Et/chapa-go"
	"go/food-recipe/internal/config"
	"go/food-recipe/internal/domain/ports"
)

type mockChapaAPI struct {
	paymentResponse *chapaSDK.PaymentResponse
	paymentErr      error
	verifyResponse  *chapaSDK.VerifyResponse
	verifyErr       error
}

func (m *mockChapaAPI) PaymentRequest(request *chapaSDK.PaymentRequest) (*chapaSDK.PaymentResponse, error) {
	return m.paymentResponse, m.paymentErr
}

func (m *mockChapaAPI) Verify(txnRef string) (*chapaSDK.VerifyResponse, error) {
	return m.verifyResponse, m.verifyErr
}

func (m *mockChapaAPI) TransferToBank(request *chapaSDK.BankTransfer) (*chapaSDK.BankTransferResponse, error) {
	return nil, nil
}

func (m *mockChapaAPI) GetTransactions() (*chapaSDK.TransactionsResponse, error) {
	return nil, nil
}

func (m *mockChapaAPI) GetBanks() (*chapaSDK.BanksResponse, error) {
	return nil, nil
}

func (m *mockChapaAPI) BulkTransfer(*chapaSDK.BulkTransferRequest) (*chapaSDK.BulkTransferResponse, error) {
	return nil, nil
}

func TestNewChapaGateway(t *testing.T) {
	gw := NewChapaGateway(config.ChapaConfig{SecretKey: "test-secret-key"})
	if gw == nil {
		t.Fatal("expected non-nil ChapaGateway")
	}

	var _ ports.PaymentGateway = gw
}

func TestChapaGateway_Initialize(t *testing.T) {
	ctx := context.Background()
	params := ports.PaymentInitParams{
		Amount:      100.50,
		Currency:    "ETB",
		FirstName:   "Abebe",
		LastName:    "Bikila",
		Email:       "abebe@example.com",
		TxRef:       "tx-12345",
		Phone:       "0911000000",
		CallbackURL: "https://example.com/callback",
	}

	t.Run("success", func(t *testing.T) {
		mockAPI := &mockChapaAPI{
			paymentResponse: &chapaSDK.PaymentResponse{
				Status:  "success",
				Message: "Hosted Link",
				Data: struct {
					CheckoutURL string `json:"checkout_url"`
				}{
					CheckoutURL: "https://checkout.chapa.co/checkout/123",
				},
			},
		}
		gw := &ChapaGateway{client: mockAPI}

		res, err := gw.Initialize(ctx, params)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.CheckoutURL != "https://checkout.chapa.co/checkout/123" {
			t.Errorf("expected checkout URL %s, got %s", "https://checkout.chapa.co/checkout/123", res.CheckoutURL)
		}
		if res.TxRef != params.TxRef {
			t.Errorf("expected tx_ref %s, got %s", params.TxRef, res.TxRef)
		}
		if res.Status != "success" {
			t.Errorf("expected status 'success', got %s", res.Status)
		}
	})

	t.Run("api error", func(t *testing.T) {
		mockAPI := &mockChapaAPI{
			paymentErr: errors.New("network failure"),
		}
		gw := &ChapaGateway{client: mockAPI}

		_, err := gw.Initialize(ctx, params)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("status not success", func(t *testing.T) {
		mockAPI := &mockChapaAPI{
			paymentResponse: &chapaSDK.PaymentResponse{
				Status:  "failed",
				Message: "invalid card",
			},
		}
		gw := &ChapaGateway{client: mockAPI}

		_, err := gw.Initialize(ctx, params)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestChapaGateway_Verify(t *testing.T) {
	ctx := context.Background()

	t.Run("success", func(t *testing.T) {
		mockAPI := &mockChapaAPI{
			verifyResponse: &chapaSDK.VerifyResponse{
				Status:  "success",
				Message: "Payment successful",
			},
		}
		gw := &ChapaGateway{client: mockAPI}

		res, err := gw.Verify(ctx, "tx-12345")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if res.Status != "success" {
			t.Errorf("expected status success, got %s", res.Status)
		}
		if res.Message != "Payment successful" {
			t.Errorf("expected message 'Payment successful', got %s", res.Message)
		}
	})

	t.Run("api error", func(t *testing.T) {
		mockAPI := &mockChapaAPI{
			verifyErr: errors.New("verification timeout"),
		}
		gw := &ChapaGateway{client: mockAPI}

		_, err := gw.Verify(ctx, "tx-12345")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

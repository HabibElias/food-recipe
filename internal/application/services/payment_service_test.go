package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go/food-recipe/internal/application/dto"
	"go/food-recipe/internal/config"
	"go/food-recipe/internal/domain/entities"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/domain/ports"
	"go/food-recipe/internal/domain/repositories"
)

type mockPaymentRepo struct {
	createTxFn func(ctx context.Context, tx *entities.Transaction) error
	getTxFn     func(ctx context.Context, txRef string) (*entities.Transaction, error)
	createPayFn func(ctx context.Context, payment *entities.Payment) error
}

var _ repositories.PaymentRepository = (*mockPaymentRepo)(nil)

func (m *mockPaymentRepo) CreateTransaction(ctx context.Context, tx *entities.Transaction) error {
	if m.createTxFn != nil {
		return m.createTxFn(ctx, tx)
	}
	return nil
}

func (m *mockPaymentRepo) GetTransactionByRef(ctx context.Context, txRef string) (*entities.Transaction, error) {
	if m.getTxFn != nil {
		return m.getTxFn(ctx, txRef)
	}
	return nil, domainerrors.ErrTransactionNotFound
}

func (m *mockPaymentRepo) CreatePayment(ctx context.Context, payment *entities.Payment) error {
	if m.createPayFn != nil {
		return m.createPayFn(ctx, payment)
	}
	return nil
}

type mockPaymentGateway struct {
	initFn   func(ctx context.Context, params ports.PaymentInitParams) (*ports.PaymentInitResult, error)
	verifyFn func(ctx context.Context, txRef string) (*ports.PaymentVerifyResult, error)
}

var _ ports.PaymentGateway = (*mockPaymentGateway)(nil)

func (m *mockPaymentGateway) Initialize(ctx context.Context, params ports.PaymentInitParams) (*ports.PaymentInitResult, error) {
	if m.initFn != nil {
		return m.initFn(ctx, params)
	}
	return &ports.PaymentInitResult{
		CheckoutURL: "https://checkout.chapa.co/test",
		TxRef:       params.TxRef,
		Status:      "success",
	}, nil
}

func (m *mockPaymentGateway) Verify(ctx context.Context, txRef string) (*ports.PaymentVerifyResult, error) {
	if m.verifyFn != nil {
		return m.verifyFn(ctx, txRef)
	}
	return &ports.PaymentVerifyResult{
		Status:  "success",
		Message: "Payment successful",
	}, nil
}

func TestPaymentService_InitializePayment(t *testing.T) {
	cfg := &config.Config{
		App: config.AppConfig{
			CallbackURL: "https://api.example.com/",
		},
	}
	ctx := context.Background()

	userID := uuid.New()
	testUser := &entities.User{
		ID:        userID,
		Email:     "user@example.com",
		FirstName: "Jane",
		LastName:  "Doe",
	}

	t.Run("successful initialization with custom phone and trailing slash stripped", func(t *testing.T) {
		var capturedTx *entities.Transaction
		var capturedParams ports.PaymentInitParams

		payRepo := &mockPaymentRepo{
			createTxFn: func(ctx context.Context, tx *entities.Transaction) error {
				capturedTx = tx
				return nil
			},
		}

		userRepo := &mockUserRepo{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*entities.User, error) {
				if id == userID {
					return testUser, nil
				}
				return nil, domainerrors.ErrUserNotFound
			},
		}

		gw := &mockPaymentGateway{
			initFn: func(ctx context.Context, params ports.PaymentInitParams) (*ports.PaymentInitResult, error) {
				capturedParams = params
				return &ports.PaymentInitResult{
					CheckoutURL: "https://checkout.chapa.co/pay/123",
					TxRef:       params.TxRef,
					Status:      "success",
				}, nil
			},
		}

		svc := NewPaymentService(payRepo, userRepo, gw, cfg)
		output, err := svc.InitializePayment(ctx, dto.InitializePaymentInputDTO{
			Amount:   250.50,
			RecipeID: 99,
			Phone:    "0911998877",
		}, userID.String())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output.CheckoutURL != "https://checkout.chapa.co/pay/123" {
			t.Errorf("expected checkout URL, got %s", output.CheckoutURL)
		}
		if output.Status != "success" {
			t.Errorf("expected status success, got %s", output.Status)
		}
		if output.TxRef == "" {
			t.Error("expected non-empty txRef")
		}

		// Verify params passed to gateway
		if capturedParams.Amount != 250.50 {
			t.Errorf("expected amount 250.50, got %f", capturedParams.Amount)
		}
		if capturedParams.Email != "user@example.com" {
			t.Errorf("expected email user@example.com, got %s", capturedParams.Email)
		}
		if capturedParams.Phone != "0911998877" {
			t.Errorf("expected phone 0911998877, got %s", capturedParams.Phone)
		}
		if capturedParams.CallbackURL != "https://api.example.com/v1/payment/verify" {
			t.Errorf("expected callback URL without duplicate slash, got %s", capturedParams.CallbackURL)
		}

		// Verify transaction record
		if capturedTx == nil {
			t.Fatal("expected transaction to be saved")
		}
		if capturedTx.RecipeID != 99 || capturedTx.UserID != userID || capturedTx.Amount != 250.50 {
			t.Errorf("unexpected transaction saved: %+v", capturedTx)
		}
	})

	t.Run("default phone fallback when omitted", func(t *testing.T) {
		var capturedParams ports.PaymentInitParams

		userRepo := &mockUserRepo{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*entities.User, error) {
				return testUser, nil
			},
		}

		gw := &mockPaymentGateway{
			initFn: func(ctx context.Context, params ports.PaymentInitParams) (*ports.PaymentInitResult, error) {
				capturedParams = params
				return &ports.PaymentInitResult{
					Status: "success",
				}, nil
			},
		}

		svc := NewPaymentService(&mockPaymentRepo{}, userRepo, gw, cfg)
		_, err := svc.InitializePayment(ctx, dto.InitializePaymentInputDTO{
			Amount:   50,
			RecipeID: 1,
		}, userID.String())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if capturedParams.Phone != "0900123456" {
			t.Errorf("expected default phone 0900123456, got %s", capturedParams.Phone)
		}
	})

	t.Run("missing session user id", func(t *testing.T) {
		svc := NewPaymentService(&mockPaymentRepo{}, &mockUserRepo{}, &mockPaymentGateway{}, cfg)
		_, err := svc.InitializePayment(ctx, dto.InitializePaymentInputDTO{
			Amount:   100,
			RecipeID: 1,
		}, "")
		if !errors.Is(err, domainerrors.ErrUnauthorized) {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("invalid session user id", func(t *testing.T) {
		svc := NewPaymentService(&mockPaymentRepo{}, &mockUserRepo{}, &mockPaymentGateway{}, cfg)
		_, err := svc.InitializePayment(ctx, dto.InitializePaymentInputDTO{
			Amount:   100,
			RecipeID: 1,
		}, "not-a-valid-uuid")
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("invalid amount <= 0", func(t *testing.T) {
		svc := NewPaymentService(&mockPaymentRepo{}, &mockUserRepo{}, &mockPaymentGateway{}, cfg)
		_, err := svc.InitializePayment(ctx, dto.InitializePaymentInputDTO{
			Amount:   0,
			RecipeID: 1,
		}, userID.String())
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("invalid recipe id <= 0", func(t *testing.T) {
		svc := NewPaymentService(&mockPaymentRepo{}, &mockUserRepo{}, &mockPaymentGateway{}, cfg)
		_, err := svc.InitializePayment(ctx, dto.InitializePaymentInputDTO{
			Amount:   10,
			RecipeID: 0,
		}, userID.String())
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*entities.User, error) {
				return nil, domainerrors.ErrUserNotFound
			},
		}
		svc := NewPaymentService(&mockPaymentRepo{}, userRepo, &mockPaymentGateway{}, cfg)
		_, err := svc.InitializePayment(ctx, dto.InitializePaymentInputDTO{
			Amount:   10,
			RecipeID: 1,
		}, userID.String())
		if !errors.Is(err, domainerrors.ErrUserNotFound) {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})

	t.Run("gateway failure", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*entities.User, error) {
				return testUser, nil
			},
		}
		gw := &mockPaymentGateway{
			initFn: func(ctx context.Context, params ports.PaymentInitParams) (*ports.PaymentInitResult, error) {
				return nil, errors.New("network error")
			},
		}
		svc := NewPaymentService(&mockPaymentRepo{}, userRepo, gw, cfg)
		_, err := svc.InitializePayment(ctx, dto.InitializePaymentInputDTO{
			Amount:   10,
			RecipeID: 1,
		}, userID.String())
		if !errors.Is(err, domainerrors.ErrPaymentFailed) {
			t.Errorf("expected ErrPaymentFailed, got %v", err)
		}
	})

	t.Run("gateway status not success", func(t *testing.T) {
		userRepo := &mockUserRepo{
			getByIDFn: func(ctx context.Context, id uuid.UUID) (*entities.User, error) {
				return testUser, nil
			},
		}
		gw := &mockPaymentGateway{
			initFn: func(ctx context.Context, params ports.PaymentInitParams) (*ports.PaymentInitResult, error) {
				return &ports.PaymentInitResult{
					Status: "failed",
				}, nil
			},
		}
		svc := NewPaymentService(&mockPaymentRepo{}, userRepo, gw, cfg)
		_, err := svc.InitializePayment(ctx, dto.InitializePaymentInputDTO{
			Amount:   10,
			RecipeID: 1,
		}, userID.String())
		if !errors.Is(err, domainerrors.ErrPaymentFailed) {
			t.Errorf("expected ErrPaymentFailed, got %v", err)
		}
	})
}

func TestPaymentService_VerifyPayment(t *testing.T) {
	ctx := context.Background()
	txRef := "TX-123456789"
	userID := uuid.New()

	existingTx := &entities.Transaction{
		ID:          uuid.New(),
		UserID:      userID,
		RecipeID:    12,
		TrxRef:      txRef,
		Amount:      500,
		Currency:    "ETB",
		CheckoutURL: "https://checkout.chapa.co/pay/123",
	}

	t.Run("successful verification", func(t *testing.T) {
		var savedPayment *entities.Payment

		payRepo := &mockPaymentRepo{
			getTxFn: func(ctx context.Context, ref string) (*entities.Transaction, error) {
				if ref == txRef {
					return existingTx, nil
				}
				return nil, domainerrors.ErrTransactionNotFound
			},
			createPayFn: func(ctx context.Context, payment *entities.Payment) error {
				savedPayment = payment
				return nil
			},
		}

		gw := &mockPaymentGateway{
			verifyFn: func(ctx context.Context, ref string) (*ports.PaymentVerifyResult, error) {
				return &ports.PaymentVerifyResult{
					Status:  "success",
					Message: "Payment verified",
				}, nil
			},
		}

		svc := NewPaymentService(payRepo, &mockUserRepo{}, gw, nil)
		output, err := svc.VerifyPayment(ctx, txRef)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if output.TxRef != txRef {
			t.Errorf("expected txRef %s, got %s", txRef, output.TxRef)
		}
		if output.Status != "success" {
			t.Errorf("expected status success, got %s", output.Status)
		}

		if savedPayment == nil {
			t.Fatal("expected payment to be recorded")
		}
		if savedPayment.UserID != userID || savedPayment.RecipeID != 12 || savedPayment.Amount != 500 {
			t.Errorf("unexpected saved payment: %+v", savedPayment)
		}
	})

	t.Run("empty txRef", func(t *testing.T) {
		svc := NewPaymentService(&mockPaymentRepo{}, &mockUserRepo{}, &mockPaymentGateway{}, nil)
		_, err := svc.VerifyPayment(ctx, "   ")
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("transaction not found in database skips gateway call", func(t *testing.T) {
		gatewayCalled := false
		payRepo := &mockPaymentRepo{
			getTxFn: func(ctx context.Context, ref string) (*entities.Transaction, error) {
				return nil, domainerrors.ErrTransactionNotFound
			},
		}
		gw := &mockPaymentGateway{
			verifyFn: func(ctx context.Context, ref string) (*ports.PaymentVerifyResult, error) {
				gatewayCalled = true
				return &ports.PaymentVerifyResult{Status: "success"}, nil
			},
		}
		svc := NewPaymentService(payRepo, &mockUserRepo{}, gw, nil)
		_, err := svc.VerifyPayment(ctx, txRef)
		if !errors.Is(err, domainerrors.ErrTransactionNotFound) {
			t.Errorf("expected ErrTransactionNotFound, got %v", err)
		}
		if gatewayCalled {
			t.Error("expected gateway.Verify NOT to be called if transaction is not found locally")
		}
	})

	t.Run("gateway verification error", func(t *testing.T) {
		payRepo := &mockPaymentRepo{
			getTxFn: func(ctx context.Context, ref string) (*entities.Transaction, error) {
				return existingTx, nil
			},
		}
		gw := &mockPaymentGateway{
			verifyFn: func(ctx context.Context, ref string) (*ports.PaymentVerifyResult, error) {
				return nil, errors.New("chapa timeout")
			},
		}
		svc := NewPaymentService(payRepo, &mockUserRepo{}, gw, nil)
		_, err := svc.VerifyPayment(ctx, txRef)
		if !errors.Is(err, domainerrors.ErrPaymentFailed) {
			t.Errorf("expected ErrPaymentFailed, got %v", err)
		}
	})

	t.Run("gateway status not success", func(t *testing.T) {
		payRepo := &mockPaymentRepo{
			getTxFn: func(ctx context.Context, ref string) (*entities.Transaction, error) {
				return existingTx, nil
			},
		}
		gw := &mockPaymentGateway{
			verifyFn: func(ctx context.Context, ref string) (*ports.PaymentVerifyResult, error) {
				return &ports.PaymentVerifyResult{
					Status:  "failed",
					Message: "insufficient funds",
				}, nil
			},
		}
		svc := NewPaymentService(payRepo, &mockUserRepo{}, gw, nil)
		_, err := svc.VerifyPayment(ctx, txRef)
		if !errors.Is(err, domainerrors.ErrPaymentFailed) {
			t.Errorf("expected ErrPaymentFailed, got %v", err)
		}
	})
}

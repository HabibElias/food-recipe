package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go/food-recipe/internal/application/dto"
	"go/food-recipe/internal/config"
	"go/food-recipe/internal/domain/entities"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/domain/ports"
	"go/food-recipe/internal/domain/repositories"
)

type PaymentService struct {
	paymentRepo repositories.PaymentRepository
	userRepo    repositories.UserRepository
	gateway     ports.PaymentGateway
	cfg         *config.Config
}

func NewPaymentService(
	paymentRepo repositories.PaymentRepository,
	userRepo repositories.UserRepository,
	gateway ports.PaymentGateway,
	cfg *config.Config,
) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
		userRepo:    userRepo,
		gateway:     gateway,
		cfg:         cfg,
	}
}

func (s *PaymentService) InitializePayment(
	ctx context.Context,
	input dto.InitializePaymentInputDTO,
	sessionUserID string,
) (*dto.InitializePaymentOutputDTO, error) {
	sessionUserID = strings.TrimSpace(sessionUserID)
	if sessionUserID == "" {
		return nil, fmt.Errorf("%w: session user id is required", domainerrors.ErrUnauthorized)
	}

	uid, err := uuid.Parse(sessionUserID)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user id in session", domainerrors.ErrInvalidInput)
	}

	if input.Amount <= 0 {
		return nil, fmt.Errorf("%w: amount must be greater than zero", domainerrors.ErrInvalidInput)
	}

	if input.RecipeID <= 0 {
		return nil, fmt.Errorf("%w: invalid recipe id", domainerrors.ErrInvalidInput)
	}

	user, err := s.userRepo.GetByID(ctx, uid)
	if err != nil {
		return nil, err
	}

	txRef := fmt.Sprintf("TX-%d", time.Now().UnixNano())

	callbackURL := ""
	if s.cfg != nil {
		callbackURL = s.cfg.App.CallbackURL + "/v1/payment/verify"
	}

	initParams := ports.PaymentInitParams{
		Amount:      input.Amount,
		Currency:    "ETB",
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		TxRef:       txRef,
		Phone:       "0900123456",
		CallbackURL: callbackURL,
	}

	res, err := s.gateway.Initialize(ctx, initParams)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerrors.ErrPaymentFailed, err)
	}
	if res == nil || res.Status != "success" {
		errMsg := "unknown failure"
		if res != nil && res.Status != "" {
			errMsg = res.Status
		}
		return nil, fmt.Errorf("%w: payment initialization failed: %s", domainerrors.ErrPaymentFailed, errMsg)
	}

	tx := &entities.Transaction{
		UserID:      uid,
		RecipeID:    input.RecipeID,
		TrxRef:      txRef,
		Amount:      input.Amount,
		Currency:    "ETB",
		CheckoutURL: res.CheckoutURL,
	}

	if err := s.paymentRepo.CreateTransaction(ctx, tx); err != nil {
		return nil, err
	}

	return &dto.InitializePaymentOutputDTO{
		CheckoutURL: res.CheckoutURL,
		TxRef:       txRef,
		Status:      res.Status,
	}, nil
}

func (s *PaymentService) VerifyPayment(ctx context.Context, txRef string) (*dto.VerifyPaymentOutputDTO, error) {
	txRef = strings.TrimSpace(txRef)
	if txRef == "" {
		return nil, fmt.Errorf("%w: transaction reference is required", domainerrors.ErrInvalidInput)
	}

	res, err := s.gateway.Verify(ctx, txRef)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerrors.ErrPaymentFailed, err)
	}
	if res == nil || res.Status != "success" {
		msg := "verification failed"
		if res != nil && res.Message != "" {
			msg = res.Message
		}
		return nil, fmt.Errorf("%w: %s", domainerrors.ErrPaymentFailed, msg)
	}

	tx, err := s.paymentRepo.GetTransactionByRef(ctx, txRef)
	if err != nil {
		return nil, err
	}

	payment := &entities.Payment{
		UserID:      tx.UserID,
		RecipeID:    tx.RecipeID,
		TxRef:       txRef,
		Amount:      tx.Amount,
		CheckoutURL: tx.CheckoutURL,
		Currency:    tx.Currency,
		Status:      "success",
	}

	if err := s.paymentRepo.CreatePayment(ctx, payment); err != nil {
		return nil, err
	}

	return &dto.VerifyPaymentOutputDTO{
		Message: "Payment verified successfully",
		TxRef:   txRef,
		Status:  "success",
	}, nil
}

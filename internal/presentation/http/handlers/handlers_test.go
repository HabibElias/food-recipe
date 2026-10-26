package handlers_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"go/food-recipe/internal/application/dto"
	"go/food-recipe/internal/application/services"
	"go/food-recipe/internal/config"
	"go/food-recipe/internal/domain/entities"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/domain/ports"
	"go/food-recipe/internal/domain/repositories"
	"go/food-recipe/internal/presentation/http/handlers"
	"go/food-recipe/internal/presentation/http/router"
	"go/food-recipe/internal/presentation/response"
)

// --- Mock Implementations ---

type mockPaymentRepo struct {
	createTxFn         func(ctx context.Context, tx *entities.Transaction) error
	getTxByRefFn       func(ctx context.Context, txRef string) (*entities.Transaction, error)
	createPaymentFn    func(ctx context.Context, payment *entities.Payment) error
}

var _ repositories.PaymentRepository = (*mockPaymentRepo)(nil)

func (m *mockPaymentRepo) CreateTransaction(ctx context.Context, tx *entities.Transaction) error {
	if m.createTxFn != nil {
		return m.createTxFn(ctx, tx)
	}
	return nil
}

func (m *mockPaymentRepo) GetTransactionByRef(ctx context.Context, txRef string) (*entities.Transaction, error) {
	if m.getTxByRefFn != nil {
		return m.getTxByRefFn(ctx, txRef)
	}
	return nil, domainerrors.ErrTransactionNotFound
}

func (m *mockPaymentRepo) CreatePayment(ctx context.Context, payment *entities.Payment) error {
	if m.createPaymentFn != nil {
		return m.createPaymentFn(ctx, payment)
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
	return &ports.PaymentInitResult{CheckoutURL: "https://chapa.co/checkout/test", TxRef: params.TxRef, Status: "success"}, nil
}

func (m *mockPaymentGateway) Verify(ctx context.Context, txRef string) (*ports.PaymentVerifyResult, error) {
	if m.verifyFn != nil {
		return m.verifyFn(ctx, txRef)
	}
	return &ports.PaymentVerifyResult{Status: "success", Message: "Payment successful"}, nil
}

type mockFileStorage struct {
	uploadFn func(ctx context.Context, base64Data, publicID, folder string) (string, error)
}

var _ ports.FileStorage = (*mockFileStorage)(nil)

func (m *mockFileStorage) UploadBase64Image(ctx context.Context, base64Data, publicID, folder string) (string, error) {
	if m.uploadFn != nil {
		return m.uploadFn(ctx, base64Data, publicID, folder)
	}
	return "https://res.cloudinary.com/demo/image/upload/sample.jpg", nil
}

type mockRecipeRepo struct {
	getCreatorFn    func(ctx context.Context, likeID uuid.UUID) (uuid.UUID, string, error)
	addRecipeImgFn  func(ctx context.Context, img *entities.RecipeImage) error
}

var _ repositories.RecipeRepository = (*mockRecipeRepo)(nil)

func (m *mockRecipeRepo) GetCreatorFromLikeID(ctx context.Context, likeID uuid.UUID) (uuid.UUID, string, error) {
	if m.getCreatorFn != nil {
		return m.getCreatorFn(ctx, likeID)
	}
	return uuid.New(), "testuser", nil
}

func (m *mockRecipeRepo) AddRecipeImage(ctx context.Context, img *entities.RecipeImage) error {
	if m.addRecipeImgFn != nil {
		return m.addRecipeImgFn(ctx, img)
	}
	return nil
}

type mockNotificationRepo struct {
	createFn func(ctx context.Context, n *entities.Notification) error
}

var _ repositories.NotificationRepository = (*mockNotificationRepo)(nil)

func (m *mockNotificationRepo) Create(ctx context.Context, n *entities.Notification) error {
	if m.createFn != nil {
		return m.createFn(ctx, n)
	}
	return nil
}

// --- PaymentHandler Tests ---

func TestPaymentHandler_InitializePayment(t *testing.T) {
	userID := uuid.New()
	userRepo := &mockUserRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*entities.User, error) {
			return &entities.User{
				ID:        userID,
				FirstName: "Abebe",
				LastName:  "Bikila",
				Email:     "abebe@example.com",
			}, nil
		},
	}
	paymentRepo := &mockPaymentRepo{}
	gateway := &mockPaymentGateway{}
	cfg := &config.Config{
		App: config.AppConfig{CallbackURL: "http://localhost:8080"},
	}
	paymentSvc := services.NewPaymentService(paymentRepo, userRepo, gateway, cfg)
	handler := handlers.NewPaymentHandler(paymentSvc)

	t.Run("Success_WithSessionVariables", func(t *testing.T) {
		app := fiber.New()
		app.Post("/payment", handler.InitializePayment)

		payload := dto.InitializePaymentActionPayloadDTO{
			SessionVariables: map[string]interface{}{
				"x-hasura-user-id": userID.String(),
			},
			Input: struct {
				Amount   float64                            `json:"amount"`
				RecipeID int                                `json:"recipe_id"`
				Phone    string                             `json:"phone,omitempty"`
				Input    *dto.InitializePaymentInputDTO     `json:"input,omitempty"`
			}{
				Amount:   150.0,
				RecipeID: 42,
				Phone:    "0911223344",
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		var out dto.InitializePaymentOutputDTO
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if out.CheckoutURL == "" || out.Status != "success" {
			t.Errorf("unexpected output: %+v", out)
		}
	})

	t.Run("Success_WithLocalsFallback", func(t *testing.T) {
		app := fiber.New()
		// Middleware injects user_id into c.Locals
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_id", userID.String())
			return c.Next()
		})
		app.Post("/payment", handler.InitializePayment)

		payload := dto.InitializePaymentActionPayloadDTO{
			Input: struct {
				Amount   float64                        `json:"amount"`
				RecipeID int                            `json:"recipe_id"`
				Phone    string                         `json:"phone,omitempty"`
				Input    *dto.InitializePaymentInputDTO `json:"input,omitempty"`
			}{
				Amount:   99.0,
				RecipeID: 10,
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("InvalidBody", func(t *testing.T) {
		app := fiber.New()
		app.Post("/payment", handler.InitializePayment)

		req := httptest.NewRequest("POST", "/payment", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("MissingSessionUserID_ReturnsUnauthorized", func(t *testing.T) {
		app := fiber.New()
		app.Post("/payment", handler.InitializePayment)

		payload := dto.InitializePaymentActionPayloadDTO{
			Input: struct {
				Amount   float64                        `json:"amount"`
				RecipeID int                            `json:"recipe_id"`
				Phone    string                         `json:"phone,omitempty"`
				Input    *dto.InitializePaymentInputDTO `json:"input,omitempty"`
			}{
				Amount:   50.0,
				RecipeID: 1,
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/payment", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})
}

func TestPaymentHandler_VerifyPayment(t *testing.T) {
	paymentRepo := &mockPaymentRepo{
		getTxByRefFn: func(ctx context.Context, txRef string) (*entities.Transaction, error) {
			if txRef == "valid-ref" {
				return &entities.Transaction{
					UserID:      uuid.New(),
					RecipeID:    1,
					TrxRef:      "valid-ref",
					Amount:      50,
					Currency:    "ETB",
					CheckoutURL: "https://example.com",
				}, nil
			}
			return nil, domainerrors.ErrTransactionNotFound
		},
	}
	gateway := &mockPaymentGateway{}
	paymentSvc := services.NewPaymentService(paymentRepo, &mockUserRepo{}, gateway, nil)
	handler := handlers.NewPaymentHandler(paymentSvc)

	app := fiber.New()
	app.Get("/payment/verify", handler.VerifyPayment)

	t.Run("Success_WithTrxRef", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/payment/verify?trx_ref=valid-ref", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		var out dto.VerifyPaymentOutputDTO
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if out.Status != "success" || out.TxRef != "valid-ref" {
			t.Errorf("unexpected output: %+v", out)
		}
	})

	t.Run("Success_WithTxRef", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/payment/verify?tx_ref=valid-ref", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("MissingQueryParam", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/payment/verify", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("NotFoundRef", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/payment/verify?trx_ref=non-existent", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusNotFound {
			t.Errorf("expected 404, got %d", resp.StatusCode)
		}
	})
}

// --- MediaHandler Tests ---

func TestMediaHandler(t *testing.T) {
	storage := &mockFileStorage{}
	recipeRepo := &mockRecipeRepo{}
	userRepo := &mockUserRepo{}
	mediaSvc := services.NewMediaService(storage, recipeRepo, userRepo)
	handler := handlers.NewMediaHandler(mediaSvc)

	validBase64 := base64.StdEncoding.EncodeToString([]byte("fake image data bytes"))

	t.Run("UploadRecipeImages_Success", func(t *testing.T) {
		app := fiber.New()
		app.Post("/upload-recipe-images", handler.UploadRecipeImages)

		payload := dto.UploadRecipeImagesActionPayloadDTO{
			Input: struct {
				Input dto.UploadRecipeImagesInputDTO `json:"input"`
			}{
				Input: dto.UploadRecipeImagesInputDTO{
					RecipeID: 1,
					RecipeImages: []dto.RecipeImageItemDTO{
						{
							ImageName:      "test.png",
							ImageType:      "image/png",
							ImageBase64Str: validBase64,
							IsThumbnail:    true,
						},
					},
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/upload-recipe-images", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		var out dto.UploadRecipeImagesOutputDTO
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !out.Success {
			t.Errorf("expected success true")
		}
	})

	t.Run("UploadRecipeImages_InvalidPayload", func(t *testing.T) {
		app := fiber.New()
		app.Post("/upload-recipe-images", handler.UploadRecipeImages)

		req := httptest.NewRequest("POST", "/upload-recipe-images", bytes.NewBufferString("{invalid"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("UploadUserImage_Success_SessionVariables", func(t *testing.T) {
		app := fiber.New()
		app.Post("/upload-user-image", handler.UploadUserImage)

		uid := uuid.New()
		payload := dto.UploadUserImageActionPayloadDTO{
			SessionVariables: map[string]interface{}{
				"x-hasura-user-id": uid.String(),
			},
			Input: struct {
				Input dto.UploadUserImageInputDTO `json:"input"`
			}{
				Input: dto.UploadUserImageInputDTO{
					UserID: uid.String(),
					Image: dto.UserImageItemDTO{
						ImageName:      "avatar.jpg",
						ImageType:      "image/jpeg",
						ImageBase64Str: validBase64,
					},
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/upload-user-image", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		var out dto.UploadUserImageOutputDTO
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !out.Success || out.ImageURL == "" {
			t.Errorf("unexpected output: %+v", out)
		}
	})

	t.Run("UploadUserImage_Success_LocalsFallback", func(t *testing.T) {
		uid := uuid.New()
		app := fiber.New()
		app.Use(func(c *fiber.Ctx) error {
			c.Locals("user_id", uid.String())
			return c.Next()
		})
		app.Post("/upload-user-image", handler.UploadUserImage)

		payload := dto.UploadUserImageActionPayloadDTO{
			Input: struct {
				Input dto.UploadUserImageInputDTO `json:"input"`
			}{
				Input: dto.UploadUserImageInputDTO{
					UserID: uid.String(),
					Image: dto.UserImageItemDTO{
						ImageName:      "avatar.png",
						ImageType:      "png",
						ImageBase64Str: validBase64,
					},
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/upload-user-image", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("UploadUserImage_UnauthorizedMismatch", func(t *testing.T) {
		sessionUID := uuid.New()
		differentUID := uuid.New()

		app := fiber.New()
		app.Post("/upload-user-image", handler.UploadUserImage)

		payload := dto.UploadUserImageActionPayloadDTO{
			SessionVariables: map[string]interface{}{
				"x-hasura-user-id": sessionUID.String(),
			},
			Input: struct {
				Input dto.UploadUserImageInputDTO `json:"input"`
			}{
				Input: dto.UploadUserImageInputDTO{
					UserID: differentUID.String(),
					Image: dto.UserImageItemDTO{
						ImageName:      "avatar.jpg",
						ImageType:      "image/jpeg",
						ImageBase64Str: validBase64,
					},
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/upload-user-image", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})
}

// --- NotificationHandler Tests ---

func TestNotificationHandler(t *testing.T) {
	recipeRepo := &mockRecipeRepo{}
	notificationRepo := &mockNotificationRepo{}
	notificationSvc := services.NewNotificationService(recipeRepo, notificationRepo)
	handler := handlers.NewNotificationHandler(notificationSvc)

	app := fiber.New()
	app.Post("/like-notification", handler.LikeNotification)

	t.Run("Success_InsertEvent", func(t *testing.T) {
		payload := dto.LikeEventPayloadDTO{
			Event: dto.LikeEventDTO{
				Op: "INSERT",
				Data: dto.LikeEventDataDTO{
					New: dto.LikeEventDataNewDTO{
						ID:       uuid.New(),
						UserID:   uuid.New(),
						RecipeID: 1,
					},
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/like-notification", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		var out dto.LikeNotificationOutputDTO
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !out.OK {
			t.Errorf("expected OK true")
		}
	})

	t.Run("Skipped_NonInsertEvent", func(t *testing.T) {
		payload := dto.LikeEventPayloadDTO{
			Event: dto.LikeEventDTO{
				Op: "DELETE",
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/like-notification", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		var out dto.LikeNotificationOutputDTO
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !out.OK || !out.Skipped {
			t.Errorf("expected OK=true, Skipped=true, got %+v", out)
		}
	})

	t.Run("InvalidPayload", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/like-notification", bytes.NewBufferString("{bad"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})
}

// --- HealthHandler Tests ---

func TestHealthHandler(t *testing.T) {
	handler := handlers.NewHealthHandler()
	app := fiber.New()
	app.Get("/healthz", handler.Healthz)
	app.Get("/err", handler.Err)

	t.Run("Healthz", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/healthz", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var out map[string]string
		_ = json.Unmarshal(body, &out)
		if out["status"] != "ok" {
			t.Errorf("expected status ok, got %v", out["status"])
		}
	})

	t.Run("Err", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/err", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusInternalServerError {
			t.Errorf("expected 500, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var errResp response.ErrorResponse
		_ = json.Unmarshal(body, &errResp)
		if errResp.Message != "Internal Server Error" {
			t.Errorf("expected 'Internal Server Error', got %q", errResp.Message)
		}
	})
}

// --- Router SetupRoutes Tests ---

func TestSetupRoutes(t *testing.T) {
	app := fiber.New()

	authSvc := services.NewAuthService(&mockUserRepo{}, &config.Config{JWT: config.JWTConfig{Secret: "test"}})
	paymentSvc := services.NewPaymentService(&mockPaymentRepo{}, &mockUserRepo{}, &mockPaymentGateway{}, nil)
	mediaSvc := services.NewMediaService(&mockFileStorage{}, &mockRecipeRepo{}, &mockUserRepo{})
	notificationSvc := services.NewNotificationService(&mockRecipeRepo{}, &mockNotificationRepo{})

	authHandler := handlers.NewAuthHandler(authSvc)
	paymentHandler := handlers.NewPaymentHandler(paymentSvc)
	mediaHandler := handlers.NewMediaHandler(mediaSvc)
	notificationHandler := handlers.NewNotificationHandler(notificationSvc)
	healthHandler := handlers.NewHealthHandler()

	router.SetupRoutes(router.RouterParams{
		App:                 app,
		AuthHandler:         authHandler,
		PaymentHandler:      paymentHandler,
		MediaHandler:        mediaHandler,
		NotificationHandler: notificationHandler,
		HealthHandler:       healthHandler,
	})

	// Verify health route
	req := httptest.NewRequest("GET", "/v1/healthz", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("route /v1/healthz failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("expected 200 for /v1/healthz, got %d", resp.StatusCode)
	}

	// Verify health err route
	req = httptest.NewRequest("GET", "/v1/err", nil)
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("route /v1/err failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("expected 500 for /v1/err, got %d", resp.StatusCode)
	}

	// Verify notification route with invalid payload returns 400 (confirming route is bound)
	req = httptest.NewRequest("POST", "/v1/like-notification", bytes.NewBufferString("not-json"))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	if err != nil {
		t.Fatalf("route /v1/like-notification failed: %v", err)
	}
	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("expected 400 for /v1/like-notification invalid body, got %d", resp.StatusCode)
	}
}

// Ensure unused import warning check
var _ = errors.New

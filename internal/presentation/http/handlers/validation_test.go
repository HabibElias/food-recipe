package handlers_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http/httptest"
	"testing"

	"go/food-recipe/internal/application/dto"
	"go/food-recipe/internal/application/services"
	"go/food-recipe/internal/config"
	"go/food-recipe/internal/domain/entities"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/pkg/validator"
	"go/food-recipe/internal/presentation/http/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func TestPaymentHandler_ValidationAndHappyPath(t *testing.T) {
	userID := uuid.New()
	userRepo := &mockUserRepo{
		getByIDFn: func(ctx context.Context, id uuid.UUID) (*entities.User, error) {
			return &entities.User{
				ID:        userID,
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john.doe@example.com",
			}, nil
		},
	}
	paymentRepo := &mockPaymentRepo{
		getTxByRefFn: func(ctx context.Context, txRef string) (*entities.Transaction, error) {
			if txRef == "valid-ref-123" {
				return &entities.Transaction{
					UserID:      userID,
					RecipeID:    12,
					TrxRef:      "valid-ref-123",
					Amount:      200,
					Currency:    "ETB",
					CheckoutURL: "https://chapa.co/checkout/test",
				}, nil
			}
			return nil, domainerrors.ErrTransactionNotFound
		},
	}
	gateway := &mockPaymentGateway{}
	cfg := &config.Config{
		App: config.AppConfig{CallbackURL: "http://localhost:8080"},
	}
	paymentSvc := services.NewPaymentService(paymentRepo, userRepo, gateway, cfg)
	handler := handlers.NewPaymentHandler(paymentSvc)

	app := fiber.New()
	app.Post("/payment/init", handler.InitializePayment)
	app.Get("/payment/verify", handler.VerifyPayment)

	t.Run("InitializePayment_HappyPath", func(t *testing.T) {
		payload := dto.InitializePaymentActionPayloadDTO{
			SessionVariables: map[string]interface{}{
				"x-hasura-user-id": userID.String(),
			},
			Input: struct {
				Amount   float64                        `json:"amount"`
				RecipeID int                            `json:"recipe_id"`
				Phone    string                         `json:"phone,omitempty"`
				Input    *dto.InitializePaymentInputDTO `json:"input,omitempty"`
			}{
				Amount:   200.0,
				RecipeID: 12,
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/payment/init", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var out dto.InitializePaymentOutputDTO
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if out.Status != "success" || out.CheckoutURL == "" {
			t.Errorf("unexpected output: %+v", out)
		}
	})

	t.Run("InitializePayment_ValidationError_InvalidAmount", func(t *testing.T) {
		payload := dto.InitializePaymentActionPayloadDTO{
			SessionVariables: map[string]interface{}{
				"x-hasura-user-id": userID.String(),
			},
			Input: struct {
				Amount   float64                        `json:"amount"`
				RecipeID int                            `json:"recipe_id"`
				Phone    string                         `json:"phone,omitempty"`
				Input    *dto.InitializePaymentInputDTO `json:"input,omitempty"`
			}{
				Amount:   0,
				RecipeID: 10,
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/payment/init", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var out struct {
			Message string                 `json:"message"`
			Errors  []validator.FieldError `json:"errors"`
		}
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(out.Errors) == 0 {
			t.Errorf("expected validation errors array to be populated")
		}
	})

	t.Run("InitializePayment_ValidationError_InvalidRecipeID", func(t *testing.T) {
		payload := dto.InitializePaymentActionPayloadDTO{
			SessionVariables: map[string]interface{}{
				"x-hasura-user-id": userID.String(),
			},
			Input: struct {
				Amount   float64                        `json:"amount"`
				RecipeID int                            `json:"recipe_id"`
				Phone    string                         `json:"phone,omitempty"`
				Input    *dto.InitializePaymentInputDTO `json:"input,omitempty"`
			}{
				Amount:   100.0,
				RecipeID: -5,
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/payment/init", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var out struct {
			Message string                 `json:"message"`
			Errors  []validator.FieldError `json:"errors"`
		}
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(out.Errors) == 0 {
			t.Errorf("expected validation errors array to be populated")
		}
	})

	t.Run("VerifyPayment_HappyPath", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/payment/verify?trx_ref=valid-ref-123", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("VerifyPayment_ValidationError_MissingTrxRef", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/payment/verify", nil)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var out struct {
			Message string                 `json:"message"`
			Errors  []validator.FieldError `json:"errors"`
		}
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(out.Errors) == 0 {
			t.Errorf("expected validation errors array to be populated")
		}
	})
}

func TestMediaHandler_ValidationAndHappyPath(t *testing.T) {
	storage := &mockFileStorage{}
	recipeRepo := &mockRecipeRepo{}
	userRepo := &mockUserRepo{}
	mediaSvc := services.NewMediaService(storage, recipeRepo, userRepo)
	handler := handlers.NewMediaHandler(mediaSvc)

	validBase64 := base64.StdEncoding.EncodeToString([]byte("valid-image-bytes"))

	app := fiber.New()
	app.Post("/recipe-images", handler.UploadRecipeImages)
	app.Post("/user-image", handler.UploadUserImage)

	t.Run("UploadRecipeImages_HappyPath", func(t *testing.T) {
		payload := dto.UploadRecipeImagesActionPayloadDTO{
			Input: struct {
				Input dto.UploadRecipeImagesInputDTO `json:"input"`
			}{
				Input: dto.UploadRecipeImagesInputDTO{
					RecipeID: 10,
					RecipeImages: []dto.RecipeImageItemDTO{
						{
							ImageName:      "food.jpg",
							ImageType:      "image/jpeg",
							ImageBase64Str: validBase64,
							IsThumbnail:    true,
						},
					},
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/recipe-images", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}

		var out dto.UploadRecipeImagesOutputDTO
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if !out.Success {
			t.Errorf("expected Success true")
		}
	})

	t.Run("UploadRecipeImages_ValidationError_InvalidRecipeID", func(t *testing.T) {
		payload := dto.UploadRecipeImagesActionPayloadDTO{
			Input: struct {
				Input dto.UploadRecipeImagesInputDTO `json:"input"`
			}{
				Input: dto.UploadRecipeImagesInputDTO{
					RecipeID: 0,
					RecipeImages: []dto.RecipeImageItemDTO{
						{
							ImageName:      "food.jpg",
							ImageType:      "image/jpeg",
							ImageBase64Str: validBase64,
						},
					},
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/recipe-images", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var out struct {
			Message string                 `json:"message"`
			Errors  []validator.FieldError `json:"errors"`
		}
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(out.Errors) == 0 {
			t.Errorf("expected validation errors array")
		}
	})

	t.Run("UploadRecipeImages_ValidationError_EmptyImages", func(t *testing.T) {
		payload := dto.UploadRecipeImagesActionPayloadDTO{
			Input: struct {
				Input dto.UploadRecipeImagesInputDTO `json:"input"`
			}{
				Input: dto.UploadRecipeImagesInputDTO{
					RecipeID:     5,
					RecipeImages: []dto.RecipeImageItemDTO{},
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/recipe-images", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("UploadRecipeImages_ValidationError_MissingNestedField", func(t *testing.T) {
		payload := dto.UploadRecipeImagesActionPayloadDTO{
			Input: struct {
				Input dto.UploadRecipeImagesInputDTO `json:"input"`
			}{
				Input: dto.UploadRecipeImagesInputDTO{
					RecipeID: 5,
					RecipeImages: []dto.RecipeImageItemDTO{
						{
							ImageName:      "",
							ImageType:      "image/jpeg",
							ImageBase64Str: validBase64,
						},
					},
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/recipe-images", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("UploadUserImage_HappyPath", func(t *testing.T) {
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

		req := httptest.NewRequest("POST", "/user-image", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
		}
	})

	t.Run("UploadUserImage_ValidationError_InvalidUUID", func(t *testing.T) {
		payload := dto.UploadUserImageActionPayloadDTO{
			Input: struct {
				Input dto.UploadUserImageInputDTO `json:"input"`
			}{
				Input: dto.UploadUserImageInputDTO{
					UserID: "not-a-valid-uuid-format",
					Image: dto.UserImageItemDTO{
						ImageName:      "avatar.jpg",
						ImageType:      "image/jpeg",
						ImageBase64Str: validBase64,
					},
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/user-image", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var out struct {
			Message string                 `json:"message"`
			Errors  []validator.FieldError `json:"errors"`
		}
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(out.Errors) == 0 {
			t.Errorf("expected validation errors array")
		}
	})

	t.Run("UploadUserImage_ValidationError_MissingImageName", func(t *testing.T) {
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
						ImageName:      "",
						ImageType:      "image/jpeg",
						ImageBase64Str: validBase64,
					},
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/user-image", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})
}

func TestNotificationHandler_ValidationAndHappyPath(t *testing.T) {
	recipeRepo := &mockRecipeRepo{}
	notificationRepo := &mockNotificationRepo{}
	notificationSvc := services.NewNotificationService(recipeRepo, notificationRepo)
	handler := handlers.NewNotificationHandler(notificationSvc)

	app := fiber.New()
	app.Post("/notification", handler.LikeNotification)

	t.Run("LikeNotification_HappyPath_Insert", func(t *testing.T) {
		payload := dto.LikeEventPayloadDTO{
			Event: dto.LikeEventDTO{
				Op: "INSERT",
				Data: dto.LikeEventDataDTO{
					New: dto.LikeEventDataNewDTO{
						ID:       uuid.New(),
						UserID:   uuid.New(),
						RecipeID: 42,
					},
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/notification", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
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

	t.Run("LikeNotification_HappyPath_NonInsertSkipped", func(t *testing.T) {
		payload := dto.LikeEventPayloadDTO{
			Event: dto.LikeEventDTO{
				Op: "DELETE",
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/notification", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Fatalf("expected 200, got %d", resp.StatusCode)
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

	t.Run("LikeNotification_ValidationError_InvalidNewData", func(t *testing.T) {
		payload := dto.LikeEventPayloadDTO{
			Event: dto.LikeEventDTO{
				Op: "INSERT",
				Data: dto.LikeEventDataDTO{
					New: dto.LikeEventDataNewDTO{
						ID:       uuid.Nil,
						UserID:   uuid.Nil,
						RecipeID: 0,
					},
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/notification", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var out struct {
			Message string                 `json:"message"`
			Errors  []validator.FieldError `json:"errors"`
		}
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if len(out.Errors) == 0 {
			t.Errorf("expected validation errors array")
		}
	})

	t.Run("LikeNotification_ValidationError_MissingNewData", func(t *testing.T) {
		payload := dto.LikeEventPayloadDTO{
			Event: dto.LikeEventDTO{
				Op:   "INSERT",
				Data: dto.LikeEventDataDTO{},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/notification", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Fatalf("expected 400, got %d", resp.StatusCode)
		}
	})
}

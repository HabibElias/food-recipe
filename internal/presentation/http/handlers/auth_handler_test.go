package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
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
	"go/food-recipe/internal/domain/repositories"
	"go/food-recipe/internal/infrastructure/security"
	"go/food-recipe/internal/presentation/http/handlers"
	"go/food-recipe/internal/presentation/response"
)

type mockUserRepo struct {
	getByIDFn       func(ctx context.Context, id uuid.UUID) (*entities.User, error)
	getByEmailFn    func(ctx context.Context, email string) (*entities.User, error)
	getByUsernameFn func(ctx context.Context, username string) (*entities.User, error)
	createFn        func(ctx context.Context, user *entities.User) (*entities.User, error)
	updateAvatarFn  func(ctx context.Context, userID uuid.UUID, avatarURL string) error
}

var _ repositories.UserRepository = (*mockUserRepo)(nil)

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, domainerrors.ErrUserNotFound
}

func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, domainerrors.ErrUserNotFound
}

func (m *mockUserRepo) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	if m.getByUsernameFn != nil {
		return m.getByUsernameFn(ctx, username)
	}
	return nil, domainerrors.ErrUserNotFound
}

func (m *mockUserRepo) Create(ctx context.Context, user *entities.User) (*entities.User, error) {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return user, nil
}

func (m *mockUserRepo) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	if m.updateAvatarFn != nil {
		return m.updateAvatarFn(ctx, userID, avatarURL)
	}
	return nil
}

func setupAuthHandlerTest(repo *mockUserRepo) (*fiber.App, *handlers.AuthHandler, string) {
	jwtSecret := "test-secret-key-12345"
	cfg := &config.Config{
		JWT: config.JWTConfig{Secret: jwtSecret},
	}
	authSvc := services.NewAuthService(repo, cfg)
	handler := handlers.NewAuthHandler(authSvc)

	app := fiber.New()
	app.Post("/login", handler.Login)
	app.Post("/signup", handler.Signup)
	app.Post("/me", handler.Me)

	return app, handler, jwtSecret
}

func TestAuthHandler_Login(t *testing.T) {
	userID := uuid.New()
	pwHash, _ := security.HashPassword("validpassword123")

	t.Run("Success", func(t *testing.T) {
		repo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
				return &entities.User{
					ID:           userID,
					Email:        "user@example.com",
					PasswordHash: pwHash,
				}, nil
			},
		}
		app, _, _ := setupAuthHandlerTest(repo)

		payload := dto.LoginActionPayloadDTO{
			Input: struct {
				Input dto.LoginInputDTO `json:"input"`
			}{
				Input: dto.LoginInputDTO{
					Email:    "user@example.com",
					Password: "validpassword123",
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}

		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		var out dto.LoginOutputDTO
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if out.UserID != userID.String() || out.Token == "" {
			t.Errorf("unexpected output: %+v", out)
		}
	})

	t.Run("InvalidBody", func(t *testing.T) {
		repo := &mockUserRepo{}
		app, _, _ := setupAuthHandlerTest(repo)

		req := httptest.NewRequest("POST", "/login", bytes.NewBufferString("not-a-json"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("InvalidCredentials", func(t *testing.T) {
		repo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
				return nil, domainerrors.ErrUserNotFound
			},
		}
		app, _, _ := setupAuthHandlerTest(repo)

		payload := dto.LoginActionPayloadDTO{
			Input: struct {
				Input dto.LoginInputDTO `json:"input"`
			}{
				Input: dto.LoginInputDTO{
					Email:    "unknown@example.com",
					Password: "wrongpassword",
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}
	})

	t.Run("ValidationError", func(t *testing.T) {
		repo := &mockUserRepo{}
		app, _, _ := setupAuthHandlerTest(repo)

		payload := dto.LoginActionPayloadDTO{
			Input: struct {
				Input dto.LoginInputDTO `json:"input"`
			}{
				Input: dto.LoginInputDTO{
					Email:    "",
					Password: "",
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(bodyBytes))
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

func TestAuthHandler_Signup(t *testing.T) {
	newUID := uuid.New()

	t.Run("Success", func(t *testing.T) {
		repo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
				return nil, domainerrors.ErrUserNotFound
			},
			getByUsernameFn: func(ctx context.Context, username string) (*entities.User, error) {
				return nil, domainerrors.ErrUserNotFound
			},
			createFn: func(ctx context.Context, user *entities.User) (*entities.User, error) {
				user.ID = newUID
				return user, nil
			},
		}
		app, _, _ := setupAuthHandlerTest(repo)

		payload := dto.SignupActionPayloadDTO{
			Input: struct {
				Input dto.SignupInputDTO `json:"input"`
			}{
				Input: dto.SignupInputDTO{
					FirstName: "Jane",
					LastName:  "Doe",
					Email:     "jane.doe@example.com",
					Username:  "janedoe",
					Password:  "secure123",
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		var out dto.SignupOutputDTO
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if out.UserID != newUID.String() || out.Token == "" {
			t.Errorf("unexpected output: %+v", out)
		}
	})

	t.Run("InvalidBody", func(t *testing.T) {
		repo := &mockUserRepo{}
		app, _, _ := setupAuthHandlerTest(repo)

		req := httptest.NewRequest("POST", "/signup", bytes.NewBufferString("{invalid-json"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("UserAlreadyExists", func(t *testing.T) {
		repo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
				return &entities.User{Email: email}, nil
			},
		}
		app, _, _ := setupAuthHandlerTest(repo)

		payload := dto.SignupActionPayloadDTO{
			Input: struct {
				Input dto.SignupInputDTO `json:"input"`
			}{
				Input: dto.SignupInputDTO{
					FirstName: "Jane",
					LastName:  "Doe",
					Email:     "jane.doe@example.com",
					Username:  "janedoe",
					Password:  "secure123",
				},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusConflict {
			t.Errorf("expected 409, got %d", resp.StatusCode)
		}
	})
}

func TestAuthHandler_Me(t *testing.T) {
	userID := uuid.New()
	email := "me@example.com"

	repo := &mockUserRepo{
		getByEmailFn: func(ctx context.Context, em string) (*entities.User, error) {
			if em == email {
				return &entities.User{ID: userID, Email: email}, nil
			}
			return nil, domainerrors.ErrUserNotFound
		},
	}
	app, _, secret := setupAuthHandlerTest(repo)

	validToken, err := security.GenerateHasuraJWT(userID.String(), email, secret)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	t.Run("Success_WithPayload", func(t *testing.T) {
		payload := dto.MeActionPayloadDTO{
			Input: struct {
				Input dto.MeInputDTO `json:"input"`
			}{
				Input: dto.MeInputDTO{Token: validToken},
			},
		}
		bodyBytes, _ := json.Marshal(payload)

		req := httptest.NewRequest("POST", "/me", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		var out dto.MeOutputDTO
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if out.UserID != userID.String() || out.Token == "" {
			t.Errorf("unexpected output: %+v", out)
		}
	})

	t.Run("Success_WithAuthHeaderFallback", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/me", nil)
		req.Header.Set("Authorization", "Bearer "+validToken)
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusOK {
			t.Errorf("expected 200, got %d", resp.StatusCode)
		}

		var out dto.MeOutputDTO
		body, _ := io.ReadAll(resp.Body)
		if err := json.Unmarshal(body, &out); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if out.UserID != userID.String() {
			t.Errorf("expected userID %s, got %s", userID.String(), out.UserID)
		}
	})

	t.Run("MissingToken_ReturnsBadRequest", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/me", bytes.NewBufferString("{}"))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusBadRequest {
			t.Errorf("expected 400, got %d", resp.StatusCode)
		}
	})

	t.Run("InvalidToken_ReturnsUnauthorized", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/me", nil)
		req.Header.Set("Authorization", "Bearer invalid-jwt-token")
		resp, err := app.Test(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		if resp.StatusCode != fiber.StatusUnauthorized {
			t.Errorf("expected 401, got %d", resp.StatusCode)
		}

		body, _ := io.ReadAll(resp.Body)
		var errResp response.ErrorResponse
		_ = json.Unmarshal(body, &errResp)
		if errResp.Message == "" {
			t.Errorf("expected error message in response")
		}
	})
}

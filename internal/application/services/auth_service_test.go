package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"go/food-recipe/internal/application/dto"
	"go/food-recipe/internal/config"
	"go/food-recipe/internal/domain/entities"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/domain/repositories"
	"go/food-recipe/internal/infrastructure/security"
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

func testConfig() *config.Config {
	return &config.Config{
		JWT: config.JWTConfig{
			Secret: "super-secret-jwt-key-for-test-32bytes!",
		},
		App: config.AppConfig{
			CallbackURL: "http://localhost:8000",
		},
	}
}

func TestAuthService_Login(t *testing.T) {
	cfg := testConfig()
	ctx := context.Background()

	hashedPassword, err := security.HashPassword("validpassword123")
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}

	testUser := &entities.User{
		ID:           uuid.New(),
		Email:        "test@example.com",
		Username:     "testuser",
		PasswordHash: hashedPassword,
	}

	t.Run("successful login", func(t *testing.T) {
		repo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
				if email == "test@example.com" {
					return testUser, nil
				}
				return nil, domainerrors.ErrUserNotFound
			},
		}

		svc := NewAuthService(repo, cfg)
		output, err := svc.Login(ctx, dto.LoginInputDTO{
			Email:    "  TEST@example.com  ",
			Password: "validpassword123",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output.UserID != testUser.ID.String() {
			t.Errorf("expected userID %s, got %s", testUser.ID.String(), output.UserID)
		}
		if output.Token == "" {
			t.Error("expected non-empty token")
		}
	})

	t.Run("validation: missing email", func(t *testing.T) {
		svc := NewAuthService(&mockUserRepo{}, cfg)
		_, err := svc.Login(ctx, dto.LoginInputDTO{
			Email:    "   ",
			Password: "password",
		})
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("validation: missing password", func(t *testing.T) {
		svc := NewAuthService(&mockUserRepo{}, cfg)
		_, err := svc.Login(ctx, dto.LoginInputDTO{
			Email:    "test@example.com",
			Password: "",
		})
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("user not found returns ErrInvalidCredentials", func(t *testing.T) {
		repo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
				return nil, domainerrors.ErrUserNotFound
			},
		}
		svc := NewAuthService(repo, cfg)
		_, err := svc.Login(ctx, dto.LoginInputDTO{
			Email:    "unknown@example.com",
			Password: "password123",
		})
		if !errors.Is(err, domainerrors.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("wrong password returns ErrInvalidCredentials", func(t *testing.T) {
		repo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
				return testUser, nil
			},
		}
		svc := NewAuthService(repo, cfg)
		_, err := svc.Login(ctx, dto.LoginInputDTO{
			Email:    "test@example.com",
			Password: "wrongpassword",
		})
		if !errors.Is(err, domainerrors.ErrInvalidCredentials) {
			t.Errorf("expected ErrInvalidCredentials, got %v", err)
		}
	})

	t.Run("database error propagated", func(t *testing.T) {
		repo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
				return nil, domainerrors.ErrServiceUnavailable
			},
		}
		svc := NewAuthService(repo, cfg)
		_, err := svc.Login(ctx, dto.LoginInputDTO{
			Email:    "test@example.com",
			Password: "validpassword123",
		})
		if !errors.Is(err, domainerrors.ErrServiceUnavailable) {
			t.Errorf("expected ErrServiceUnavailable, got %v", err)
		}
	})
}

func TestAuthService_Signup(t *testing.T) {
	cfg := testConfig()
	ctx := context.Background()

	validInput := dto.SignupInputDTO{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		Username:  "johndoe",
		Password:  "secret123",
	}

	t.Run("successful signup", func(t *testing.T) {
		newID := uuid.New()
		var createdUserEntity *entities.User

		repo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
				return nil, domainerrors.ErrUserNotFound
			},
			getByUsernameFn: func(ctx context.Context, username string) (*entities.User, error) {
				return nil, domainerrors.ErrUserNotFound
			},
			createFn: func(ctx context.Context, user *entities.User) (*entities.User, error) {
				createdUserEntity = user
				userCopy := *user
				userCopy.ID = newID
				return &userCopy, nil
			},
		}

		svc := NewAuthService(repo, cfg)
		output, err := svc.Signup(ctx, validInput)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output.UserID != newID.String() {
			t.Errorf("expected UserID %s, got %s", newID.String(), output.UserID)
		}
		if output.Token == "" {
			t.Error("expected non-empty token")
		}
		if createdUserEntity == nil {
			t.Fatal("expected user to be created")
		}
		if createdUserEntity.PasswordHash == "secret123" || createdUserEntity.PasswordHash == "" {
			t.Error("expected password to be hashed")
		}
	})

	t.Run("validations", func(t *testing.T) {
		tests := []struct {
			name  string
			input dto.SignupInputDTO
		}{
			{"empty first name", dto.SignupInputDTO{FirstName: "", LastName: "Doe", Email: "a@b.com", Username: "user1", Password: "pass12"}},
			{"short first name", dto.SignupInputDTO{FirstName: "J", LastName: "Doe", Email: "a@b.com", Username: "user1", Password: "pass12"}},
			{"empty last name", dto.SignupInputDTO{FirstName: "John", LastName: "", Email: "a@b.com", Username: "user1", Password: "pass12"}},
			{"short last name", dto.SignupInputDTO{FirstName: "John", LastName: "D", Email: "a@b.com", Username: "user1", Password: "pass12"}},
			{"empty password", dto.SignupInputDTO{FirstName: "John", LastName: "Doe", Email: "a@b.com", Username: "user1", Password: ""}},
			{"short password", dto.SignupInputDTO{FirstName: "John", LastName: "Doe", Email: "a@b.com", Username: "user1", Password: "12345"}},
			{"empty username", dto.SignupInputDTO{FirstName: "John", LastName: "Doe", Email: "a@b.com", Username: "", Password: "pass12"}},
			{"short username", dto.SignupInputDTO{FirstName: "John", LastName: "Doe", Email: "a@b.com", Username: "user", Password: "pass12"}},
			{"empty email", dto.SignupInputDTO{FirstName: "John", LastName: "Doe", Email: "", Username: "user1", Password: "pass12"}},
			{"invalid email format", dto.SignupInputDTO{FirstName: "John", LastName: "Doe", Email: "invalidemail", Username: "user1", Password: "pass12"}},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				svc := NewAuthService(&mockUserRepo{}, cfg)
				_, err := svc.Signup(ctx, tc.input)
				if !errors.Is(err, domainerrors.ErrInvalidInput) {
					t.Errorf("expected ErrInvalidInput, got %v", err)
				}
			})
		}
	})

	t.Run("email already registered", func(t *testing.T) {
		repo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
				return &entities.User{ID: uuid.New(), Email: email}, nil
			},
		}
		svc := NewAuthService(repo, cfg)
		_, err := svc.Signup(ctx, validInput)
		if !errors.Is(err, domainerrors.ErrUserAlreadyExists) {
			t.Errorf("expected ErrUserAlreadyExists, got %v", err)
		}
	})

	t.Run("username already taken", func(t *testing.T) {
		repo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
				return nil, domainerrors.ErrUserNotFound
			},
			getByUsernameFn: func(ctx context.Context, username string) (*entities.User, error) {
				return &entities.User{ID: uuid.New(), Username: username}, nil
			},
		}
		svc := NewAuthService(repo, cfg)
		_, err := svc.Signup(ctx, validInput)
		if !errors.Is(err, domainerrors.ErrUsernameTaken) {
			t.Errorf("expected ErrUsernameTaken, got %v", err)
		}
	})
}

func TestAuthService_Me(t *testing.T) {
	cfg := testConfig()
	ctx := context.Background()

	testUser := &entities.User{
		ID:        uuid.New(),
		Email:     "user@example.com",
		Username:  "sampleuser",
		FirstName: "Sample",
		LastName:  "User",
	}

	validToken, err := security.GenerateHasuraJWT(testUser.ID.String(), testUser.Email, cfg.JWT.Secret)
	if err != nil {
		t.Fatalf("failed to generate valid token: %v", err)
	}

	t.Run("successful me query", func(t *testing.T) {
		repo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
				if email == testUser.Email {
					return testUser, nil
				}
				return nil, domainerrors.ErrUserNotFound
			},
		}

		svc := NewAuthService(repo, cfg)
		output, err := svc.Me(ctx, validToken)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if output.UserID != testUser.ID.String() {
			t.Errorf("expected userID %s, got %s", testUser.ID.String(), output.UserID)
		}
		if output.Token == "" {
			t.Error("expected refreshed token")
		}
	})

	t.Run("empty token", func(t *testing.T) {
		svc := NewAuthService(&mockUserRepo{}, cfg)
		_, err := svc.Me(ctx, "  ")
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("invalid token", func(t *testing.T) {
		svc := NewAuthService(&mockUserRepo{}, cfg)
		_, err := svc.Me(ctx, "invalid.jwt.token")
		if !errors.Is(err, domainerrors.ErrUnauthorized) {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		repo := &mockUserRepo{
			getByEmailFn: func(ctx context.Context, email string) (*entities.User, error) {
				return nil, domainerrors.ErrUserNotFound
			},
		}
		svc := NewAuthService(repo, cfg)
		_, err := svc.Me(ctx, validToken)
		if !errors.Is(err, domainerrors.ErrUserNotFound) {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}

// Auth action profiles must satisfy the frontend without returning persistence
// entities, password hashes, or another user's private email.
func TestAuthActionsReturnSafeAuthenticatedProfile(t *testing.T) {
	cfg := testConfig()
	hash, err := security.HashPassword("profile-password")
	if err != nil {
		t.Fatal(err)
	}
	user := &entities.User{ID: uuid.New(), Email: "profile@example.invalid", Username: "profileuser", FirstName: "Profile", LastName: "User", PasswordHash: hash, Bio: "A cook", CreatedAt: time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), UpdatedAt: time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)}
	repo := &mockUserRepo{getByEmailFn: func(context.Context, string) (*entities.User, error) { return user, nil }}
	svc := NewAuthService(repo, cfg)
	token, err := security.GenerateHasuraJWT(user.ID.String(), user.Email, cfg.JWT.Secret)
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name, profileKey string
		run              func() (any, error)
	}{
		{"login", "loggedInUser", func() (any, error) {
			return svc.Login(context.Background(), dto.LoginInputDTO{Email: user.Email, Password: "profile-password"})
		}},
		{"me", "user", func() (any, error) { return svc.Me(context.Background(), token) }},
		{"signup", "createdUser", func() (any, error) {
			signup := NewAuthService(&mockUserRepo{createFn: func(context.Context, *entities.User) (*entities.User, error) { return user, nil }}, cfg)
			return signup.Signup(context.Background(), dto.SignupInputDTO{Email: user.Email, Password: "profile-password", Username: user.Username, FirstName: user.FirstName, LastName: user.LastName})
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := tc.run()
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := json.Marshal(result)
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(encoded), hash) || strings.Contains(strings.ToLower(string(encoded)), "password") {
				t.Fatal("auth response leaked password information")
			}
			var body map[string]any
			if err := json.Unmarshal(encoded, &body); err != nil {
				t.Fatal(err)
			}
			profile, ok := body[tc.profileKey].(map[string]any)
			if !ok {
				t.Fatal("missing frontend profile object")
			}
			if body["user_id"] != user.ID.String() || body["token"] == "" || profile["id"] != user.ID.String() || profile["email"] != user.Email || profile["bio"] != user.Bio || profile["created_at"] != "2026-01-01T00:00:00Z" {
				t.Fatal("auth response did not contain the authenticated user's profile")
			}
		})
	}
}

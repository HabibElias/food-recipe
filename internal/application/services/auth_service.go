package services

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"go/food-recipe/internal/application/dto"
	"go/food-recipe/internal/config"
	"go/food-recipe/internal/domain/entities"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/domain/repositories"
	"go/food-recipe/internal/infrastructure/security"
)

type AuthService struct {
	userRepo repositories.UserRepository
	cfg      *config.Config
}

func NewAuthService(userRepo repositories.UserRepository, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

func (s *AuthService) Login(ctx context.Context, input dto.LoginInputDTO) (*dto.LoginOutputDTO, error) {
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" {
		return nil, fmt.Errorf("%w: email is required", domainerrors.ErrInvalidInput)
	}
	if input.Password == "" {
		return nil, fmt.Errorf("%w: password is required", domainerrors.ErrInvalidInput)
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domainerrors.ErrUserNotFound) {
			return nil, domainerrors.ErrInvalidCredentials
		}
		return nil, err
	}

	if err := security.ComparePassword(user.PasswordHash, input.Password); err != nil {
		return nil, domainerrors.ErrInvalidCredentials
	}

	token, err := security.GenerateHasuraJWT(user.ID.String(), user.Email, s.cfg.JWT.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.LoginOutputDTO{
		UserID: user.ID.String(),
		Token:  token,
	}, nil
}

func (s *AuthService) Signup(ctx context.Context, input dto.SignupInputDTO) (*dto.SignupOutputDTO, error) {
	firstName := strings.TrimSpace(input.FirstName)
	if firstName == "" {
		return nil, fmt.Errorf("%w: first name is required", domainerrors.ErrInvalidInput)
	}
	if len(firstName) < 2 {
		return nil, fmt.Errorf("%w: first name must be at least 2 characters", domainerrors.ErrInvalidInput)
	}

	lastName := strings.TrimSpace(input.LastName)
	if lastName == "" {
		return nil, fmt.Errorf("%w: last name is required", domainerrors.ErrInvalidInput)
	}
	if len(lastName) < 2 {
		return nil, fmt.Errorf("%w: last name must be at least 2 characters", domainerrors.ErrInvalidInput)
	}

	if input.Password == "" {
		return nil, fmt.Errorf("%w: password is required", domainerrors.ErrInvalidInput)
	}
	if len(input.Password) < 6 {
		return nil, fmt.Errorf("%w: password must be at least 6 characters", domainerrors.ErrInvalidInput)
	}

	username := strings.TrimSpace(input.Username)
	if username == "" {
		return nil, fmt.Errorf("%w: username is required", domainerrors.ErrInvalidInput)
	}
	if len(username) < 5 {
		return nil, fmt.Errorf("%w: username must be at least 5 characters", domainerrors.ErrInvalidInput)
	}

	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" {
		return nil, fmt.Errorf("%w: email is required", domainerrors.ErrInvalidInput)
	}
	if !strings.Contains(email, "@") {
		return nil, fmt.Errorf("%w: email must be a valid email address", domainerrors.ErrInvalidInput)
	}

	// Check email uniqueness
	existingUser, err := s.userRepo.GetByEmail(ctx, email)
	if err == nil && existingUser != nil {
		return nil, domainerrors.ErrUserAlreadyExists
	}
	if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
		return nil, err
	}

	// Check username uniqueness
	existingUser, err = s.userRepo.GetByUsername(ctx, username)
	if err == nil && existingUser != nil {
		return nil, domainerrors.ErrUsernameTaken
	}
	if err != nil && !errors.Is(err, domainerrors.ErrUserNotFound) {
		return nil, err
	}

	hash, err := security.HashPassword(input.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to process password: %w", err)
	}

	newUser := &entities.User{
		Email:        email,
		Username:     username,
		FirstName:    firstName,
		LastName:     lastName,
		PasswordHash: hash,
	}

	createdUser, err := s.userRepo.Create(ctx, newUser)
	if err != nil {
		return nil, err
	}

	token, err := security.GenerateHasuraJWT(createdUser.ID.String(), createdUser.Email, s.cfg.JWT.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate authentication token: %w", err)
	}

	return &dto.SignupOutputDTO{
		UserID: createdUser.ID.String(),
		Token:  token,
	}, nil
}

func (s *AuthService) Me(ctx context.Context, token string) (*dto.MeOutputDTO, error) {
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("%w: token is required", domainerrors.ErrInvalidInput)
	}

	claims, err := security.VerifyHasuraJWT(token, s.cfg.JWT.Secret)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", domainerrors.ErrUnauthorized, err)
	}

	emailRaw, ok := claims["email"]
	if !ok {
		return nil, fmt.Errorf("%w: invalid token claims: email missing", domainerrors.ErrUnauthorized)
	}
	email, ok := emailRaw.(string)
	if !ok || email == "" {
		return nil, fmt.Errorf("%w: invalid token claims: email not string", domainerrors.ErrUnauthorized)
	}

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	newToken, err := security.GenerateHasuraJWT(user.ID.String(), user.Email, s.cfg.JWT.Secret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &dto.MeOutputDTO{
		UserID: user.ID.String(),
		Token:  newToken,
	}, nil
}

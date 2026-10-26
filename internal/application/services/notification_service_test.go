package services

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"go/food-recipe/internal/application/dto"
	"go/food-recipe/internal/domain/entities"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/domain/repositories"
)

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

func TestNotificationService_HandleLikeNotification(t *testing.T) {
	ctx := context.Background()
	likeID := uuid.New()
	initiatorID := uuid.New()
	ownerID := uuid.New()

	t.Run("successful insert like notification", func(t *testing.T) {
		var createdNotification *entities.Notification

		recipeRepo := &mockRecipeRepo{
			getCreatorFn: func(ctx context.Context, id uuid.UUID) (uuid.UUID, string, error) {
				if id == likeID {
					return ownerID, "chefalice", nil
				}
				return uuid.Nil, "", errors.New("not found")
			},
		}

		notifRepo := &mockNotificationRepo{
			createFn: func(ctx context.Context, n *entities.Notification) error {
				createdNotification = n
				return nil
			},
		}

		svc := NewNotificationService(recipeRepo, notifRepo)
		payload := dto.LikeEventPayloadDTO{
			Event: dto.LikeEventDTO{
				Op: "INSERT",
				Data: dto.LikeEventDataDTO{
					New: dto.LikeEventDataNewDTO{
						ID:        likeID,
						UserID:    initiatorID,
						RecipeID:  15,
						CreatedAt: "2026-10-26T12:00:00Z",
					},
				},
			},
		}

		err := svc.HandleLikeNotification(ctx, payload)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if createdNotification == nil {
			t.Fatal("expected notification to be created")
		}
		if createdNotification.UserID != ownerID {
			t.Errorf("expected UserID %v, got %v", ownerID, createdNotification.UserID)
		}
		if createdNotification.InitiatorID != initiatorID {
			t.Errorf("expected InitiatorID %v, got %v", initiatorID, createdNotification.InitiatorID)
		}
		expectedMsg := "user @chefalice liked your recipe"
		if createdNotification.Message != expectedMsg {
			t.Errorf("expected message %q, got %q", expectedMsg, createdNotification.Message)
		}
	})

	t.Run("skipped when op is not INSERT", func(t *testing.T) {
		notifCalled := false
		recipeCalled := false

		recipeRepo := &mockRecipeRepo{
			getCreatorFn: func(ctx context.Context, id uuid.UUID) (uuid.UUID, string, error) {
				recipeCalled = true
				return ownerID, "user", nil
			},
		}

		notifRepo := &mockNotificationRepo{
			createFn: func(ctx context.Context, n *entities.Notification) error {
				notifCalled = true
				return nil
			},
		}

		svc := NewNotificationService(recipeRepo, notifRepo)
		payload := dto.LikeEventPayloadDTO{
			Event: dto.LikeEventDTO{
				Op: "UPDATE",
				Data: dto.LikeEventDataDTO{
					New: dto.LikeEventDataNewDTO{
						ID:     likeID,
						UserID: initiatorID,
					},
				},
			},
		}

		err := svc.HandleLikeNotification(ctx, payload)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if recipeCalled || notifCalled {
			t.Error("expected non-insert event to be skipped without repo calls")
		}
	})

	t.Run("missing like ID", func(t *testing.T) {
		svc := NewNotificationService(&mockRecipeRepo{}, &mockNotificationRepo{})
		payload := dto.LikeEventPayloadDTO{
			Event: dto.LikeEventDTO{
				Op: "INSERT",
				Data: dto.LikeEventDataDTO{
					New: dto.LikeEventDataNewDTO{
						ID:     uuid.Nil,
						UserID: initiatorID,
					},
				},
			},
		}
		err := svc.HandleLikeNotification(ctx, payload)
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("missing initiator ID", func(t *testing.T) {
		svc := NewNotificationService(&mockRecipeRepo{}, &mockNotificationRepo{})
		payload := dto.LikeEventPayloadDTO{
			Event: dto.LikeEventDTO{
				Op: "INSERT",
				Data: dto.LikeEventDataDTO{
					New: dto.LikeEventDataNewDTO{
						ID:     likeID,
						UserID: uuid.Nil,
					},
				},
			},
		}
		err := svc.HandleLikeNotification(ctx, payload)
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("recipe creator lookup failure", func(t *testing.T) {
		recipeRepo := &mockRecipeRepo{
			getCreatorFn: func(ctx context.Context, id uuid.UUID) (uuid.UUID, string, error) {
				return uuid.Nil, "", errors.New("recipe like not found")
			},
		}
		svc := NewNotificationService(recipeRepo, &mockNotificationRepo{})
		payload := dto.LikeEventPayloadDTO{
			Event: dto.LikeEventDTO{
				Op: "INSERT",
				Data: dto.LikeEventDataDTO{
					New: dto.LikeEventDataNewDTO{
						ID:     likeID,
						UserID: initiatorID,
					},
				},
			},
		}
		err := svc.HandleLikeNotification(ctx, payload)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("notification creation failure", func(t *testing.T) {
		recipeRepo := &mockRecipeRepo{
			getCreatorFn: func(ctx context.Context, id uuid.UUID) (uuid.UUID, string, error) {
				return ownerID, "user", nil
			},
		}
		notifRepo := &mockNotificationRepo{
			createFn: func(ctx context.Context, n *entities.Notification) error {
				return errors.New("database insert failure")
			},
		}
		svc := NewNotificationService(recipeRepo, notifRepo)
		payload := dto.LikeEventPayloadDTO{
			Event: dto.LikeEventDTO{
				Op: "INSERT",
				Data: dto.LikeEventDataDTO{
					New: dto.LikeEventDataNewDTO{
						ID:     likeID,
						UserID: initiatorID,
					},
				},
			},
		}
		err := svc.HandleLikeNotification(ctx, payload)
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

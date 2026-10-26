package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go/food-recipe/internal/application/dto"
	"go/food-recipe/internal/domain/entities"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/domain/repositories"
)

type NotificationService struct {
	recipeRepo       repositories.RecipeRepository
	notificationRepo repositories.NotificationRepository
}

func NewNotificationService(
	recipeRepo repositories.RecipeRepository,
	notificationRepo repositories.NotificationRepository,
) *NotificationService {
	return &NotificationService{
		recipeRepo:       recipeRepo,
		notificationRepo: notificationRepo,
	}
}

func (s *NotificationService) HandleLikeNotification(
	ctx context.Context,
	payload dto.LikeEventPayloadDTO,
) (*dto.LikeNotificationOutputDTO, error) {
	// Only handle INSERT events; others are skipped without error
	if !strings.EqualFold(payload.Event.Op, "INSERT") {
		return &dto.LikeNotificationOutputDTO{OK: true, Skipped: true}, nil
	}

	likeID := payload.Event.Data.New.ID
	if likeID == uuid.Nil {
		return nil, fmt.Errorf("%w: like id is required", domainerrors.ErrInvalidInput)
	}

	initiatorID := payload.Event.Data.New.UserID
	if initiatorID == uuid.Nil {
		return nil, fmt.Errorf("%w: initiator user id is required", domainerrors.ErrInvalidInput)
	}

	ownerUUID, likerUsername, err := s.recipeRepo.GetCreatorFromLikeID(ctx, likeID)
	if err != nil {
		return nil, err
	}

	// Do not notify a user when liking their own recipe
	if ownerUUID == initiatorID {
		return &dto.LikeNotificationOutputDTO{OK: true, Skipped: true}, nil
	}

	notification := &entities.Notification{
		UserID:      ownerUUID,
		Message:     "user @" + likerUsername + " liked your recipe",
		InitiatorID: initiatorID,
	}

	if err := s.notificationRepo.Create(ctx, notification); err != nil {
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	return &dto.LikeNotificationOutputDTO{OK: true}, nil
}

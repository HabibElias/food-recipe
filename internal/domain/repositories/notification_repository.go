package repositories

import (
	"context"

	"go/food-recipe/internal/domain/entities"
)

type NotificationRepository interface {
	Create(ctx context.Context, n *entities.Notification) error
}

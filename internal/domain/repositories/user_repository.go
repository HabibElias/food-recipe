package repositories

import (
	"context"

	"github.com/google/uuid"
	"go/food-recipe/internal/domain/entities"
)

type UserRepository interface {
	GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	GetByUsername(ctx context.Context, username string) (*entities.User, error)
	Create(ctx context.Context, user *entities.User) (*entities.User, error)
	UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error
}

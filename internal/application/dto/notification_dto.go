package dto

import (
	"go/food-recipe/internal/pkg/validator"

	"github.com/google/uuid"
)

type LikeEventDataNewDTO struct {
	ID        uuid.UUID `json:"id" validate:"required"`
	UserID    uuid.UUID `json:"user_id" validate:"required"`
	RecipeID  int       `json:"recipe_id" validate:"required,gt=0"`
	CreatedAt string    `json:"created_at"`
}

func (d LikeEventDataNewDTO) Validate() error {
	return validator.ValidateStruct(d)
}

type LikeEventDataDTO struct {
	New LikeEventDataNewDTO `json:"new" validate:"required"`
}

type LikeEventDTO struct {
	Op   string           `json:"op"`
	Data LikeEventDataDTO `json:"data" validate:"required"`
}

type LikeEventPayloadDTO struct {
	Event LikeEventDTO `json:"event" validate:"required"`
}

func (d LikeEventPayloadDTO) Validate() error {
	return validator.ValidateStruct(d)
}

type LikeNotificationOutputDTO struct {
	OK      bool `json:"ok"`
	Skipped bool `json:"skipped,omitempty"`
}

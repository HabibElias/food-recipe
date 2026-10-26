package dto

import "github.com/google/uuid"

type LikeEventDataNewDTO struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	RecipeID  int       `json:"recipe_id"`
	CreatedAt string    `json:"created_at"`
}

type LikeEventDataDTO struct {
	New LikeEventDataNewDTO `json:"new"`
}

type LikeEventDTO struct {
	Op   string           `json:"op"`
	Data LikeEventDataDTO `json:"data"`
}

type LikeEventPayloadDTO struct {
	Event LikeEventDTO `json:"event"`
}

type LikeNotificationOutputDTO struct {
	OK      bool `json:"ok"`
	Skipped bool `json:"skipped,omitempty"`
}

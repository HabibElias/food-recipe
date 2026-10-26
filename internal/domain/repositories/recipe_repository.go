package repositories

import (
	"context"

	"github.com/google/uuid"
	"go/food-recipe/internal/domain/entities"
)

type RecipeRepository interface {
	GetCreatorFromLikeID(ctx context.Context, likeID uuid.UUID) (recipeOwnerID uuid.UUID, likerUsername string, err error)
	AddRecipeImage(ctx context.Context, img *entities.RecipeImage) error
}

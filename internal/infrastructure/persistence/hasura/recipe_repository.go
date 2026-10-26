package hasura

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go/food-recipe/internal/domain/entities"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/domain/repositories"
)

type recipeRepository struct {
	client *Client
}

var _ repositories.RecipeRepository = (*recipeRepository)(nil)

func NewRecipeRepository(client *Client) repositories.RecipeRepository {
	return &recipeRepository{client: client}
}

func (r *recipeRepository) GetCreatorFromLikeID(ctx context.Context, likeID uuid.UUID) (recipeOwnerID uuid.UUID, likerUsername string, err error) {
	if r.client == nil || r.client.GQL == nil {
		return uuid.Nil, "", domainerrors.ErrServiceUnavailable
	}

	var query struct {
		Like *struct {
			ID     string `graphql:"id"`
			Recipe struct {
				UserID string `graphql:"user_id"`
			} `graphql:"recipe"`
			User struct {
				ID       string `graphql:"id"`
				Username string `graphql:"username"`
			} `graphql:"user"`
		} `graphql:"recipe_like_by_pk(id: $id)"`
	}

	variables := map[string]interface{}{
		"id": GQLUUID(likeID),
	}

	if err := r.client.GQL.Query(ctx, &query, variables); err != nil {
		return uuid.Nil, "", err
	}

	if query.Like == nil || query.Like.ID == "" {
		return uuid.Nil, "", fmt.Errorf("recipe like not found")
	}

	ownerUUID, err := uuid.Parse(query.Like.Recipe.UserID)
	if err != nil {
		return uuid.Nil, "", fmt.Errorf("invalid recipe owner id: %w", err)
	}

	return ownerUUID, query.Like.User.Username, nil
}

func (r *recipeRepository) AddRecipeImage(ctx context.Context, img *entities.RecipeImage) error {
	if r.client == nil || r.client.GQL == nil {
		return domainerrors.ErrServiceUnavailable
	}
	if img == nil {
		return domainerrors.ErrInvalidInput
	}

	var mutation struct {
		InsertRecipeImage struct {
			Returning []struct {
				ID string `graphql:"id"`
			} `graphql:"returning"`
		} `graphql:"insert_recipe_image(objects: {img_url: $img_url, recipe_id: $recipe_id, is_thumbnail: $is_thumbnail})"`
	}

	variables := map[string]interface{}{
		"img_url":      img.ImageURL,
		"recipe_id":    img.RecipeID,
		"is_thumbnail": img.IsThumbnail,
	}

	if err := r.client.GQL.Mutate(ctx, &mutation, variables); err != nil {
		return err
	}

	if len(mutation.InsertRecipeImage.Returning) == 0 {
		return fmt.Errorf("failed to save recipe image")
	}

	img.ID = mutation.InsertRecipeImage.Returning[0].ID
	return nil
}

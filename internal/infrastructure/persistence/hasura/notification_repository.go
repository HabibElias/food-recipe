package hasura

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go/food-recipe/internal/domain/entities"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/domain/repositories"
)

type notificationRepository struct {
	client *Client
}

var _ repositories.NotificationRepository = (*notificationRepository)(nil)

func NewNotificationRepository(client *Client) repositories.NotificationRepository {
	return &notificationRepository{client: client}
}

func (r *notificationRepository) Create(ctx context.Context, n *entities.Notification) error {
	if r.client == nil || r.client.GQL == nil {
		return domainerrors.ErrServiceUnavailable
	}
	if n == nil {
		return domainerrors.ErrInvalidInput
	}

	var mutation struct {
		InsertNotifications struct {
			Returning []struct {
				ID uuid.UUID `graphql:"id"`
			} `graphql:"returning"`
		} `graphql:"insert_notifications(objects: {user_id: $user_id, message: $message, initiator_id: $initiator_id})"`
	}

	variables := map[string]interface{}{
		"user_id":      GQLUUID(n.UserID),
		"message":      n.Message,
		"initiator_id": GQLUUID(n.InitiatorID),
	}

	if err := r.client.GQL.Mutate(ctx, &mutation, variables); err != nil {
		return err
	}

	if len(mutation.InsertNotifications.Returning) == 0 {
		return fmt.Errorf("failed to create notification")
	}

	n.ID = mutation.InsertNotifications.Returning[0].ID
	return nil
}

package hasura

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"go/food-recipe/internal/domain/entities"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/domain/repositories"
)

type userRepository struct {
	client *Client
}

var _ repositories.UserRepository = (*userRepository)(nil)

func NewUserRepository(client *Client) repositories.UserRepository {
	return &userRepository{client: client}
}

type gqlUser struct {
	ID        uuid.UUID `graphql:"id"`
	Email     string    `graphql:"email"`
	Password  string    `graphql:"password"`
	Username  string    `graphql:"username"`
	FirstName string    `graphql:"FirstName"`
	LastName  string    `graphql:"LastName"`
	AvatarURL string    `graphql:"avatar_url"`
}

func mapGQLUserToEntity(u *gqlUser) *entities.User {
	if u == nil {
		return nil
	}
	return &entities.User{
		ID:           u.ID,
		Email:        u.Email,
		Username:     u.Username,
		FirstName:    u.FirstName,
		LastName:     u.LastName,
		PasswordHash: u.Password,
		AvatarURL:    u.AvatarURL,
	}
}

func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	if r.client == nil || r.client.GQL == nil {
		return nil, domainerrors.ErrServiceUnavailable
	}

	var query struct {
		User *gqlUser `graphql:"user_by_pk(id: $id)"`
	}
	variables := map[string]interface{}{
		"id": GQLUUID(id),
	}

	if err := r.client.GQL.Query(ctx, &query, variables); err != nil {
		return nil, err
	}

	if query.User == nil || query.User.ID == uuid.Nil {
		return nil, domainerrors.ErrUserNotFound
	}

	return mapGQLUserToEntity(query.User), nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	if r.client == nil || r.client.GQL == nil {
		return nil, domainerrors.ErrServiceUnavailable
	}

	var query struct {
		Users []gqlUser `graphql:"user(where: {email: {_eq: $email}})"`
	}
	variables := map[string]interface{}{
		"email": email,
	}

	if err := r.client.GQL.Query(ctx, &query, variables); err != nil {
		return nil, err
	}

	if len(query.Users) == 0 {
		return nil, domainerrors.ErrUserNotFound
	}

	return mapGQLUserToEntity(&query.Users[0]), nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*entities.User, error) {
	if r.client == nil || r.client.GQL == nil {
		return nil, domainerrors.ErrServiceUnavailable
	}

	var query struct {
		Users []gqlUser `graphql:"user(where: {username: {_eq: $username}})"`
	}
	variables := map[string]interface{}{
		"username": username,
	}

	if err := r.client.GQL.Query(ctx, &query, variables); err != nil {
		return nil, err
	}

	if len(query.Users) == 0 {
		return nil, domainerrors.ErrUserNotFound
	}

	return mapGQLUserToEntity(&query.Users[0]), nil
}

func (r *userRepository) Create(ctx context.Context, user *entities.User) (*entities.User, error) {
	if r.client == nil || r.client.GQL == nil {
		return nil, domainerrors.ErrServiceUnavailable
	}
	if user == nil {
		return nil, domainerrors.ErrInvalidInput
	}

	var mutation struct {
		InsertUsers struct {
			Returning []gqlUser `graphql:"returning"`
		} `graphql:"insert_user(objects: {email: $email, password: $password, username: $username, FirstName: $first_name, LastName: $last_name})"`
	}

	variables := map[string]interface{}{
		"email":      user.Email,
		"password":   user.PasswordHash,
		"username":   user.Username,
		"first_name": user.FirstName,
		"last_name":  user.LastName,
	}

	if err := r.client.GQL.Mutate(ctx, &mutation, variables); err != nil {
		return nil, err
	}

	if len(mutation.InsertUsers.Returning) == 0 {
		return nil, fmt.Errorf("failed to create user")
	}

	created := mapGQLUserToEntity(&mutation.InsertUsers.Returning[0])
	if created.PasswordHash == "" {
		created.PasswordHash = user.PasswordHash
	}
	return created, nil
}

func (r *userRepository) UpdateAvatar(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	if r.client == nil || r.client.GQL == nil {
		return domainerrors.ErrServiceUnavailable
	}

	var mutation struct {
		UpdateUsers struct {
			Returning []struct {
				ID uuid.UUID `graphql:"id"`
			} `graphql:"returning"`
		} `graphql:"update_user(where: {id: {_eq: $id}}, _set: {avatar_url: $avatar_url})"`
	}

	variables := map[string]interface{}{
		"id":         GQLUUID(userID),
		"avatar_url": avatarURL,
	}

	if err := r.client.GQL.Mutate(ctx, &mutation, variables); err != nil {
		return err
	}

	if len(mutation.UpdateUsers.Returning) == 0 {
		return domainerrors.ErrUserNotFound
	}

	return nil
}

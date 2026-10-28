package dto

import (
	"time"

	"go/food-recipe/internal/pkg/validator"
)

// Validatable represents an input DTO that can validate itself.
type Validatable interface {
	Validate() error
}

type LoginInputDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

func (d LoginInputDTO) Validate() error {
	return validator.ValidateStruct(d)
}

type LoginOutputDTO struct {
	UserID       string         `json:"user_id"`
	Token        string         `json:"token"`
	LoggedInUser AuthProfileDTO `json:"loggedInUser"`
}

// AuthProfileDTO contains only fields safe to return to the authenticated user.
// Keep it separate from the persistence entity, which also holds a password hash.
type AuthProfileDTO struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Username  string    `json:"username"`
	FirstName string    `json:"FirstName"`
	LastName  string    `json:"LastName"`
	AvatarURL string    `json:"avatar_url"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type SignupInputDTO struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
	Email     string `json:"email" validate:"required,email"`
	Username  string `json:"username" validate:"required,min=5,max=50"`
	Password  string `json:"password" validate:"required,min=6,max=100"`
}

func (d SignupInputDTO) Validate() error {
	return validator.ValidateStruct(d)
}

type SignupOutputDTO struct {
	UserID      string         `json:"user_id"`
	Token       string         `json:"token"`
	CreatedUser AuthProfileDTO `json:"createdUser"`
}

type MeInputDTO struct {
	Token string `json:"token" validate:"required"`
}

func (d MeInputDTO) Validate() error {
	return validator.ValidateStruct(d)
}

type MeOutputDTO struct {
	UserID string         `json:"user_id"`
	Token  string         `json:"token"`
	User   AuthProfileDTO `json:"user"`
}

type LoginActionPayloadDTO struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            struct {
		Input LoginInputDTO `json:"input"`
	} `json:"input"`
}

type SignupActionPayloadDTO struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            struct {
		Input SignupInputDTO `json:"input"`
	} `json:"input"`
}

type MeActionPayloadDTO struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            struct {
		Input MeInputDTO `json:"input"`
	} `json:"input"`
}

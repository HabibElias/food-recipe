package dto

import "go/food-recipe/internal/presentation/validator"

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
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

type SignupInputDTO struct {
	FirstName string `json:"first_name" validate:"required,min=2,max=100"`
	LastName  string `json:"last_name" validate:"required,min=2,max=100"`
	Email     string `json:"email" validate:"required,email"`
	Username  string `json:"username" validate:"required,min=3,max=50"`
	Password  string `json:"password" validate:"required,min=6,max=100"`
}

func (d SignupInputDTO) Validate() error {
	return validator.ValidateStruct(d)
}

type SignupOutputDTO struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

type MeInputDTO struct {
	Token string `json:"token" validate:"required"`
}

func (d MeInputDTO) Validate() error {
	return validator.ValidateStruct(d)
}

type MeOutputDTO struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
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

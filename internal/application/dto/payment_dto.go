package dto

import (
	"encoding/json"

	"go/food-recipe/internal/pkg/validator"
)

type InitializePaymentInputDTO struct {
	Amount   float64 `json:"amount" validate:"required,gt=0"`
	RecipeID int     `json:"recipe_id" validate:"required,gt=0"`
	Phone    string  `json:"phone,omitempty" validate:"omitempty,min=9"`
}

func (d InitializePaymentInputDTO) Validate() error {
	return validator.ValidateStruct(d)
}

type InitializePaymentOutputDTO struct {
	CheckoutURL string `json:"checkout_url"`
	TxRef       string `json:"tx_ref"`
	Status      string `json:"status"`
}

type VerifyPaymentInputDTO struct {
	TrxRef string `json:"trx_ref" validate:"required"`
}

func (d VerifyPaymentInputDTO) Validate() error {
	return validator.ValidateStruct(d)
}

type VerifyPaymentOutputDTO struct {
	Message string `json:"message"`
	TxRef   string `json:"tx_ref"`
	Status  string `json:"status"`
}

type InitializePaymentActionPayloadDTO struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            struct {
		Amount   float64                    `json:"amount"`
		RecipeID int                        `json:"recipe_id"`
		Phone    string                     `json:"phone,omitempty"`
		Input    *InitializePaymentInputDTO `json:"input,omitempty"`
	} `json:"input"`
}

func (p InitializePaymentActionPayloadDTO) GetSessionUserID() string {
	if p.SessionVariables != nil {
		if v, ok := p.SessionVariables["x-hasura-user-id"]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

func (p InitializePaymentActionPayloadDTO) ToInputDTO() InitializePaymentInputDTO {
	amount := p.Input.Amount
	recipeID := p.Input.RecipeID
	phone := p.Input.Phone
	if p.Input.Input != nil {
		if p.Input.Input.Amount != 0 {
			amount = p.Input.Input.Amount
		}
		if p.Input.Input.RecipeID != 0 {
			recipeID = p.Input.Input.RecipeID
		}
		if p.Input.Input.Phone != "" {
			phone = p.Input.Input.Phone
		}
	}
	return InitializePaymentInputDTO{
		Amount:   amount,
		RecipeID: recipeID,
		Phone:    phone,
	}
}

func (p *InitializePaymentActionPayloadDTO) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if svRaw, ok := raw["session_variables"]; ok {
		if err := json.Unmarshal(svRaw, &p.SessionVariables); err != nil {
			return err
		}
	}
	if inRaw, ok := raw["input"]; ok {
		if err := json.Unmarshal(inRaw, &p.Input); err != nil {
			return err
		}
	}
	return nil
}

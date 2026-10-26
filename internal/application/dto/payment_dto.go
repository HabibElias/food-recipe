package dto

import "encoding/json"

type InitializePaymentInputDTO struct {
	Amount   float64 `json:"amount"`
	RecipeID int     `json:"recipe_id"`
}

type InitializePaymentOutputDTO struct {
	CheckoutURL string `json:"checkout_url"`
	TxRef       string `json:"tx_ref"`
	Status      string `json:"status"`
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
	if p.Input.Input != nil {
		if p.Input.Input.Amount != 0 {
			amount = p.Input.Input.Amount
		}
		if p.Input.Input.RecipeID != 0 {
			recipeID = p.Input.Input.RecipeID
		}
	}
	return InitializePaymentInputDTO{
		Amount:   amount,
		RecipeID: recipeID,
	}
}

func (p *InitializePaymentActionPayloadDTO) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if svRaw, ok := raw["session_variables"]; ok {
		_ = json.Unmarshal(svRaw, &p.SessionVariables)
	}
	if inRaw, ok := raw["input"]; ok {
		_ = json.Unmarshal(inRaw, &p.Input)
	}
	return nil
}

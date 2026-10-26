package dto

type LoginInputDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginOutputDTO struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

type SignupInputDTO struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Username  string `json:"username"`
	Password  string `json:"password"`
}

type SignupOutputDTO struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

type MeInputDTO struct {
	Token string `json:"token"`
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

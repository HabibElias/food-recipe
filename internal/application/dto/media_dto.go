package dto

type RecipeImageItemDTO struct {
	ImageName      string `json:"image_name"`
	ImageType      string `json:"image_type"`
	ImageBase64Str string `json:"image_base64str"`
	IsThumbnail    bool   `json:"is_thumbnail"`
}

type UploadRecipeImagesInputDTO struct {
	RecipeID     int                  `json:"recipe_id"`
	RecipeImages []RecipeImageItemDTO `json:"recipe_images"`
}

type UploadRecipeImagesOutputDTO struct {
	Success bool `json:"success"`
}

type UploadRecipeImagesActionPayloadDTO struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            struct {
		Input UploadRecipeImagesInputDTO `json:"input"`
	} `json:"input"`
}

type UserImageItemDTO struct {
	ImageName      string `json:"image_name"`
	ImageType      string `json:"image_type"`
	ImageBase64Str string `json:"image_base64str"`
}

type UploadUserImageInputDTO struct {
	UserID string           `json:"user_id"`
	Image  UserImageItemDTO `json:"image"`
}

type UploadUserImageOutputDTO struct {
	Success  bool   `json:"success"`
	ImageURL string `json:"image_url,omitempty"`
}

type UploadUserImageActionPayloadDTO struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            struct {
		Input UploadUserImageInputDTO `json:"input"`
	} `json:"input"`
}

func (p UploadUserImageActionPayloadDTO) GetSessionUserID() string {
	if p.SessionVariables != nil {
		if v, ok := p.SessionVariables["x-hasura-user-id"]; ok {
			if s, ok := v.(string); ok {
				return s
			}
		}
	}
	return ""
}

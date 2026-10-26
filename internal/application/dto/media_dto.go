package dto

import "go/food-recipe/internal/presentation/validator"

type RecipeImageItemDTO struct {
	ImageName      string `json:"image_name" validate:"required"`
	ImageType      string `json:"image_type" validate:"required"`
	ImageBase64Str string `json:"image_base64str" validate:"required"`
	IsThumbnail    bool   `json:"is_thumbnail"`
}

type UploadRecipeImagesInputDTO struct {
	RecipeID     int                  `json:"recipe_id" validate:"required,gt=0"`
	RecipeImages []RecipeImageItemDTO `json:"recipe_images" validate:"required,min=1,dive"`
}

func (d UploadRecipeImagesInputDTO) Validate() error {
	return validator.ValidateStruct(d)
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
	ImageName      string `json:"image_name" validate:"required"`
	ImageType      string `json:"image_type" validate:"required"`
	ImageBase64Str string `json:"image_base64str" validate:"required"`
}

type UploadUserImageInputDTO struct {
	UserID string           `json:"user_id" validate:"required,uuid"`
	Image  UserImageItemDTO `json:"image" validate:"required"`
}

func (d UploadUserImageInputDTO) Validate() error {
	return validator.ValidateStruct(d)
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

package handlers

import (
	"encoding/json"
	"fmt"
	"go/food-recipe/clients/cloudinary"
	"go/food-recipe/clients/hasura"
	"go/food-recipe/helpers"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ImageInput struct {
	ImageName      string `json:"image_name"`
	ImageType      string `json:"image_type"`
	ImageBase64Str string `json:"image_base64str"`
}

type UploadUserImageInput struct {
	UserID string     `json:"user_id"`
	Image  ImageInput `json:"Image"`
}

type UploadUserImageArgs struct {
	Input UploadUserImageInput `json:"input"`
}

type UploadUserImageActionPayload struct {
	SessionVariables map[string]interface{} `json:"session_variables"`
	Input            UploadUserImageArgs    `json:"input"`
}

type UploadUserImageOutput struct {
	Success bool `json:"success"`
}

// UploadUserImageHandler handles Hasura action to upload a single user image
func UploadUserImageHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	var payload UploadUserImageActionPayload
	if err := json.Unmarshal(body, &payload); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	in := payload.Input.Input

	// determine user id: prefer input.user_id, fallback to session x-hasura-user-id
	userIDStr := strings.TrimSpace(in.UserID)
	if userIDStr == "" {
		if v, ok := payload.SessionVariables["x-hasura-user-id"]; ok {
			if s, ok2 := v.(string); ok2 {
				userIDStr = s
			}
		}
	}

	if userIDStr == "" {
		helpers.RespondWithError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	// validate image input using shared helper from uploadRecipeImagesHandler
	img := RecipeImageInput{
		ImageName:      in.Image.ImageName,
		ImageType:      in.Image.ImageType,
		ImageBase64Str: in.Image.ImageBase64Str,
		IsThumbnail:    false,
	}

	if err := validateImageInput(img); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Prepare base64 string
	base64Data := img.ImageBase64Str
	if !strings.HasPrefix(base64Data, "data:") {
		mimeType := getMimeType(img.ImageType)
		base64Data = fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
	}

	// Cloudinary upload
	cld := cloudinary.Get()
	if cld == nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "image service unavailable")
		return
	}

	publicID := fmt.Sprintf("user_%s_%s_%d", userIDStr, sanitizeFileName(img.ImageName), time.Now().UnixNano())
	imageURL, err := cloudinary.UploadBase64Image(r.Context(), base64Data, publicID, "users")
	if err != nil {
		log.Printf("failed to upload user image: %v", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to upload image")
		return
	}

	// Update user's avatar_url in Hasura
	client := hasura.Get()
	if client == nil {
		helpers.RespondWithError(w, http.StatusInternalServerError, "database service unavailable")
		return
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid user_id")
		return
	}

	vars := map[string]interface{}{
		"id":         hasura.UUID(uid),
		"avatar_url": imageURL,
	}

	mut := hasura.UpdateUserAvatar
	if err := client.Mutate(r.Context(), &mut, vars); err != nil {
		log.Printf("failed to update user avatar: %v", err)
		helpers.RespondWithError(w, http.StatusInternalServerError, "failed to save avatar")
		return
	}

	helpers.RespondWithJson(w, http.StatusOK, UploadUserImageOutput{Success: true})
}

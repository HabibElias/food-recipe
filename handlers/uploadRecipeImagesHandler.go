package handlers

import (
	"context"
	"encoding/base64"
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
)

// RecipeImageInput represents a single recipe image input
type RecipeImageInput struct {
	ImageName      string `json:"image_name"`
	ImageType      string `json:"image_type"`
	ImageBase64Str string `json:"image_base64str"`
	IsThumbnail    bool   `json:"is_thumbnail"`
}

// UploadRecipeImagesInput represents the Hasura action input
type UploadRecipeImagesInput struct {
	RecipeID     int                `json:"recipe_id"`
	RecipeImages []RecipeImageInput `json:"recipe_images"`
}

// UploadRecipeImagesArgs represents the intermediate input wrapper
type UploadRecipeImagesArgs struct {
	Input UploadRecipeImagesInput `json:"input"`
}

// UploadRecipeImagesActionPayload represents the full Hasura action payload
type UploadRecipeImagesActionPayload struct {
	SessionVariables map[string]interface{}  `json:"session_variables"`
	Input            UploadRecipeImagesArgs  `json:"input"`
}

// UploadRecipeImagesOutput represents the Hasura action output
type UploadRecipeImagesOutput struct {
	Success bool `json:"success"`
}

// UploadRecipeImagesHandler handles the Hasura action for uploading recipe images
func UploadRecipeImagesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Read request body
	reqBody, err := io.ReadAll(r.Body)
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	// Parse the body as action payload
	var actionPayload UploadRecipeImagesActionPayload
	if err := json.Unmarshal(reqBody, &actionPayload); err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, "invalid payload")
		return
	}

	// Validate input
	if actionPayload.Input.Input.RecipeID <= 0 {
		helpers.RespondWithError(w, http.StatusBadRequest, "recipe_id must be greater than zero")
		return
	}

	if len(actionPayload.Input.Input.RecipeImages) == 0 {
		helpers.RespondWithError(w, http.StatusBadRequest, "at least one recipe image is required")
		return
	}

	// Process the upload
	result, err := uploadRecipeImages(actionPayload.Input.Input, r.Context())
	if err != nil {
		helpers.RespondWithError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Write the response as JSON
	helpers.RespondWithJson(w, http.StatusOK, result)
}

// uploadRecipeImages processes the image uploads and inserts them into Hasura
func uploadRecipeImages(input UploadRecipeImagesInput, ctx context.Context) (UploadRecipeImagesOutput, error) {
	// Validate Cloudinary client
	cldClient := cloudinary.Get()
	if cldClient == nil {
		return UploadRecipeImagesOutput{Success: false}, fmt.Errorf("image upload service temporarily unavailable")
	}

	// Validate Hasura client
	hasuraClient := hasura.Get()
	if hasuraClient == nil {
		return UploadRecipeImagesOutput{Success: false}, fmt.Errorf("database service temporarily unavailable")
	}

	// Process each image
	for i, img := range input.RecipeImages {
		// Validate image input
		if err := validateImageInput(img); err != nil {
			return UploadRecipeImagesOutput{Success: false}, fmt.Errorf("image %d: %w", i+1, err)
		}

		// Prepare base64 string with data URI prefix if not present
		base64Data := img.ImageBase64Str
		if !strings.HasPrefix(base64Data, "data:") {
			// Add data URI prefix based on image type
			mimeType := getMimeType(img.ImageType)
			base64Data = fmt.Sprintf("data:%s;base64,%s", mimeType, base64Data)
		}

		// Generate unique public ID for the image
		publicID := fmt.Sprintf("recipe_%d_%s_%d", input.RecipeID, sanitizeFileName(img.ImageName), time.Now().UnixNano())

		// Upload to Cloudinary
		imageURL, err := cloudinary.UploadBase64Image(ctx, base64Data, publicID, "recipes")
		if err != nil {
			log.Printf("Failed to upload image %d to Cloudinary: %v", i+1, err)
			return UploadRecipeImagesOutput{Success: false}, fmt.Errorf("failed to upload image %d: %w", i+1, err)
		}

		// Insert into Hasura
		insertVariables := map[string]interface{}{
			"img_url":      imageURL,
			"recipe_id":    input.RecipeID,
			"is_thumbnail": img.IsThumbnail,
		}

		insertMutation := hasura.InsertRecipeImage
		if err := hasuraClient.Mutate(ctx, &insertMutation, insertVariables); err != nil {
			log.Printf("Failed to insert recipe image %d: %v", i+1, err)
			return UploadRecipeImagesOutput{Success: false}, fmt.Errorf("failed to save image %d to database", i+1)
		}

		if len(insertMutation.InsertRecipeImage.Returning) == 0 {
			return UploadRecipeImagesOutput{Success: false}, fmt.Errorf("failed to save image %d to database", i+1)
		}

		log.Printf("Successfully uploaded and saved image %d for recipe %d", i+1, input.RecipeID)
	}

	return UploadRecipeImagesOutput{Success: true}, nil
}

// validateImageInput validates a single image input
func validateImageInput(img RecipeImageInput) error {
	if strings.TrimSpace(img.ImageName) == "" {
		return fmt.Errorf("image_name is required")
	}

	if strings.TrimSpace(img.ImageType) == "" {
		return fmt.Errorf("image_type is required")
	}

	// Validate image type
	validTypes := map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
		"gif":  true,
		"webp": true,
	}

	// Handle both MIME types (e.g., "image/png") and file extensions (e.g., "png")
	imgType := strings.ToLower(strings.TrimSpace(img.ImageType))
	
	// If it's a MIME type, extract the extension
	if strings.Contains(imgType, "/") {
		parts := strings.Split(imgType, "/")
		if len(parts) > 1 {
			imgType = parts[1]
			// Handle "jpeg" in MIME type "image/jpeg"
			if imgType == "jpeg" {
				imgType = "jpg"
			}
		}
	}
	
	if !validTypes[imgType] {
		return fmt.Errorf("unsupported image type: %s (supported: jpg, jpeg, png, gif, webp)", img.ImageType)
	}

	if strings.TrimSpace(img.ImageBase64Str) == "" {
		return fmt.Errorf("image_base64str is required")
	}

	// Validate base64 string (remove data URI prefix if present)
	base64Data := img.ImageBase64Str
	if strings.Contains(base64Data, ",") {
		parts := strings.SplitN(base64Data, ",", 2)
		if len(parts) == 2 {
			base64Data = parts[1]
		}
	}

	// Try to decode to validate it's valid base64
	if _, err := base64.StdEncoding.DecodeString(base64Data); err != nil {
		return fmt.Errorf("invalid base64 string")
	}

	return nil
}

// getMimeType returns the MIME type for a given image type
// Handles both MIME types (e.g., "image/png") and file extensions (e.g., "png")
func getMimeType(imgType string) string {
	imgType = strings.ToLower(strings.TrimSpace(imgType))
	
	// If it's already a MIME type, return it as-is
	if strings.HasPrefix(imgType, "image/") {
		return imgType
	}
	
	// Otherwise, convert extension to MIME type
	mimeTypes := map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"webp": "image/webp",
	}

	if mimeType, ok := mimeTypes[imgType]; ok {
		return mimeType
	}

	return "image/jpeg" // default
}

// sanitizeFileName removes invalid characters from file name for Cloudinary public ID
func sanitizeFileName(name string) string {
	// Remove file extension if present
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[:idx]
	}

	// Replace invalid characters with underscores
	invalidChars := []string{"/", "\\", " ", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}

	// Limit length
	if len(result) > 50 {
		result = result[:50]
	}

	return result
}

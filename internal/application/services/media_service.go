package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"go/food-recipe/internal/application/dto"
	"go/food-recipe/internal/domain/entities"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/domain/ports"
	"go/food-recipe/internal/domain/repositories"
)

const maxImageSizeBytes = 10 * 1024 * 1024 // 10MB

type MediaService struct {
	storage    ports.FileStorage
	recipeRepo repositories.RecipeRepository
	userRepo   repositories.UserRepository
}

func NewMediaService(
	storage ports.FileStorage,
	recipeRepo repositories.RecipeRepository,
	userRepo repositories.UserRepository,
) *MediaService {
	return &MediaService{
		storage:    storage,
		recipeRepo: recipeRepo,
		userRepo:   userRepo,
	}
}

func (s *MediaService) UploadRecipeImages(
	ctx context.Context,
	input dto.UploadRecipeImagesInputDTO,
) (*dto.UploadRecipeImagesOutputDTO, error) {
	if input.RecipeID <= 0 {
		return nil, fmt.Errorf("%w: recipe_id must be greater than zero", domainerrors.ErrInvalidInput)
	}

	if len(input.RecipeImages) == 0 {
		return nil, fmt.Errorf("%w: at least one recipe image is required", domainerrors.ErrInvalidInput)
	}

	for i, img := range input.RecipeImages {
		formattedURI, err := validateAndFormatImage(img.ImageName, img.ImageType, img.ImageBase64Str)
		if err != nil {
			return nil, fmt.Errorf("%w: image %d: %v", domainerrors.ErrInvalidInput, i+1, err)
		}

		publicID := fmt.Sprintf("recipe_%d_%s_%d", input.RecipeID, sanitizeFileName(img.ImageName), time.Now().UnixNano())

		imageURL, err := s.storage.UploadBase64Image(ctx, formattedURI, publicID, "recipes")
		if err != nil {
			return nil, fmt.Errorf("failed to upload image %d: %w", i+1, err)
		}

		recipeImg := &entities.RecipeImage{
			RecipeID:    input.RecipeID,
			ImageURL:    imageURL,
			IsThumbnail: img.IsThumbnail,
		}

		if err := s.recipeRepo.AddRecipeImage(ctx, recipeImg); err != nil {
			return nil, fmt.Errorf("failed to save image %d to database: %w", i+1, err)
		}
	}

	return &dto.UploadRecipeImagesOutputDTO{Success: true}, nil
}

func (s *MediaService) UploadUserImage(
	ctx context.Context,
	input dto.UploadUserImageInputDTO,
	sessionUserID string,
) (*dto.UploadUserImageOutputDTO, error) {
	userIDStr := strings.TrimSpace(input.UserID)
	if userIDStr == "" {
		userIDStr = strings.TrimSpace(sessionUserID)
	}

	if userIDStr == "" {
		return nil, fmt.Errorf("%w: user_id is required", domainerrors.ErrInvalidInput)
	}

	uid, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid user_id", domainerrors.ErrInvalidInput)
	}

	formattedURI, err := validateAndFormatImage(input.Image.ImageName, input.Image.ImageType, input.Image.ImageBase64Str)
	if err != nil {
		return nil, err
	}

	publicID := fmt.Sprintf("user_%s_%s_%d", uid.String(), sanitizeFileName(input.Image.ImageName), time.Now().UnixNano())

	imageURL, err := s.storage.UploadBase64Image(ctx, formattedURI, publicID, "users")
	if err != nil {
		return nil, fmt.Errorf("failed to upload user image: %w", err)
	}

	if err := s.userRepo.UpdateAvatar(ctx, uid, imageURL); err != nil {
		return nil, err
	}

	return &dto.UploadUserImageOutputDTO{
		Success:  true,
		ImageURL: imageURL,
	}, nil
}

func validateAndFormatImage(imageName, imageType, imageBase64Str string) (string, error) {
	if strings.TrimSpace(imageName) == "" {
		return "", fmt.Errorf("%w: image_name is required", domainerrors.ErrInvalidInput)
	}

	imgType := strings.ToLower(strings.TrimSpace(imageType))
	if imgType == "" {
		return "", fmt.Errorf("%w: image_type is required", domainerrors.ErrInvalidInput)
	}

	// Handle MIME type e.g. "image/png"
	cleanExt := imgType
	if strings.Contains(cleanExt, "/") {
		parts := strings.Split(cleanExt, "/")
		if len(parts) > 1 {
			cleanExt = parts[1]
			if cleanExt == "jpeg" {
				cleanExt = "jpg"
			}
		}
	}

	validTypes := map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
		"gif":  true,
		"webp": true,
	}

	if !validTypes[cleanExt] {
		return "", fmt.Errorf("%w: unsupported image type: %s (supported: jpg, jpeg, png, gif, webp)", domainerrors.ErrInvalidInput, imageType)
	}

	base64Raw := strings.TrimSpace(imageBase64Str)
	if base64Raw == "" {
		return "", fmt.Errorf("%w: image_base64str is required", domainerrors.ErrInvalidInput)
	}

	payload := base64Raw
	if strings.Contains(payload, ",") {
		parts := strings.SplitN(payload, ",", 2)
		if len(parts) == 2 {
			payload = parts[1]
		}
	}

	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		return "", fmt.Errorf("%w: invalid base64 string", domainerrors.ErrInvalidInput)
	}

	if len(decoded) > maxImageSizeBytes {
		return "", fmt.Errorf("%w: image exceeds maximum size limit of 10MB", domainerrors.ErrInvalidInput)
	}

	mimeType := getMimeType(cleanExt)
	formattedURI := fmt.Sprintf("data:%s;base64,%s", mimeType, payload)
	return formattedURI, nil
}

func getMimeType(ext string) string {
	ext = strings.ToLower(strings.TrimSpace(ext))
	if strings.HasPrefix(ext, "image/") {
		return ext
	}

	mimeTypes := map[string]string{
		"jpg":  "image/jpeg",
		"jpeg": "image/jpeg",
		"png":  "image/png",
		"gif":  "image/gif",
		"webp": "image/webp",
	}

	if mt, ok := mimeTypes[ext]; ok {
		return mt
	}
	return "image/jpeg"
}

func sanitizeFileName(name string) string {
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[:idx]
	}

	invalidChars := []string{"/", "\\", " ", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, char := range invalidChars {
		result = strings.ReplaceAll(result, char, "_")
	}

	if len(result) > 50 {
		result = result[:50]
	}

	return result
}

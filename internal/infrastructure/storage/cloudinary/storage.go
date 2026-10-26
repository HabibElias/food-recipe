package cloudinary

import (
	"context"
	"fmt"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
	"go/food-recipe/internal/config"
	"go/food-recipe/internal/domain/ports"
)

type CloudinaryStorage struct {
	client *cloudinary.Cloudinary
}

func NewCloudinaryStorage(cfg config.CloudinaryConfig) (*CloudinaryStorage, error) {
	if cfg.CloudName == "" || cfg.APIKey == "" || cfg.APISecret == "" {
		return &CloudinaryStorage{client: nil}, nil
	}

	client, err := cloudinary.NewFromParams(cfg.CloudName, cfg.APIKey, cfg.APISecret)
	if err != nil {
		return nil, fmt.Errorf("failed to init cloudinary: %w", err)
	}

	return &CloudinaryStorage{client: client}, nil
}

var _ ports.FileStorage = (*CloudinaryStorage)(nil)

func (s *CloudinaryStorage) UploadBase64Image(ctx context.Context, base64Data, publicID, folder string) (string, error) {
	if s == nil || s.client == nil {
		return "", fmt.Errorf("cloudinary client not initialized")
	}
	if base64Data == "" {
		return "", fmt.Errorf("base64 data cannot be empty")
	}

	result, err := s.client.Upload.Upload(ctx, base64Data, uploader.UploadParams{
		PublicID:     publicID,
		Folder:       folder,
		ResourceType: "image",
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}
	if result == nil {
		return "", fmt.Errorf("empty response from cloudinary")
	}

	return result.SecureURL, nil
}

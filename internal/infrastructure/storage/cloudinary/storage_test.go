package cloudinary_test

import (
	"context"
	"strings"
	"testing"

	"go/food-recipe/internal/config"
	"go/food-recipe/internal/domain/ports"
	"go/food-recipe/internal/infrastructure/storage/cloudinary"
)

func TestCloudinaryStorage_EmptyConfig(t *testing.T) {
	storage, err := cloudinary.NewCloudinaryStorage(config.CloudinaryConfig{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var _ ports.FileStorage = storage

	_, err = storage.UploadBase64Image(context.Background(), "data:image/png;base64,abc", "test_id", "test_folder")
	if err == nil {
		t.Fatal("expected error on uninitialized client, got nil")
	}
	if !strings.Contains(err.Error(), "cloudinary client not initialized") {
		t.Errorf("expected 'cloudinary client not initialized' error, got %v", err)
	}
}

func TestCloudinaryStorage_ValidConfig(t *testing.T) {
	cfg := config.CloudinaryConfig{
		CloudName: "testcloud",
		APIKey:    "testkey",
		APISecret: "testsecret",
	}

	storage, err := cloudinary.NewCloudinaryStorage(cfg)
	if err != nil {
		t.Fatalf("unexpected error creating storage: %v", err)
	}

	var _ ports.FileStorage = storage
}

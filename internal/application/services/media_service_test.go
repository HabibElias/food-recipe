package services

import (
	"context"
	"encoding/base64"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"go/food-recipe/internal/application/dto"
	"go/food-recipe/internal/domain/entities"
	domainerrors "go/food-recipe/internal/domain/errors"
	"go/food-recipe/internal/domain/ports"
	"go/food-recipe/internal/domain/repositories"
)

type mockFileStorage struct {
	uploadFn func(ctx context.Context, base64Data, publicID, folder string) (string, error)
}

var _ ports.FileStorage = (*mockFileStorage)(nil)

func (m *mockFileStorage) UploadBase64Image(ctx context.Context, base64Data, publicID, folder string) (string, error) {
	if m.uploadFn != nil {
		return m.uploadFn(ctx, base64Data, publicID, folder)
	}
	return "https://res.cloudinary.com/demo/image/upload/sample.jpg", nil
}

type mockRecipeRepo struct {
	getCreatorFn    func(ctx context.Context, likeID uuid.UUID) (uuid.UUID, string, error)
	addRecipeImgFn  func(ctx context.Context, img *entities.RecipeImage) error
}

var _ repositories.RecipeRepository = (*mockRecipeRepo)(nil)

func (m *mockRecipeRepo) GetCreatorFromLikeID(ctx context.Context, likeID uuid.UUID) (uuid.UUID, string, error) {
	if m.getCreatorFn != nil {
		return m.getCreatorFn(ctx, likeID)
	}
	return uuid.Nil, "", nil
}

func (m *mockRecipeRepo) AddRecipeImage(ctx context.Context, img *entities.RecipeImage) error {
	if m.addRecipeImgFn != nil {
		return m.addRecipeImgFn(ctx, img)
	}
	return nil
}

func TestMediaService_UploadRecipeImages(t *testing.T) {
	ctx := context.Background()
	smallData := base64.StdEncoding.EncodeToString([]byte("fake image data"))

	t.Run("successful upload of multiple images", func(t *testing.T) {
		var uploadedFolders []string
		var savedImages []*entities.RecipeImage

		storage := &mockFileStorage{
			uploadFn: func(ctx context.Context, base64Data, publicID, folder string) (string, error) {
				uploadedFolders = append(uploadedFolders, folder)
				if !strings.HasPrefix(base64Data, "data:image/") {
					t.Errorf("expected data URI prefix, got %s", base64Data)
				}
				return "https://cloudinary.com/" + publicID + ".jpg", nil
			},
		}

		recipeRepo := &mockRecipeRepo{
			addRecipeImgFn: func(ctx context.Context, img *entities.RecipeImage) error {
				savedImages = append(savedImages, img)
				return nil
			},
		}

		svc := NewMediaService(storage, recipeRepo, &mockUserRepo{})
		input := dto.UploadRecipeImagesInputDTO{
			RecipeID: 42,
			RecipeImages: []dto.RecipeImageItemDTO{
				{
					ImageName:      "main_dish.jpg",
					ImageType:      "jpg",
					ImageBase64Str: smallData,
					IsThumbnail:    true,
				},
				{
					ImageName:      "prep.png",
					ImageType:      "image/png",
					ImageBase64Str: "data:image/png;base64," + smallData,
					IsThumbnail:    false,
				},
			},
		}

		output, err := svc.UploadRecipeImages(ctx, input)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !output.Success {
			t.Error("expected Success to be true")
		}

		if len(uploadedFolders) != 2 {
			t.Fatalf("expected 2 uploads, got %d", len(uploadedFolders))
		}
		for _, folder := range uploadedFolders {
			if folder != "recipes" {
				t.Errorf("expected folder recipes, got %s", folder)
			}
		}

		if len(savedImages) != 2 {
			t.Fatalf("expected 2 saved images, got %d", len(savedImages))
		}
		if savedImages[0].RecipeID != 42 || !savedImages[0].IsThumbnail {
			t.Errorf("unexpected saved image 0: %+v", savedImages[0])
		}
		if savedImages[1].RecipeID != 42 || savedImages[1].IsThumbnail {
			t.Errorf("unexpected saved image 1: %+v", savedImages[1])
		}
	})

	t.Run("invalid recipe_id <= 0", func(t *testing.T) {
		svc := NewMediaService(&mockFileStorage{}, &mockRecipeRepo{}, &mockUserRepo{})
		_, err := svc.UploadRecipeImages(ctx, dto.UploadRecipeImagesInputDTO{
			RecipeID: 0,
			RecipeImages: []dto.RecipeImageItemDTO{
				{ImageName: "dish", ImageType: "jpg", ImageBase64Str: smallData},
			},
		})
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("empty recipe images", func(t *testing.T) {
		svc := NewMediaService(&mockFileStorage{}, &mockRecipeRepo{}, &mockUserRepo{})
		_, err := svc.UploadRecipeImages(ctx, dto.UploadRecipeImagesInputDTO{
			RecipeID:     1,
			RecipeImages: nil,
		})
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("invalid image: empty name", func(t *testing.T) {
		svc := NewMediaService(&mockFileStorage{}, &mockRecipeRepo{}, &mockUserRepo{})
		_, err := svc.UploadRecipeImages(ctx, dto.UploadRecipeImagesInputDTO{
			RecipeID: 1,
			RecipeImages: []dto.RecipeImageItemDTO{
				{ImageName: "   ", ImageType: "jpg", ImageBase64Str: smallData},
			},
		})
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("invalid image: unsupported type", func(t *testing.T) {
		svc := NewMediaService(&mockFileStorage{}, &mockRecipeRepo{}, &mockUserRepo{})
		_, err := svc.UploadRecipeImages(ctx, dto.UploadRecipeImagesInputDTO{
			RecipeID: 1,
			RecipeImages: []dto.RecipeImageItemDTO{
				{ImageName: "doc.pdf", ImageType: "pdf", ImageBase64Str: smallData},
			},
		})
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("invalid image: bad base64 string", func(t *testing.T) {
		svc := NewMediaService(&mockFileStorage{}, &mockRecipeRepo{}, &mockUserRepo{})
		_, err := svc.UploadRecipeImages(ctx, dto.UploadRecipeImagesInputDTO{
			RecipeID: 1,
			RecipeImages: []dto.RecipeImageItemDTO{
				{ImageName: "dish.jpg", ImageType: "jpg", ImageBase64Str: "not-valid-base64!!!"},
			},
		})
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("invalid image: size exceeds 10MB", func(t *testing.T) {
		oversizedBytes := make([]byte, 10*1024*1024+10)
		oversizedBase64 := base64.StdEncoding.EncodeToString(oversizedBytes)

		svc := NewMediaService(&mockFileStorage{}, &mockRecipeRepo{}, &mockUserRepo{})
		_, err := svc.UploadRecipeImages(ctx, dto.UploadRecipeImagesInputDTO{
			RecipeID: 1,
			RecipeImages: []dto.RecipeImageItemDTO{
				{ImageName: "big.jpg", ImageType: "jpg", ImageBase64Str: oversizedBase64},
			},
		})
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput for >10MB, got %v", err)
		}
	})

	t.Run("storage upload error", func(t *testing.T) {
		storage := &mockFileStorage{
			uploadFn: func(ctx context.Context, base64Data, publicID, folder string) (string, error) {
				return "", errors.New("network upload error")
			},
		}
		svc := NewMediaService(storage, &mockRecipeRepo{}, &mockUserRepo{})
		_, err := svc.UploadRecipeImages(ctx, dto.UploadRecipeImagesInputDTO{
			RecipeID: 1,
			RecipeImages: []dto.RecipeImageItemDTO{
				{ImageName: "dish.jpg", ImageType: "jpg", ImageBase64Str: smallData},
			},
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("database save error", func(t *testing.T) {
		recipeRepo := &mockRecipeRepo{
			addRecipeImgFn: func(ctx context.Context, img *entities.RecipeImage) error {
				return errors.New("db error")
			},
		}
		svc := NewMediaService(&mockFileStorage{}, recipeRepo, &mockUserRepo{})
		_, err := svc.UploadRecipeImages(ctx, dto.UploadRecipeImagesInputDTO{
			RecipeID: 1,
			RecipeImages: []dto.RecipeImageItemDTO{
				{ImageName: "dish.jpg", ImageType: "jpg", ImageBase64Str: smallData},
			},
		})
		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestMediaService_UploadUserImage(t *testing.T) {
	ctx := context.Background()
	smallData := base64.StdEncoding.EncodeToString([]byte("user photo bytes"))
	userID := uuid.New()

	t.Run("successful upload with user_id in input", func(t *testing.T) {
		var updatedUserID uuid.UUID
		var updatedURL string

		storage := &mockFileStorage{
			uploadFn: func(ctx context.Context, base64Data, publicID, folder string) (string, error) {
				if folder != "users" {
					t.Errorf("expected folder users, got %s", folder)
				}
				return "https://cloudinary.com/avatar.jpg", nil
			},
		}

		userRepo := &mockUserRepo{
			updateAvatarFn: func(ctx context.Context, uid uuid.UUID, avatarURL string) error {
				updatedUserID = uid
				updatedURL = avatarURL
				return nil
			},
		}

		svc := NewMediaService(storage, &mockRecipeRepo{}, userRepo)
		output, err := svc.UploadUserImage(ctx, dto.UploadUserImageInputDTO{
			UserID: userID.String(),
			Image: dto.UserImageItemDTO{
				ImageName:      "profile.png",
				ImageType:      "png",
				ImageBase64Str: smallData,
			},
		}, "")

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !output.Success {
			t.Error("expected Success to be true")
		}
		if output.ImageURL != "https://cloudinary.com/avatar.jpg" {
			t.Errorf("expected avatar URL, got %s", output.ImageURL)
		}
		if updatedUserID != userID || updatedURL != "https://cloudinary.com/avatar.jpg" {
			t.Errorf("unexpected avatar update values: uid=%v, url=%s", updatedUserID, updatedURL)
		}
	})

	t.Run("successful upload with fallback to sessionUserID", func(t *testing.T) {
		var updatedUserID uuid.UUID

		userRepo := &mockUserRepo{
			updateAvatarFn: func(ctx context.Context, uid uuid.UUID, avatarURL string) error {
				updatedUserID = uid
				return nil
			},
		}

		svc := NewMediaService(&mockFileStorage{}, &mockRecipeRepo{}, userRepo)
		output, err := svc.UploadUserImage(ctx, dto.UploadUserImageInputDTO{
			UserID: "",
			Image: dto.UserImageItemDTO{
				ImageName:      "profile.jpg",
				ImageType:      "jpeg",
				ImageBase64Str: smallData,
			},
		}, userID.String())

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !output.Success {
			t.Error("expected Success to be true")
		}
		if updatedUserID != userID {
			t.Errorf("expected userID %v, got %v", userID, updatedUserID)
		}
	})

	t.Run("missing user id entirely", func(t *testing.T) {
		svc := NewMediaService(&mockFileStorage{}, &mockRecipeRepo{}, &mockUserRepo{})
		_, err := svc.UploadUserImage(ctx, dto.UploadUserImageInputDTO{
			UserID: "",
			Image: dto.UserImageItemDTO{
				ImageName:      "profile.jpg",
				ImageType:      "jpeg",
				ImageBase64Str: smallData,
			},
		}, "")
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("invalid user id uuid", func(t *testing.T) {
		svc := NewMediaService(&mockFileStorage{}, &mockRecipeRepo{}, &mockUserRepo{})
		_, err := svc.UploadUserImage(ctx, dto.UploadUserImageInputDTO{
			UserID: "not-a-uuid",
			Image: dto.UserImageItemDTO{
				ImageName:      "profile.jpg",
				ImageType:      "jpeg",
				ImageBase64Str: smallData,
			},
		}, "")
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})
}

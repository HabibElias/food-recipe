package dto

import (
	"encoding/json"
	"errors"
	"testing"

	domainerrors "go/food-recipe/internal/domain/errors"

	"github.com/google/uuid"
)

func TestPaymentActionPayloadDTO(t *testing.T) {
	t.Run("unmarshals flat input with session map and phone", func(t *testing.T) {
		rawJSON := `{
			"session_variables": {
				"x-hasura-user-id": "123e4567-e89b-12d3-a456-426614174000"
			},
			"input": {
				"amount": 99.5,
				"recipe_id": 42,
				"phone": "0911223344"
			}
		}`
		var payload InitializePaymentActionPayloadDTO
		err := json.Unmarshal([]byte(rawJSON), &payload)
		if err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}

		if payload.GetSessionUserID() != "123e4567-e89b-12d3-a456-426614174000" {
			t.Errorf("expected session user id, got %s", payload.GetSessionUserID())
		}

		inputDTO := payload.ToInputDTO()
		if inputDTO.Amount != 99.5 || inputDTO.RecipeID != 42 || inputDTO.Phone != "0911223344" {
			t.Errorf("unexpected inputDTO values: %+v", inputDTO)
		}
	})

	t.Run("unmarshals nested input with phone", func(t *testing.T) {
		rawJSON := `{
			"session_variables": {
				"x-hasura-user-id": "123e4567-e89b-12d3-a456-426614174000"
			},
			"input": {
				"input": {
					"amount": 150.0,
					"recipe_id": 10,
					"phone": "0922334455"
				}
			}
		}`
		var payload InitializePaymentActionPayloadDTO
		err := json.Unmarshal([]byte(rawJSON), &payload)
		if err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}

		inputDTO := payload.ToInputDTO()
		if inputDTO.Amount != 150.0 || inputDTO.RecipeID != 10 || inputDTO.Phone != "0922334455" {
			t.Errorf("unexpected inputDTO values: %+v", inputDTO)
		}
	})

	t.Run("handles empty session variables", func(t *testing.T) {
		var payload InitializePaymentActionPayloadDTO
		if payload.GetSessionUserID() != "" {
			t.Errorf("expected empty string, got %s", payload.GetSessionUserID())
		}
	})

	t.Run("returns error on malformed session_variables JSON", func(t *testing.T) {
		rawJSON := `{
			"session_variables": "not-a-map-or-object"
		}`
		var payload InitializePaymentActionPayloadDTO
		err := json.Unmarshal([]byte(rawJSON), &payload)
		if err == nil {
			t.Error("expected unmarshal error on malformed session_variables, got nil")
		}
	})

	t.Run("returns error on malformed input JSON", func(t *testing.T) {
		rawJSON := `{
			"input": "not-an-object"
		}`
		var payload InitializePaymentActionPayloadDTO
		err := json.Unmarshal([]byte(rawJSON), &payload)
		if err == nil {
			t.Error("expected unmarshal error on malformed input, got nil")
		}
	})
}

func TestUploadUserImageActionPayloadDTO(t *testing.T) {
	rawJSON := `{
		"session_variables": {
			"x-hasura-user-id": "user-uuid-123"
		},
		"input": {
			"input": {
				"user_id": "",
				"image": {
					"image_name": "avatar.png",
					"image_type": "png",
					"image_base64str": "dGVzdA=="
				}
			}
		}
	}`
	var payload UploadUserImageActionPayloadDTO
	err := json.Unmarshal([]byte(rawJSON), &payload)
	if err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if payload.GetSessionUserID() != "user-uuid-123" {
		t.Errorf("expected user-uuid-123, got %s", payload.GetSessionUserID())
	}
}

func TestLikeEventPayloadDTO(t *testing.T) {
	likeID := uuid.New()
	userID := uuid.New()
	rawJSON := `{
		"event": {
			"op": "INSERT",
			"data": {
				"new": {
					"id": "` + likeID.String() + `",
					"user_id": "` + userID.String() + `",
					"recipe_id": 12,
					"created_at": "2026-10-26T12:00:00Z"
				}
			}
		}
	}`
	var payload LikeEventPayloadDTO
	err := json.Unmarshal([]byte(rawJSON), &payload)
	if err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	if payload.Event.Op != "INSERT" {
		t.Errorf("expected INSERT, got %s", payload.Event.Op)
	}
	if payload.Event.Data.New.ID != likeID {
		t.Errorf("expected likeID %v, got %v", likeID, payload.Event.Data.New.ID)
	}
	if payload.Event.Data.New.UserID != userID {
		t.Errorf("expected userID %v, got %v", userID, payload.Event.Data.New.UserID)
	}
	if payload.Event.Data.New.RecipeID != 12 {
		t.Errorf("expected recipeID 12, got %d", payload.Event.Data.New.RecipeID)
	}
}

func TestLikeNotificationOutputDTO(t *testing.T) {
	t.Run("marshals success output", func(t *testing.T) {
		out := LikeNotificationOutputDTO{OK: true}
		bytes, err := json.Marshal(out)
		if err != nil {
			t.Fatalf("unexpected marshal error: %v", err)
		}
		expected := `{"ok":true}`
		if string(bytes) != expected {
			t.Errorf("expected %s, got %s", expected, string(bytes))
		}
	})

	t.Run("marshals skipped output", func(t *testing.T) {
		out := LikeNotificationOutputDTO{OK: true, Skipped: true}
		bytes, err := json.Marshal(out)
		if err != nil {
			t.Fatalf("unexpected marshal error: %v", err)
		}
		expected := `{"ok":true,"skipped":true}`
		if string(bytes) != expected {
			t.Errorf("expected %s, got %s", expected, string(bytes))
		}
	})
}

func TestLoginInputDTO_Validate(t *testing.T) {
	t.Run("valid input passes", func(t *testing.T) {
		dto := LoginInputDTO{
			Email:    "test@example.com",
			Password: "password123",
		}
		if err := dto.Validate(); err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
	})

	t.Run("invalid email fails", func(t *testing.T) {
		dto := LoginInputDTO{
			Email:    "not-an-email",
			Password: "password123",
		}
		err := dto.Validate()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got: %v", err)
		}
	})

	t.Run("empty email fails", func(t *testing.T) {
		dto := LoginInputDTO{
			Email:    "",
			Password: "password123",
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("short password fails", func(t *testing.T) {
		dto := LoginInputDTO{
			Email:    "test@example.com",
			Password: "12345",
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})

	t.Run("empty password fails", func(t *testing.T) {
		dto := LoginInputDTO{
			Email:    "test@example.com",
			Password: "",
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestSignupInputDTO_Validate(t *testing.T) {
	valid := SignupInputDTO{
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Username:  "johndoe",
		Password:  "secret123",
	}

	t.Run("valid input passes", func(t *testing.T) {
		if err := valid.Validate(); err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
	})

	t.Run("short first name fails", func(t *testing.T) {
		bad := valid
		bad.FirstName = "J"
		if err := bad.Validate(); err == nil {
			t.Fatal("expected error for short first_name, got nil")
		}
	})

	t.Run("short last name fails", func(t *testing.T) {
		bad := valid
		bad.LastName = "D"
		if err := bad.Validate(); err == nil {
			t.Fatal("expected error for short last_name, got nil")
		}
	})

	t.Run("invalid email fails", func(t *testing.T) {
		bad := valid
		bad.Email = "invalid-email"
		err := bad.Validate()
		if err == nil {
			t.Fatal("expected error for invalid email, got nil")
		}
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got: %v", err)
		}
	})

	t.Run("short username fails", func(t *testing.T) {
		bad := valid
		bad.Username = "jo"
		if err := bad.Validate(); err == nil {
			t.Fatal("expected error for short username, got nil")
		}
	})

	t.Run("short password fails", func(t *testing.T) {
		bad := valid
		bad.Password = "12345"
		if err := bad.Validate(); err == nil {
			t.Fatal("expected error for short password, got nil")
		}
	})

	t.Run("empty struct fails", func(t *testing.T) {
		var empty SignupInputDTO
		if err := empty.Validate(); err == nil {
			t.Fatal("expected error for empty struct, got nil")
		}
	})
}

func TestMeInputDTO_Validate(t *testing.T) {
	t.Run("valid token passes", func(t *testing.T) {
		dto := MeInputDTO{Token: "valid-token"}
		if err := dto.Validate(); err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
	})

	t.Run("empty token fails", func(t *testing.T) {
		dto := MeInputDTO{Token: ""}
		err := dto.Validate()
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got: %v", err)
		}
	})
}

func TestInitializePaymentInputDTO_Validate(t *testing.T) {
	t.Run("valid input without phone passes", func(t *testing.T) {
		dto := InitializePaymentInputDTO{
			Amount:   150.50,
			RecipeID: 42,
		}
		if err := dto.Validate(); err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
	})

	t.Run("valid input with phone passes", func(t *testing.T) {
		dto := InitializePaymentInputDTO{
			Amount:   150.50,
			RecipeID: 42,
			Phone:    "0911223344",
		}
		if err := dto.Validate(); err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
	})

	t.Run("zero amount fails", func(t *testing.T) {
		dto := InitializePaymentInputDTO{
			Amount:   0,
			RecipeID: 42,
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error for zero amount, got nil")
		}
	})

	t.Run("negative amount fails", func(t *testing.T) {
		dto := InitializePaymentInputDTO{
			Amount:   -10.0,
			RecipeID: 42,
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error for negative amount, got nil")
		}
	})

	t.Run("zero recipe ID fails", func(t *testing.T) {
		dto := InitializePaymentInputDTO{
			Amount:   100.0,
			RecipeID: 0,
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error for zero recipe_id, got nil")
		}
	})

	t.Run("short phone fails", func(t *testing.T) {
		dto := InitializePaymentInputDTO{
			Amount:   100.0,
			RecipeID: 42,
			Phone:    "123",
		}
		err := dto.Validate()
		if err == nil {
			t.Fatal("expected error for short phone, got nil")
		}
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got: %v", err)
		}
	})
}

func TestVerifyPaymentInputDTO_Validate(t *testing.T) {
	t.Run("valid trx_ref passes", func(t *testing.T) {
		dto := VerifyPaymentInputDTO{TrxRef: "chapa-tx-123"}
		if err := dto.Validate(); err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
	})

	t.Run("empty trx_ref fails", func(t *testing.T) {
		dto := VerifyPaymentInputDTO{TrxRef: ""}
		err := dto.Validate()
		if err == nil {
			t.Fatal("expected error for empty trx_ref, got nil")
		}
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got: %v", err)
		}
	})
}

func TestUploadRecipeImagesInputDTO_Validate(t *testing.T) {
	validImage := RecipeImageItemDTO{
		ImageName:      "food.jpg",
		ImageType:      "image/jpeg",
		ImageBase64Str: "dGVzdA==",
		IsThumbnail:    true,
	}

	t.Run("valid input passes", func(t *testing.T) {
		dto := UploadRecipeImagesInputDTO{
			RecipeID:     10,
			RecipeImages: []RecipeImageItemDTO{validImage},
		}
		if err := dto.Validate(); err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
	})

	t.Run("zero recipe ID fails", func(t *testing.T) {
		dto := UploadRecipeImagesInputDTO{
			RecipeID:     0,
			RecipeImages: []RecipeImageItemDTO{validImage},
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error for zero recipe_id, got nil")
		}
	})

	t.Run("empty recipe images fails", func(t *testing.T) {
		dto := UploadRecipeImagesInputDTO{
			RecipeID:     10,
			RecipeImages: []RecipeImageItemDTO{},
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error for empty recipe_images, got nil")
		}
	})

	t.Run("nested image with empty image_name fails", func(t *testing.T) {
		badImage := validImage
		badImage.ImageName = ""
		dto := UploadRecipeImagesInputDTO{
			RecipeID:     10,
			RecipeImages: []RecipeImageItemDTO{badImage},
		}
		err := dto.Validate()
		if err == nil {
			t.Fatal("expected error for empty image_name in slice, got nil")
		}
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got: %v", err)
		}
	})

	t.Run("nested image with empty image_type fails", func(t *testing.T) {
		badImage := validImage
		badImage.ImageType = ""
		dto := UploadRecipeImagesInputDTO{
			RecipeID:     10,
			RecipeImages: []RecipeImageItemDTO{badImage},
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error for empty image_type in slice, got nil")
		}
	})

	t.Run("nested image with empty image_base64str fails", func(t *testing.T) {
		badImage := validImage
		badImage.ImageBase64Str = ""
		dto := UploadRecipeImagesInputDTO{
			RecipeID:     10,
			RecipeImages: []RecipeImageItemDTO{badImage},
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error for empty image_base64str in slice, got nil")
		}
	})
}

func TestUploadUserImageInputDTO_Validate(t *testing.T) {
	validImage := UserImageItemDTO{
		ImageName:      "avatar.png",
		ImageType:      "image/png",
		ImageBase64Str: "dGVzdA==",
	}

	t.Run("valid input passes", func(t *testing.T) {
		dto := UploadUserImageInputDTO{
			UserID: "123e4567-e89b-12d3-a456-426614174000",
			Image:  validImage,
		}
		if err := dto.Validate(); err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
	})

	t.Run("invalid non-uuid user ID fails", func(t *testing.T) {
		dto := UploadUserImageInputDTO{
			UserID: "not-a-uuid",
			Image:  validImage,
		}
		err := dto.Validate()
		if err == nil {
			t.Fatal("expected error for non-uuid user_id, got nil")
		}
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got: %v", err)
		}
	})

	t.Run("empty user ID fails", func(t *testing.T) {
		dto := UploadUserImageInputDTO{
			UserID: "",
			Image:  validImage,
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error for empty user_id, got nil")
		}
	})

	t.Run("nested image with empty image_name fails", func(t *testing.T) {
		badImage := validImage
		badImage.ImageName = ""
		dto := UploadUserImageInputDTO{
			UserID: "123e4567-e89b-12d3-a456-426614174000",
			Image:  badImage,
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error for empty image_name, got nil")
		}
	})

	t.Run("nested image with empty image_type fails", func(t *testing.T) {
		badImage := validImage
		badImage.ImageType = ""
		dto := UploadUserImageInputDTO{
			UserID: "123e4567-e89b-12d3-a456-426614174000",
			Image:  badImage,
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error for empty image_type, got nil")
		}
	})

	t.Run("nested image with empty image_base64str fails", func(t *testing.T) {
		badImage := validImage
		badImage.ImageBase64Str = ""
		dto := UploadUserImageInputDTO{
			UserID: "123e4567-e89b-12d3-a456-426614174000",
			Image:  badImage,
		}
		if err := dto.Validate(); err == nil {
			t.Fatal("expected error for empty image_base64str, got nil")
		}
	})
}

func TestLikeEventDataNewDTO_Validate(t *testing.T) {
	valid := LikeEventDataNewDTO{
		ID:        uuid.New(),
		UserID:    uuid.New(),
		RecipeID:  15,
		CreatedAt: "2026-10-26T12:00:00Z",
	}

	t.Run("valid input passes", func(t *testing.T) {
		if err := valid.Validate(); err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
	})

	t.Run("nil id fails", func(t *testing.T) {
		bad := valid
		bad.ID = uuid.Nil
		err := bad.Validate()
		if err == nil {
			t.Fatal("expected error for nil id, got nil")
		}
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got: %v", err)
		}
	})

	t.Run("nil user_id fails", func(t *testing.T) {
		bad := valid
		bad.UserID = uuid.Nil
		if err := bad.Validate(); err == nil {
			t.Fatal("expected error for nil user_id, got nil")
		}
	})

	t.Run("zero recipe_id fails", func(t *testing.T) {
		bad := valid
		bad.RecipeID = 0
		if err := bad.Validate(); err == nil {
			t.Fatal("expected error for zero recipe_id, got nil")
		}
	})
}

func TestLikeEventPayloadDTO_Validate(t *testing.T) {
	t.Run("valid payload passes", func(t *testing.T) {
		dto := LikeEventPayloadDTO{
			Event: LikeEventDTO{
				Op: "INSERT",
				Data: LikeEventDataDTO{
					New: LikeEventDataNewDTO{
						ID:       uuid.New(),
						UserID:   uuid.New(),
						RecipeID: 10,
					},
				},
			},
		}
		if err := dto.Validate(); err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
	})

	t.Run("empty event fails", func(t *testing.T) {
		dto := LikeEventPayloadDTO{}
		err := dto.Validate()
		if err == nil {
			t.Fatal("expected error for empty event, got nil")
		}
		if !errors.Is(err, domainerrors.ErrInvalidInput) {
			t.Errorf("expected ErrInvalidInput, got: %v", err)
		}
	})
}


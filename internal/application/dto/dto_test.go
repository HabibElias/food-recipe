package dto

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"
)

func TestPaymentActionPayloadDTO(t *testing.T) {
	t.Run("unmarshals flat input with session map", func(t *testing.T) {
		rawJSON := `{
			"session_variables": {
				"x-hasura-user-id": "123e4567-e89b-12d3-a456-426614174000"
			},
			"input": {
				"amount": 99.5,
				"recipe_id": 42
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
		if inputDTO.Amount != 99.5 || inputDTO.RecipeID != 42 {
			t.Errorf("unexpected inputDTO values: %+v", inputDTO)
		}
	})

	t.Run("unmarshals nested input", func(t *testing.T) {
		rawJSON := `{
			"session_variables": {
				"x-hasura-user-id": "123e4567-e89b-12d3-a456-426614174000"
			},
			"input": {
				"input": {
					"amount": 150.0,
					"recipe_id": 10
				}
			}
		}`
		var payload InitializePaymentActionPayloadDTO
		err := json.Unmarshal([]byte(rawJSON), &payload)
		if err != nil {
			t.Fatalf("unexpected unmarshal error: %v", err)
		}

		inputDTO := payload.ToInputDTO()
		if inputDTO.Amount != 150.0 || inputDTO.RecipeID != 10 {
			t.Errorf("unexpected inputDTO values: %+v", inputDTO)
		}
	})

	t.Run("handles empty session variables", func(t *testing.T) {
		var payload InitializePaymentActionPayloadDTO
		if payload.GetSessionUserID() != "" {
			t.Errorf("expected empty string, got %s", payload.GetSessionUserID())
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

package entities_test

import (
	"testing"

	"github.com/google/uuid"
	"go/food-recipe/internal/domain/entities"
)

func TestUserEntity(t *testing.T) {
	id := uuid.New()

	t.Run("both first and last name present", func(t *testing.T) {
		u := entities.User{
			ID:           id,
			Email:        "test@example.com",
			Username:     "tester",
			FirstName:    "Test",
			LastName:     "User",
			PasswordHash: "hashed_secret",
			AvatarURL:    "https://example.com/avatar.jpg",
		}
		if u.FullName() != "Test User" {
			t.Errorf("expected 'Test User', got '%s'", u.FullName())
		}
		if u.PasswordHash != "hashed_secret" {
			t.Errorf("expected 'hashed_secret', got '%s'", u.PasswordHash)
		}
	})

	t.Run("empty first and last name returns username", func(t *testing.T) {
		u := entities.User{
			ID:       id,
			Email:    "test@example.com",
			Username: "tester",
		}
		if u.FullName() != "tester" {
			t.Errorf("expected 'tester', got '%s'", u.FullName())
		}
	})

	t.Run("only first name present", func(t *testing.T) {
		u := entities.User{
			ID:        id,
			Username:  "tester",
			FirstName: "Test",
		}
		if u.FullName() != "Test" {
			t.Errorf("expected 'Test', got '%s'", u.FullName())
		}
	})

	t.Run("only last name present", func(t *testing.T) {
		u := entities.User{
			ID:       id,
			Username: "tester",
			LastName: "User",
		}
		if u.FullName() != "User" {
			t.Errorf("expected 'User', got '%s'", u.FullName())
		}
	})
}

func TestTransactionEntity(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	tx := entities.Transaction{
		ID:          id,
		UserID:      userID,
		RecipeID:    42,
		TrxRef:      "tx-12345",
		Amount:      99.99,
		Currency:    "ETB",
		CheckoutURL: "https://checkout.chapa.co/tx-12345",
	}

	if tx.ID != id || tx.UserID != userID || tx.RecipeID != 42 || tx.TrxRef != "tx-12345" {
		t.Errorf("unexpected transaction fields: %+v", tx)
	}
	if tx.Amount != 99.99 || tx.Currency != "ETB" || tx.CheckoutURL != "https://checkout.chapa.co/tx-12345" {
		t.Errorf("unexpected transaction fields: %+v", tx)
	}
}

func TestPaymentEntity(t *testing.T) {
	userID := uuid.New()
	p := entities.Payment{
		ID:          "pay-1",
		UserID:      userID,
		RecipeID:    101,
		TxRef:       "ref-abc",
		Amount:      50.0,
		CheckoutURL: "https://checkout.chapa.co/ref-abc",
		Currency:    "USD",
		Status:      "success",
	}

	if p.ID != "pay-1" || p.UserID != userID || p.RecipeID != 101 || p.TxRef != "ref-abc" {
		t.Errorf("unexpected payment fields: %+v", p)
	}
	if p.Amount != 50.0 || p.CheckoutURL != "https://checkout.chapa.co/ref-abc" || p.Currency != "USD" || p.Status != "success" {
		t.Errorf("unexpected payment fields: %+v", p)
	}
}

func TestRecipeImageEntity(t *testing.T) {
	img := entities.RecipeImage{
		ID:          "img-123",
		RecipeID:    7,
		ImageURL:    "https://res.cloudinary.com/demo/image.png",
		IsThumbnail: true,
	}

	if img.ID != "img-123" || img.RecipeID != 7 || img.ImageURL != "https://res.cloudinary.com/demo/image.png" || !img.IsThumbnail {
		t.Errorf("unexpected recipe image fields: %+v", img)
	}
}

func TestNotificationEntity(t *testing.T) {
	id := uuid.New()
	userID := uuid.New()
	initiatorID := uuid.New()
	n := entities.Notification{
		ID:          id,
		UserID:      userID,
		Message:     "Someone liked your recipe!",
		InitiatorID: initiatorID,
	}

	if n.ID != id || n.UserID != userID || n.Message != "Someone liked your recipe!" || n.InitiatorID != initiatorID {
		t.Errorf("unexpected notification fields: %+v", n)
	}
}

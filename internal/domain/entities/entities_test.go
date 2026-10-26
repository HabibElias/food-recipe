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
			ID:        id,
			Email:     "test@example.com",
			Username:  "tester",
			FirstName: "Test",
			LastName:  "User",
		}
		if u.FullName() != "Test User" {
			t.Errorf("expected 'Test User', got '%s'", u.FullName())
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

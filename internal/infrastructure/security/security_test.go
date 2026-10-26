package security_test

import (
	"testing"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"go/food-recipe/internal/infrastructure/security"
)

func TestPasswordHashingAndVerification(t *testing.T) {
	raw := "supersecret123"
	hash, err := security.HashPassword(raw)
	if err != nil {
		t.Fatalf("unexpected error hashing: %v", err)
	}

	if err := security.ComparePassword(hash, raw); err != nil {
		t.Fatalf("expected valid password comparison, got: %v", err)
	}

	if err := security.ComparePassword(hash, "wrongpass"); err == nil {
		t.Fatalf("expected error on wrong password, got nil")
	}
}

func TestJWTGenerationAndVerification(t *testing.T) {
	secret := "secret-jwt-key-with-sufficient-length-12345"
	userID := "00000000-0000-0000-0000-000000000001"
	email := "user@example.com"

	token, err := security.GenerateHasuraJWT(userID, email, secret)
	if err != nil {
		t.Fatalf("failed to generate JWT: %v", err)
	}

	claims, err := security.VerifyHasuraJWT(token, secret)
	if err != nil {
		t.Fatalf("failed to verify JWT: %v", err)
	}

	if claims["sub"] != userID {
		t.Errorf("expected sub %s, got %v", userID, claims["sub"])
	}
	if claims["email"] != email {
		t.Errorf("expected email %s, got %v", email, claims["email"])
	}

	hasuraClaims, ok := claims["https://hasura.io/jwt/claims"].(map[string]interface{})
	if !ok {
		t.Fatalf("expected hasura claims map, got %T", claims["https://hasura.io/jwt/claims"])
	}
	if hasuraClaims["x-hasura-user-id"] != userID {
		t.Errorf("expected x-hasura-user-id %s, got %v", userID, hasuraClaims["x-hasura-user-id"])
	}
	if hasuraClaims["x-hasura-default-role"] != "user" {
		t.Errorf("expected default role 'user', got %v", hasuraClaims["x-hasura-default-role"])
	}

	// Verify failure on wrong secret
	if _, err := security.VerifyHasuraJWT(token, "wrong-secret"); err == nil {
		t.Fatalf("expected error on wrong secret, got nil")
	}

	// Verify failure on empty secret for generation
	if _, err := security.GenerateHasuraJWT(userID, email, ""); err == nil {
		t.Fatalf("expected error generating JWT with empty secret, got nil")
	}

	// Verify failure on empty secret for verification
	if _, err := security.VerifyHasuraJWT(token, ""); err == nil {
		t.Fatalf("expected error verifying JWT with empty secret, got nil")
	}

	// Verify failure on malformed token string
	if _, err := security.VerifyHasuraJWT("not.a.valid.jwt", secret); err == nil {
		t.Fatalf("expected error on malformed token string, got nil")
	}
}

func TestJWTExpiredToken(t *testing.T) {
	secret := "secret-jwt-key-with-sufficient-length-12345"
	userID := "00000000-0000-0000-0000-000000000001"
	email := "user@example.com"

	// Create an expired token manually
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"exp":   time.Now().Add(-1 * time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to sign token: %v", err)
	}

	if _, err := security.VerifyHasuraJWT(tokenStr, secret); err == nil {
		t.Fatalf("expected error on expired token, got nil")
	}
}

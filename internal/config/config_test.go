package config_test

import (
	"testing"

	"go/food-recipe/internal/config"
)

func TestLoadConfig(t *testing.T) {
	t.Setenv("PORT", "9000")
	t.Setenv("ALLOWED_ORIGINS", "http://localhost:3000")
	t.Setenv("CALLBACK_URL", "http://localhost:8000/callback")
	t.Setenv("HASURA_URL", "http://localhost:8080/v1/graphql")
	t.Setenv("HASURA_GRAPHQL_ADMIN_SECRET", "secret")
	t.Setenv("HASURA_ACTION_SECRET", "webhook-secret")
	t.Setenv("HASURA_JWT_SECRET", "jwtsecret123456789012345678901234")
	t.Setenv("CLOUDINARY_CLOUD_NAME", "cloud")
	t.Setenv("CLOUDINARY_API_KEY", "key")
	t.Setenv("CLOUDINARY_API_SECRET", "secret")
	t.Setenv("CHAPA_SECRET_KEY", "chapa_key")

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.App.Port != "9000" {
		t.Errorf("expected port 9000, got %s", cfg.App.Port)
	}
	if cfg.App.AllowedOrigins != "http://localhost:3000" {
		t.Errorf("expected allowed origins http://localhost:3000, got %s", cfg.App.AllowedOrigins)
	}
	if cfg.App.CallbackURL != "http://localhost:8000/callback" {
		t.Errorf("expected callback url http://localhost:8000/callback, got %s", cfg.App.CallbackURL)
	}
	if cfg.Hasura.URL != "http://localhost:8080/v1/graphql" {
		t.Errorf("expected hasura url, got %s", cfg.Hasura.URL)
	}
	if cfg.Hasura.AdminSecret != "secret" {
		t.Errorf("expected admin secret, got %s", cfg.Hasura.AdminSecret)
	}
	if cfg.Hasura.ActionSecret != "webhook-secret" {
		t.Error("expected configured Hasura webhook secret")
	}
	if cfg.JWT.Secret != "jwtsecret123456789012345678901234" {
		t.Errorf("expected jwt secret, got %s", cfg.JWT.Secret)
	}
	if cfg.Cloudinary.CloudName != "cloud" {
		t.Errorf("expected cloud name, got %s", cfg.Cloudinary.CloudName)
	}
	if cfg.Cloudinary.APIKey != "key" {
		t.Errorf("expected cloudinary api key, got %s", cfg.Cloudinary.APIKey)
	}
	if cfg.Cloudinary.APISecret != "secret" {
		t.Errorf("expected cloudinary api secret, got %s", cfg.Cloudinary.APISecret)
	}
	if cfg.Chapa.SecretKey != "chapa_key" {
		t.Errorf("expected chapa secret key, got %s", cfg.Chapa.SecretKey)
	}
	if config.CFG != cfg {
		t.Errorf("expected CFG singleton to match returned cfg")
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("ALLOWED_ORIGINS", "")

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.App.Port != "8000" {
		t.Errorf("expected default port 8000, got %s", cfg.App.Port)
	}
	if cfg.App.AllowedOrigins != "*" {
		t.Errorf("expected default allowed origins *, got %s", cfg.App.AllowedOrigins)
	}
}

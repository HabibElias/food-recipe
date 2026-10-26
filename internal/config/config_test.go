package config_test

import (
	"os"
	"testing"

	"go/food-recipe/internal/config"
)

func TestLoadConfig(t *testing.T) {
	os.Setenv("PORT", "9000")
	os.Setenv("HASURA_URL", "http://localhost:8080/v1/graphql")
	os.Setenv("HASURA_GRAPHQL_ADMIN_SECRET", "secret")
	os.Setenv("HASURA_JWT_SECRET", "jwtsecret123456789012345678901234")
	os.Setenv("CLOUDINARY_CLOUD_NAME", "cloud")
	os.Setenv("CLOUDINARY_API_KEY", "key")
	os.Setenv("CLOUDINARY_API_SECRET", "secret")
	os.Setenv("CHAPA_SECRET_KEY", "chapa_key")

	cfg, err := config.LoadConfig()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if cfg.App.Port != "9000" {
		t.Errorf("expected port 9000, got %s", cfg.App.Port)
	}
	if cfg.Hasura.URL != "http://localhost:8080/v1/graphql" {
		t.Errorf("expected hasura url, got %s", cfg.Hasura.URL)
	}
	if cfg.Hasura.AdminSecret != "secret" {
		t.Errorf("expected admin secret, got %s", cfg.Hasura.AdminSecret)
	}
	if cfg.JWT.Secret != "jwtsecret123456789012345678901234" {
		t.Errorf("expected jwt secret, got %s", cfg.JWT.Secret)
	}
	if cfg.Cloudinary.CloudName != "cloud" {
		t.Errorf("expected cloud name, got %s", cfg.Cloudinary.CloudName)
	}
	if cfg.Chapa.SecretKey != "chapa_key" {
		t.Errorf("expected chapa secret key, got %s", cfg.Chapa.SecretKey)
	}
	if config.CFG != cfg {
		t.Errorf("expected CFG singleton to match returned cfg")
	}
}

func TestLoadConfig_Defaults(t *testing.T) {
	os.Unsetenv("PORT")
	os.Unsetenv("ALLOWED_ORIGINS")

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

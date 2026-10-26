package config

import (
	"os"

	"github.com/joho/godotenv"
)

type AppConfig struct {
	Port           string
	AllowedOrigins string
	CallbackURL    string
}

type HasuraConfig struct {
	URL         string
	AdminSecret string
}

type JWTConfig struct {
	Secret string
}

type CloudinaryConfig struct {
	CloudName string
	APIKey    string
	APISecret string
}

type ChapaConfig struct {
	SecretKey string
}

type Config struct {
	App        AppConfig
	Hasura     HasuraConfig
	JWT        JWTConfig
	Cloudinary CloudinaryConfig
	Chapa      ChapaConfig
}

var CFG *Config

func LoadConfig() (*Config, error) {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	origins := os.Getenv("ALLOWED_ORIGINS")
	if origins == "" {
		origins = "*"
	}

	cfg := &Config{
		App: AppConfig{
			Port:           port,
			AllowedOrigins: origins,
			CallbackURL:    os.Getenv("CALLBACK_URL"),
		},
		Hasura: HasuraConfig{
			URL:         os.Getenv("HASURA_URL"),
			AdminSecret: os.Getenv("HASURA_GRAPHQL_ADMIN_SECRET"),
		},
		JWT: JWTConfig{
			Secret: os.Getenv("HASURA_JWT_SECRET"),
		},
		Cloudinary: CloudinaryConfig{
			CloudName: os.Getenv("CLOUDINARY_CLOUD_NAME"),
			APIKey:    os.Getenv("CLOUDINARY_API_KEY"),
			APISecret: os.Getenv("CLOUDINARY_API_SECRET"),
		},
		Chapa: ChapaConfig{
			SecretKey: os.Getenv("CHAPA_SECRET_KEY"),
		},
	}

	CFG = cfg
	return cfg, nil
}

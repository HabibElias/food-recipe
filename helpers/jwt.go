package helpers

import (
	"log"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func GenerateHasuraJWT(userID, email string) (string, error) {
	var jwtSecret = []byte(os.Getenv("HASURA_JWT_SECRET"))

	if jwtSecret == nil {
		log.Fatal("HASURA_JWT_SECRET is not set on the env file")
	}

	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"https://hasura.io/jwt/claims": map[string]interface{}{
			"x-hasura-allowed-roles": []string{"user", "anon", "admin"},
			"x-hasura-default-role":  "user",
			"x-hasura-user-id":       userID,
		},
		"exp": time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

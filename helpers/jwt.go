package helpers

import (
	"fmt"
	"os"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func GenerateHasuraJWT(userID, email string) (string, error) {
	jwtSecretStr := os.Getenv("HASURA_JWT_SECRET")
	if jwtSecretStr == "" {
		return "", fmt.Errorf("HASURA_JWT_SECRET is not set in environment")
	}
	jwtSecret := []byte(jwtSecretStr)

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


func VerifyHasuraJWT(tokenStr string) (jwt.MapClaims, error) {
	jwtSecretStr := os.Getenv("HASURA_JWT_SECRET")
	if jwtSecretStr == "" {
		return nil, fmt.Errorf("HASURA_JWT_SECRET is not set in environment")
	}

	token, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		// ensure token uses HMAC signing
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(jwtSecretStr), nil
	})
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// additional expiry guard (ParseWithClaims typically validates this, but keep as a safeguard)
	if expRaw, ok := claims["exp"]; ok {
		switch exp := expRaw.(type) {
		case float64:
			if int64(exp) < time.Now().Unix() {
				return nil, fmt.Errorf("token expired")
			}
		case int64:
			if exp < time.Now().Unix() {
				return nil, fmt.Errorf("token expired")
			}
		}
	}

	return claims, nil
}
package security

import (
	"fmt"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

func GenerateHasuraJWT(userID, email, secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("jwt secret is empty")
	}
	if userID == "" {
		return "", fmt.Errorf("user id cannot be empty")
	}

	now := time.Now()
	claims := jwt.MapClaims{
		"sub":   userID,
		"email": email,
		"iat":   now.Unix(),
		"exp":   now.Add(24 * time.Hour).Unix(),
		"https://hasura.io/jwt/claims": map[string]interface{}{
			"x-hasura-allowed-roles": []string{"user", "anon"},
			"x-hasura-default-role":  "user",
			"x-hasura-user-id":       userID,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func VerifyHasuraJWT(tokenStr, secret string) (jwt.MapClaims, error) {
	if secret == "" {
		return nil, fmt.Errorf("jwt secret is empty")
	}

	token, err := jwt.ParseWithClaims(tokenStr, jwt.MapClaims{}, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(secret), nil
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

	if expRaw, ok := claims["exp"]; ok {
		var expUnix int64
		switch exp := expRaw.(type) {
		case float64:
			expUnix = int64(exp)
		case int64:
			expUnix = exp
		}
		if expUnix > 0 && expUnix < time.Now().Unix() {
			return nil, fmt.Errorf("token expired")
		}
	}

	return claims, nil
}

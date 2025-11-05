package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"go/food-recipe/helpers"

	jwt "github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	ctxUserIDKey contextKey = "sub"
	ctxEmailKey  contextKey = "email"
)

// RequireAuth verifies a Bearer JWT and injects user id and email into the request context.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			helpers.RespondWithError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			helpers.RespondWithError(w, http.StatusUnauthorized, "invalid authorization header")
			return
		}
		tokenStr := parts[1]
		secret := []byte(os.Getenv("HASURA_JWT_SECRET"))
		if len(secret) == 0 {
			helpers.RespondWithError(w, http.StatusInternalServerError, "server misconfiguration")
			return
		}

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method")
			}

			return secret, nil
		})
		if err != nil || !token.Valid {
			helpers.RespondWithError(w, http.StatusUnauthorized, "invalid token here is the problem")
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			helpers.RespondWithError(w, http.StatusUnauthorized, "invalid token claims")
			return
		}

		sub, _ := claims["sub"].(string)
		email, _ := claims["email"].(string)

		ctx := context.WithValue(r.Context(), ctxUserIDKey, sub)
		ctx = context.WithValue(ctx, ctxEmailKey, email)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetUserFromContext returns user id and email stored by RequireAuth. ok=false if missing.
func GetUserFromContext(ctx context.Context) (userID string, email string, ok bool) {
	uid, ok1 := ctx.Value(ctxUserIDKey).(string)
	em, ok2 := ctx.Value(ctxEmailKey).(string)
	return uid, em, ok1 && ok2
}

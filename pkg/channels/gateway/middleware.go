package gateway

import (
	"context"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/pomclaw/pomclaw/pkg/channels/gateway/handlers"
)

// jwtClaims mirrors the claims produced by handlers.auth.generateToken.
type jwtClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// jwtMiddleware validates the Bearer token and injects the user ID into the request context.
func jwtMiddleware(secret string, next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"code":"unauthorized","message":"missing or invalid Authorization header"}`, http.StatusUnauthorized)
			return
		}
		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &jwtClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, `{"code":"unauthorized","message":"invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), handlers.UserIDKey, claims.UserID)
		next(w, r.WithContext(ctx))
	})
}

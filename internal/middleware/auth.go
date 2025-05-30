package middleware

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/golang-jwt/jwt/v4"
)

// key type for context storage
type ctxKey string

const (
	// ContextKeyClaims is the context key where we store JWT claims
	ContextKeyClaims ctxKey = "jwtClaims"

	// Env var for your HMAC secret
	envJWTSecret = "JWT_SECRET"
)

// RequireJWT wraps an http.Handler, enforcing a valid Bearer token.
func RequireJWT(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			http.Error(w, "missing or invalid Authorization header", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(auth, "Bearer ")
		secret := os.Getenv(envJWTSecret)
		if secret == "" {
			http.Error(w, "server misconfiguration", http.StatusInternalServerError)
			return
		}

		// Parse and validate the token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Ensure token method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(secret), nil
		})
		if err != nil || !token.Valid {
			http.Error(w, "invalid token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// Optionally, you can inspect claims here (e.g. roles)
		// e.g. claims["role"] == "admin"
		// and reject if not authorized.

		// Store claims in context for downstream handlers
		ctx := context.WithValue(r.Context(), ContextKeyClaims, token.Claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// FromContext retrieves the JWT claims from the request context.
func FromContext(ctx context.Context) (jwt.MapClaims, bool) {
	raw, ok := ctx.Value(ContextKeyClaims).(jwt.Claims)
	if !ok {
		return nil, false
	}
	mapClaims, ok := raw.(jwt.MapClaims)
	return mapClaims, ok
}

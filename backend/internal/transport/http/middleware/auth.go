package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// Auth returns a middleware that accepts either:
//   - a valid API key in the "Authorization: Bearer <key>" header
//   - a valid JWT in the "session" httpOnly cookie
//
// Returns 401 JSON if neither is present or valid.
func Auth(jwtSecret []byte, apiKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 1. Check Authorization header for API key
			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.SplitN(authHeader, " ", 2)
				if len(parts) == 2 && parts[0] == "Bearer" && parts[1] == apiKey {
					next.ServeHTTP(w, r)
					return
				}
			}

			// 2. Check JWT session cookie
			cookie, err := r.Cookie("session")
			if err == nil && cookie.Value != "" {
				token, err := jwt.ParseWithClaims(cookie.Value, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
					if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
						return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
					}
					return jwtSecret, nil
				})
				if err == nil && token.Valid {
					next.ServeHTTP(w, r)
					return
				}
			}

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"message":"unauthorized"}`))
		})
	}
}

package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AuthHandler struct {
	jwtSecret     []byte
	adminPassword string
}

func NewAuthHandler(jwtSecret []byte, adminPassword string) *AuthHandler {
	return &AuthHandler{jwtSecret: jwtSecret, adminPassword: adminPassword}
}

// POST /auth/login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondError(w, http.StatusBadRequest, "invalid request body", nil)
		return
	}

	if req.Username != "admin" || req.Password != h.adminPassword {
		respondError(w, http.StatusUnauthorized, "invalid credentials", nil)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "admin",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})

	signed, err := token.SignedString(h.jwtSecret)
	if err != nil {
		respondError(w, http.StatusInternalServerError, "internal error", nil)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    signed,
		HttpOnly: true,
		Path:     "/",
		MaxAge:   3600,
		SameSite: http.SameSiteLaxMode,
	})

	w.WriteHeader(http.StatusOK)
}

// POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session",
		Value:    "",
		HttpOnly: true,
		Path:     "/",
		MaxAge:   -1,
		SameSite: http.SameSiteLaxMode,
	})
	w.WriteHeader(http.StatusOK)
}

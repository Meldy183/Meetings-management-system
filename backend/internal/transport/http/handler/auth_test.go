package handler

import (
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func newTestAuthHandler(t *testing.T) (*AuthHandler, []byte) {
	t.Helper()
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	return NewAuthHandler(secret, "testpassword"), secret
}

// POST /auth/login — correct credentials

func TestAuthLogin_OK(t *testing.T) {
	h, _ := newTestAuthHandler(t)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/login", h.Login)
	w := doRequest(mux, "POST", "/auth/login", map[string]string{
		"username": "admin", "password": "testpassword",
	})

	if w.Code != http.StatusOK {
		t.Fatalf("want 200, got %d: %s", w.Code, w.Body.String())
	}

	var sessionCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "session" {
			sessionCookie = c
			break
		}
	}
	if sessionCookie == nil {
		t.Fatal("expected session cookie to be set")
	}
	if !sessionCookie.HttpOnly {
		t.Error("session cookie must be HttpOnly")
	}
	if sessionCookie.MaxAge != 3600 {
		t.Errorf("want MaxAge=3600, got %d", sessionCookie.MaxAge)
	}
}

// POST /auth/login — wrong password

func TestAuthLogin_WrongPassword(t *testing.T) {
	h, _ := newTestAuthHandler(t)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/login", h.Login)
	w := doRequest(mux, "POST", "/auth/login", map[string]string{
		"username": "admin", "password": "wrongpassword",
	})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", w.Code)
	}
}

// POST /auth/login — wrong username

func TestAuthLogin_WrongUsername(t *testing.T) {
	h, _ := newTestAuthHandler(t)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/login", h.Login)
	w := doRequest(mux, "POST", "/auth/login", map[string]string{
		"username": "notadmin", "password": "testpassword",
	})

	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", w.Code)
	}
}

// POST /auth/login — invalid JSON body

func TestAuthLogin_InvalidBody(t *testing.T) {
	h, _ := newTestAuthHandler(t)

	r := httptest.NewRequest("POST", "/auth/login", strings.NewReader("not json"))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	h.Login(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("want 400, got %d", w.Code)
	}
}

// POST /auth/login — JWT is signed with the handler's secret and can be validated

func TestAuthLogin_TokenIsValid(t *testing.T) {
	h, secret := newTestAuthHandler(t)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/login", h.Login)
	w := doRequest(mux, "POST", "/auth/login", map[string]string{
		"username": "admin", "password": "testpassword",
	})

	var sessionCookie *http.Cookie
	for _, c := range w.Result().Cookies() {
		if c.Name == "session" {
			sessionCookie = c
		}
	}
	if sessionCookie == nil {
		t.Fatal("no session cookie")
	}

	token, err := jwt.ParseWithClaims(sessionCookie.Value, &jwt.RegisteredClaims{}, func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})
	if err != nil || !token.Valid {
		t.Errorf("token should be valid: %v", err)
	}

	claims, ok := token.Claims.(*jwt.RegisteredClaims)
	if !ok {
		t.Fatal("unexpected claims type")
	}
	if claims.Subject != "admin" {
		t.Errorf("want subject=admin, got %s", claims.Subject)
	}
	if time.Until(claims.ExpiresAt.Time) > time.Hour || time.Until(claims.ExpiresAt.Time) < 59*time.Minute {
		t.Errorf("expiry should be ~1h from now, got %v", claims.ExpiresAt.Time)
	}
}

// POST /auth/logout — clears the session cookie

func TestAuthLogout_ClearsCookie(t *testing.T) {
	h, _ := newTestAuthHandler(t)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/logout", h.Logout)
	w := doRequest(mux, "POST", "/auth/logout", nil)

	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}

	for _, c := range w.Result().Cookies() {
		if c.Name == "session" && c.MaxAge < 0 {
			return // found cleared cookie — test passes
		}
	}
	t.Error("expected session cookie to be cleared (MaxAge < 0)")
}

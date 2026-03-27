package middleware

import (
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func makeTestSecret(t *testing.T) []byte {
	t.Helper()
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		t.Fatal(err)
	}
	return secret
}

func makeJWT(t *testing.T, secret []byte, expiry time.Duration) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   "admin",
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})
	signed, err := token.SignedString(secret)
	if err != nil {
		t.Fatal(err)
	}
	return signed
}

// Valid API key in Authorization header → allowed

func TestAuth_APIKey_Valid(t *testing.T) {
	secret := makeTestSecret(t)
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/people", nil)
	r.Header.Set("Authorization", "Bearer myapikey")
	w := httptest.NewRecorder()
	Auth(secret, "myapikey")(next).ServeHTTP(w, r)

	if !called {
		t.Error("expected next to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
}

// Wrong API key → 401

func TestAuth_APIKey_Wrong(t *testing.T) {
	secret := makeTestSecret(t)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/people", nil)
	r.Header.Set("Authorization", "Bearer wrongkey")
	w := httptest.NewRecorder()
	Auth(secret, "correctkey")(next).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", w.Code)
	}
}

// Valid JWT cookie → allowed

func TestAuth_JWT_ValidCookie(t *testing.T) {
	secret := makeTestSecret(t)
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	tokenStr := makeJWT(t, secret, time.Hour)
	r := httptest.NewRequest("GET", "/people", nil)
	r.AddCookie(&http.Cookie{Name: "session", Value: tokenStr})
	w := httptest.NewRecorder()
	Auth(secret, "myapikey")(next).ServeHTTP(w, r)

	if !called {
		t.Error("expected next to be called")
	}
	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
}

// Expired JWT cookie → 401

func TestAuth_JWT_ExpiredCookie(t *testing.T) {
	secret := makeTestSecret(t)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tokenStr := makeJWT(t, secret, -time.Hour)
	r := httptest.NewRequest("GET", "/people", nil)
	r.AddCookie(&http.Cookie{Name: "session", Value: tokenStr})
	w := httptest.NewRecorder()
	Auth(secret, "myapikey")(next).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", w.Code)
	}
}

// JWT signed with different secret → 401

func TestAuth_JWT_WrongSecret(t *testing.T) {
	secret := makeTestSecret(t)
	wrongSecret := makeTestSecret(t)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	tokenStr := makeJWT(t, wrongSecret, time.Hour)
	r := httptest.NewRequest("GET", "/people", nil)
	r.AddCookie(&http.Cookie{Name: "session", Value: tokenStr})
	w := httptest.NewRecorder()
	Auth(secret, "myapikey")(next).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", w.Code)
	}
}

// No credentials at all → 401

func TestAuth_NoCredentials(t *testing.T) {
	secret := makeTestSecret(t)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/people", nil)
	w := httptest.NewRecorder()
	Auth(secret, "myapikey")(next).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", w.Code)
	}
}

// Malformed Authorization header (no Bearer prefix) → 401

func TestAuth_APIKey_MalformedHeader(t *testing.T) {
	secret := makeTestSecret(t)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/people", nil)
	r.Header.Set("Authorization", "myapikey") // missing "Bearer "
	w := httptest.NewRecorder()
	Auth(secret, "myapikey")(next).ServeHTTP(w, r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("want 401, got %d", w.Code)
	}
}

// 401 response body is valid JSON

func TestAuth_Unauthorized_ReturnsJSON(t *testing.T) {
	secret := makeTestSecret(t)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/people", nil)
	w := httptest.NewRecorder()
	Auth(secret, "myapikey")(next).ServeHTTP(w, r)

	if ct := w.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("want Content-Type application/json, got %s", ct)
	}
	body := w.Body.String()
	if body == "" {
		t.Error("expected non-empty body")
	}
}

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"meetings-editor/internal/testutil"
	"meetings-editor/pkg/logger"
)

// --- CORS ---

func TestCORS_SetsHeaders(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	CORS(next).ServeHTTP(w, r)

	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Errorf("missing Access-Control-Allow-Origin header")
	}
	if w.Header().Get("Access-Control-Allow-Methods") == "" {
		t.Errorf("missing Access-Control-Allow-Methods header")
	}
	if w.Header().Get("Access-Control-Allow-Headers") == "" {
		t.Errorf("missing Access-Control-Allow-Headers header")
	}
}

func TestCORS_Options_Returns204AndDoesNotCallNext(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	r := httptest.NewRequest("OPTIONS", "/", nil)
	w := httptest.NewRecorder()
	CORS(next).ServeHTTP(w, r)

	if w.Code != http.StatusNoContent {
		t.Errorf("want 204, got %d", w.Code)
	}
	if called {
		t.Error("next handler should not be called for OPTIONS")
	}
}

func TestCORS_NonOptions_CallsNext(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("POST", "/", nil)
	w := httptest.NewRecorder()
	CORS(next).ServeHTTP(w, r)

	if !called {
		t.Error("next handler should be called for non-OPTIONS")
	}
	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
}

// --- Logging ---

func TestLogging_CapturesStatusCode(t *testing.T) {
	log := logger.New("dev")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	r := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	Logging(log)(next).ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Errorf("want 201, got %d", w.Code)
	}
}

func TestLogging_InjectsLoggerIntoContext(t *testing.T) {
	log := logger.New("dev")

	var loggerFromCtx logger.Logger
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		loggerFromCtx = logger.FromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	r := httptest.NewRequest("GET", "/", nil)
	r = r.WithContext(testutil.Ctx())
	w := httptest.NewRecorder()
	Logging(log)(next).ServeHTTP(w, r)

	if loggerFromCtx == nil {
		t.Error("expected logger to be injected into context")
	}
}

func TestLogging_DefaultStatusIs200(t *testing.T) {
	log := logger.New("dev")

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// no explicit WriteHeader — should default to 200
		_, _ = w.Write([]byte("ok"))
	})

	r := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	Logging(log)(next).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("want 200, got %d", w.Code)
	}
}

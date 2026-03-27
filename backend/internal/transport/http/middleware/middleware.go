package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"meetings-editor/pkg/logger"
)

// Logging logs method, path, status code, and duration for every request.
func Logging(log logger.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			ctx := logger.WithLogger(r.Context(), log)
			next.ServeHTTP(rw, r.WithContext(ctx))

			log.Info(ctx, "http request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", rw.status),
				zap.Duration("duration", time.Since(start)),
			)
		})
	}
}

// CORS adds permissive CORS headers (suitable for same-network / docker-compose deployments).
// When a real Origin is present, echoes it back and allows credentials (cookies).
// When no Origin header is present (non-browser clients), uses * without credentials.
func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			// Browser request: echo origin so credentials (cookies) can be included.
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Access-Control-Allow-Credentials", "true")
		} else {
			// Non-browser client (curl, console): wildcard is fine, no credentials needed.
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.status = code
	rw.ResponseWriter.WriteHeader(code)
}

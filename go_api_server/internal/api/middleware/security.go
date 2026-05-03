// Package middleware provides HTTP middleware for the CT Explorer Go API server.
package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
)

// Security returns a Chi middleware that adds security headers and logs requests.
//
// Security headers on every response:
//   - X-Content-Type-Options: nosniff
//   - X-Frame-Options: DENY
//   - Content-Type charset enforcement
//
// Additional headers on auth endpoints (path starts with /api/auth/):
//   - Cache-Control: no-store
//
// Request logging:
//   - INFO level for 1xx–3xx responses
//   - WARNING level for 4xx responses
//   - ERROR level for 5xx responses
func Security(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Set security headers before the response is written.
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-Frame-Options", "DENY")
		if strings.HasPrefix(r.URL.Path, "/api/auth/") {
			w.Header().Set("Cache-Control", "no-store")
		}

		ww := chimiddleware.NewWrapResponseWriter(w, r.ProtoMajor)
		next.ServeHTTP(ww, r)

		// Charset enforcement after handler has set Content-Type.
		ct := ww.Header().Get("Content-Type")
		if ct != "" && !strings.Contains(ct, "charset") {
			ww.Header().Set("Content-Type", ct+"; charset=utf-8")
		}

		// Request logging.
		statusCode := ww.Status()
		durationMs := time.Since(start).Milliseconds()

		logAttrs := []any{
			"log_type", "request",
			"endpoint", r.URL.Path,
			"method", r.Method,
			"status_code", statusCode,
			"client_ip", r.RemoteAddr,
			"duration_ms", durationMs,
		}

		// Log OIDC identity headers from ALB authentication (if present).
		// These headers are set by ALB after successful OIDC authentication.
		// Never log the full JWT or access token — only the identity (email/username).
		if oidcIdentity := r.Header.Get("X-Amzn-Oidc-Identity"); oidcIdentity != "" {
			logAttrs = append(logAttrs, "oidc_identity", oidcIdentity)
		}
		if oidcIssuer := r.Header.Get("X-Amzn-Oidc-Issuer"); oidcIssuer != "" {
			logAttrs = append(logAttrs, "oidc_issuer", oidcIssuer)
		}
		// Log presence of data/access token headers (not the values — security)
		if r.Header.Get("X-Amzn-Oidc-Data") != "" {
			logAttrs = append(logAttrs, "oidc_data_present", true)
		}
		if r.Header.Get("X-Amzn-Oidc-Accesstoken") != "" {
			logAttrs = append(logAttrs, "oidc_accesstoken_present", true)
		}

		switch {
		case statusCode >= 500:
			slog.Error("request_completed", logAttrs...)
		case statusCode >= 400:
			slog.Warn("request_completed", logAttrs...)
		default:
			slog.Info("request_completed", logAttrs...)
		}
	})
}

// CORSOptions returns cors.Options configured from the CORS_ALLOWED_ORIGINS
// environment variable (default "http://localhost:5173"). Allowed methods are
// GET and POST; the only allowed header is Content-Type.
func CORSOptions() cors.Options {
	origins := os.Getenv("CORS_ALLOWED_ORIGINS")
	if origins == "" {
		origins = "http://localhost:5173"
	}
	return cors.Options{
		AllowedOrigins: strings.Split(origins, ","),
		AllowedMethods: []string{"GET", "POST"},
		AllowedHeaders: []string{"Content-Type", "Authorization"},
	}
}

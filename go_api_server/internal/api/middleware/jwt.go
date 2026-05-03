package middleware

import (
	"crypto"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"go_api_server/internal/api"
)

// jwksCache caches the Cognito JWKS keys.
var (
	jwksOnce sync.Once
	jwksKeys map[string]*rsa.PublicKey
)

type jwksResponse struct {
	Keys []jwksKey `json:"keys"`
}

type jwksKey struct {
	Kid string `json:"kid"`
	Kty string `json:"kty"`
	N   string `json:"n"`
	E   string `json:"e"`
	Use string `json:"use"`
}

// loadJWKS fetches and caches the Cognito JWKS public keys.
func loadJWKS() {
	jwksKeys = make(map[string]*rsa.PublicKey)

	resp, err := http.Get(api.CognitoJWKSURL)
	if err != nil {
		slog.Error("jwks_fetch_failed", "error", err, "url", api.CognitoJWKSURL)
		return
	}
	defer resp.Body.Close()

	var jwks jwksResponse
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		slog.Error("jwks_decode_failed", "error", err)
		return
	}

	for _, key := range jwks.Keys {
		if key.Kty != "RSA" || key.Use != "sig" {
			continue
		}
		nBytes, err := base64.RawURLEncoding.DecodeString(key.N)
		if err != nil {
			continue
		}
		eBytes, err := base64.RawURLEncoding.DecodeString(key.E)
		if err != nil {
			continue
		}
		n := new(big.Int).SetBytes(nBytes)
		e := 0
		for _, b := range eBytes {
			e = e<<8 + int(b)
		}
		jwksKeys[key.Kid] = &rsa.PublicKey{N: n, E: e}
	}

	slog.Info("jwks_loaded", "key_count", len(jwksKeys))
}

// jwtHeader is the decoded JWT header.
type jwtHeader struct {
	Kid string `json:"kid"`
	Alg string `json:"alg"`
}

// jwtPayload is the decoded JWT payload (claims we care about).
type jwtPayload struct {
	Sub      string `json:"sub"`
	Email    string `json:"email"`
	Iss      string `json:"iss"`
	Exp      int64  `json:"exp"`
	TokenUse string `json:"token_use"`
}

// validateJWT validates a Cognito JWT token. Returns the payload if valid.
func validateJWT(tokenStr string) (*jwtPayload, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}

	// Decode header
	headerBytes, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid header encoding")
	}
	var header jwtHeader
	if err := json.Unmarshal(headerBytes, &header); err != nil {
		return nil, fmt.Errorf("invalid header JSON")
	}

	// Decode payload
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid payload encoding")
	}
	var payload jwtPayload
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return nil, fmt.Errorf("invalid payload JSON")
	}

	// Check expiry
	if payload.Exp < time.Now().Unix() {
		return nil, fmt.Errorf("token expired")
	}

	// Check issuer
	expectedIssuer := fmt.Sprintf("https://cognito-idp.%s.amazonaws.com/%s", api.CognitoRegion, api.CognitoUserPoolID)
	if payload.Iss != expectedIssuer {
		return nil, fmt.Errorf("invalid issuer: %s", payload.Iss)
	}

	// Verify signature using JWKS
	jwksOnce.Do(loadJWKS)
	if len(jwksKeys) > 0 {
		pubKey, ok := jwksKeys[header.Kid]
		if !ok {
			return nil, fmt.Errorf("unknown signing key: %s", header.Kid)
		}
		sigBytes, err := base64.RawURLEncoding.DecodeString(parts[2])
		if err != nil {
			return nil, fmt.Errorf("invalid signature encoding")
		}
		signingInput := parts[0] + "." + parts[1]
		digest := sha256.Sum256([]byte(signingInput))
		if err := rsa.VerifyPKCS1v15(pubKey, crypto.SHA256, digest[:], sigBytes); err != nil {
			return nil, fmt.Errorf("invalid token signature")
		}
	}

	return &payload, nil
}

// JWTAuth returns a Chi middleware that validates Cognito JWT tokens.
// When api.AuthEnabled is false, this middleware is a no-op (passes through).
// Skips auth for /api/health (ALB health check) and /api/auth/config (SPA config).
func JWTAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip auth if disabled
		if !api.AuthEnabled {
			next.ServeHTTP(w, r)
			return
		}

		// Skip auth for health check and auth config endpoints
		path := r.URL.Path
		if path == "/api/health" || path == "/api/auth/config" {
			next.ServeHTTP(w, r)
			return
		}

		// Skip auth for non-API paths (SPA static files)
		if !strings.HasPrefix(path, "/api/") && !strings.HasPrefix(path, "/ws/") {
			next.ServeHTTP(w, r)
			return
		}

		// Skip auth for WebSocket (auth via query param or cookie instead)
		if strings.HasPrefix(path, "/ws/") {
			next.ServeHTTP(w, r)
			return
		}

		// Extract Bearer token
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, `{"detail":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, `{"detail":"Invalid authorization header"}`, http.StatusUnauthorized)
			return
		}

		payload, err := validateJWT(parts[1])
		if err != nil {
			slog.Warn("jwt_validation_failed", "error", err.Error())
			http.Error(w, `{"detail":"Invalid or expired token"}`, http.StatusUnauthorized)
			return
		}

		// Add identity to request headers for downstream logging
		identity := payload.Email
		if identity == "" {
			identity = payload.Sub
		}
		r.Header.Set("X-Amzn-Oidc-Identity", identity)

		next.ServeHTTP(w, r)
	})
}

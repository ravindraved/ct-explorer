// Package api provides shared helpers for the CT Explorer API server.
//
// # Auth Configuration
//
// Authentication is enabled by default using Cognito. The SPA redirects to
// the Cognito Hosted UI for login and receives JWT tokens via the implicit
// OAuth flow. The Go backend validates the JWT on every API request.
//
// To disable auth (e.g., local development on EC2/laptop), set the
// CT_AUTH_ENABLED environment variable to "false". When disabled, the JWT
// middleware is skipped and the app works without login — same as the
// original single-binary behavior.
//
// For Fargate deployment, set the Cognito environment variables to match
// the deployed Cognito User Pool. For local use, leave them unset and
// disable auth.
package api

import (
	"fmt"
	"os"
	"strings"
)

// --------------------------------------------------------------------------
// Auth Configuration (loaded from environment variables at startup)
//
// CT_AUTH_ENABLED:      "true" = Cognito login required (default)
//                       "false" = no login, open access (local use)
// CT_COGNITO_DOMAIN:    Cognito Hosted UI base URL
// CT_COGNITO_CLIENT_ID: Cognito User Pool Client ID
// CT_COGNITO_POOL_ID:   Cognito User Pool ID (for JWT issuer validation)
// CT_COGNITO_REGION:    AWS region of the Cognito User Pool (default: ap-south-1)
// --------------------------------------------------------------------------

var (
	// AuthEnabled controls whether Cognito JWT authentication is enforced.
	// Default: true (Cognito login required).
	// Set CT_AUTH_ENABLED=false for local EC2/laptop use without authentication.
	AuthEnabled = envBool("CT_AUTH_ENABLED", true)

	// CognitoDomain is the Cognito Hosted UI base URL.
	CognitoDomain = envOr("CT_COGNITO_DOMAIN", "")

	// CognitoClientID is the Cognito User Pool Client ID.
	CognitoClientID = envOr("CT_COGNITO_CLIENT_ID", "")

	// CognitoUserPoolID is the Cognito User Pool ID (for JWT issuer validation).
	CognitoUserPoolID = envOr("CT_COGNITO_POOL_ID", "")

	// CognitoRegion is the AWS region where the Cognito User Pool is deployed.
	CognitoRegion = envOr("CT_COGNITO_REGION", "ap-south-1")

	// CognitoJWKSURL is the URL for Cognito's JSON Web Key Set (for JWT signature verification).
	// Derived from CognitoRegion and CognitoUserPoolID.
	CognitoJWKSURL = cognitoJWKSURL()
)

// envOr returns the value of the environment variable or the fallback.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// envBool returns the boolean value of the environment variable or the fallback.
// Recognizes "true", "1", "yes" as true; everything else (including unset) uses fallback.
func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	switch strings.ToLower(v) {
	case "true", "1", "yes":
		return true
	case "false", "0", "no":
		return false
	default:
		return fallback
	}
}

// cognitoJWKSURL builds the JWKS URL from region and pool ID.
// Returns empty string if pool ID is not configured.
func cognitoJWKSURL() string {
	if CognitoUserPoolID == "" {
		return ""
	}
	return fmt.Sprintf(
		"https://cognito-idp.%s.amazonaws.com/%s/.well-known/jwks.json",
		CognitoRegion, CognitoUserPoolID,
	)
}

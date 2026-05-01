package api

import (
	"strings"
	"sync"
)

// MaskAccountID masks an AWS account ID, showing only the last 5 digits.
// Example: "084828584616" → "*******84616"
// Maintains the original string length. Non-12-digit strings are returned as-is.
func MaskAccountID(id string) string {
	if len(id) != 12 {
		return id
	}
	return strings.Repeat("*", 7) + id[7:]
}

// MaskARN masks the account ID portion of an AWS ARN.
// Example: "arn:aws:organizations::084828584616:account/..." → "arn:aws:organizations::*******84616:account/..."
func MaskARN(arn string) string {
	// AWS ARNs have account ID at position 4 (0-indexed) when split by ":"
	// arn:aws:service:region:account-id:resource
	parts := strings.SplitN(arn, ":", 6)
	if len(parts) >= 5 && len(parts[4]) == 12 {
		parts[4] = MaskAccountID(parts[4])
		return strings.Join(parts, ":")
	}
	return arn
}

// AccountIDResolver maintains a server-side mapping of masked → real account IDs.
// This allows the frontend to send masked IDs and the backend to resolve them.
type AccountIDResolver struct {
	mu           sync.RWMutex
	maskedToReal map[string]string
}

// NewAccountIDResolver creates a new resolver.
func NewAccountIDResolver() *AccountIDResolver {
	return &AccountIDResolver{
		maskedToReal: make(map[string]string),
	}
}

// Register adds a real account ID to the resolver and returns the masked version.
func (r *AccountIDResolver) Register(realID string) string {
	masked := MaskAccountID(realID)
	r.mu.Lock()
	r.maskedToReal[masked] = realID
	r.mu.Unlock()
	return masked
}

// Resolve returns the real account ID for a given input.
// Accepts both masked and real IDs — tries masked lookup first, then returns input as-is.
func (r *AccountIDResolver) Resolve(input string) string {
	r.mu.RLock()
	real, ok := r.maskedToReal[input]
	r.mu.RUnlock()
	if ok {
		return real
	}
	return input // might be a real ID from internal calls
}

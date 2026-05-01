package models

// ErrorInfo represents an error encountered during AWS API operations.
type ErrorInfo struct {
	Source      string `json:"source"`
	Message     string `json:"message"`
	Recoverable bool   `json:"recoverable"`
}

package models

// ControlType represents the type of a Control Tower control.
type ControlType string

const (
	ControlTypePreventive ControlType = "PREVENTIVE"
	ControlTypeDetective  ControlType = "DETECTIVE"
	ControlTypeProactive  ControlType = "PROACTIVE"
)

// EnforcementStatus represents the enforcement status of a control.
type EnforcementStatus string

const (
	EnforcementEnabled EnforcementStatus = "ENABLED"
	EnforcementFailed  EnforcementStatus = "FAILED"
)

// Control represents an enabled Control Tower control on a target.
type Control struct {
	ARN         string            `json:"arn"`
	ControlID   string            `json:"control_id"`
	Name        string            `json:"name"`
	ControlType ControlType       `json:"control_type"`
	Enforcement EnforcementStatus `json:"enforcement"`
	TargetID    string            `json:"target_id"`
	Description string            `json:"description"`
}

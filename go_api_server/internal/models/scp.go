package models

// Policy represents a parsed SCP policy document.
type Policy struct {
	Version   string           `json:"version"`
	Statement []map[string]any `json:"statement"`
}

// NewPolicy creates a Policy with a properly initialized Statement slice.
func NewPolicy() Policy {
	return Policy{
		Statement: make([]map[string]any, 0),
	}
}

// SCP represents a Service Control Policy.
type SCP struct {
	ID             string   `json:"id"`
	ARN            string   `json:"arn"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	PolicyDocument string   `json:"policy_document"`
	Policy         *Policy  `json:"policy"`
	TargetIDs      []string `json:"target_ids"`
}

// NewSCP creates an SCP with properly initialized slices.
func NewSCP() SCP {
	return SCP{
		TargetIDs: make([]string, 0),
	}
}

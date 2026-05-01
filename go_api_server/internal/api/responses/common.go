package responses

// ErrorResponse is the standard error shape matching Python's {"detail": "message"}.
type ErrorResponse struct {
	Detail string `json:"detail"`
}

// AuthStatusResponse matches Python's AuthStatusResponse.
type AuthStatusResponse struct {
	Authenticated bool    `json:"authenticated"`
	AccountID     *string `json:"account_id"`
	Region        *string `json:"region"`
	Error         *string `json:"error"`
	AuthMode      string  `json:"auth_mode"`
}

// AuthMetadataResponse matches Python's AuthMetadataResponse.
type AuthMetadataResponse struct {
	Available        bool    `json:"available"`
	InstanceID       *string `json:"instance_id"`
	InstanceType     *string `json:"instance_type"`
	AvailabilityZone *string `json:"availability_zone"`
	Region           *string `json:"region"`
	IAMRole          *string `json:"iam_role"`
	AccountID        *string `json:"account_id"`
	AuthMode         string  `json:"auth_mode"`
}

// AuthConfigureRequest is the request body for POST /api/auth/configure.
type AuthConfigureRequest struct {
	AuthMode       string  `json:"auth_mode"`
	AccessKeyID    *string `json:"access_key_id"`
	SecretAccessKey *string `json:"secret_access_key"`
	SessionToken   *string `json:"session_token"`
	ProfileName    *string `json:"profile_name"`
	Region         *string `json:"region"`
}

// AuthConfigureResponse matches Python's AuthConfigureResponse.
type AuthConfigureResponse struct {
	Success   bool    `json:"success"`
	AccountID *string `json:"account_id"`
	Region    *string `json:"region"`
	AuthMode  string  `json:"auth_mode"`
	Error     *string `json:"error"`
}

// RefreshStatusResponse matches Python's RefreshStatusResponse.
type RefreshStatusResponse struct {
	Status          string  `json:"status"`
	Phase           *string `json:"phase"`
	Error           *string `json:"error"`
	LastRefreshedAt *string `json:"last_refreshed_at"`
}

// OrgTreeNode represents a recursive org tree node matching Python's OrgTreeNode.
type OrgTreeNode struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Type     string        `json:"type"`
	ARN      string        `json:"arn"`
	Children []OrgTreeNode `json:"children"`
	Email    *string       `json:"email,omitempty"`
	Status   *string       `json:"status,omitempty"`
}

// NewOrgTreeNode creates an OrgTreeNode with properly initialized slices.
func NewOrgTreeNode() OrgTreeNode {
	return OrgTreeNode{
		Children: make([]OrgTreeNode, 0),
	}
}

// AccountDetailResponse represents account details with inherited controls and SCPs.
type AccountDetailResponse struct {
	ID       string               `json:"id"`
	Name     string               `json:"name"`
	ARN      string               `json:"arn"`
	Email    string               `json:"email"`
	Status   string               `json:"status"`
	OUID     string               `json:"ou_id"`
	OUName   string               `json:"ou_name"`
	Controls []ControlResponse    `json:"controls"`
	SCPs     []SCPSummaryResponse `json:"scps"`
}

// NewAccountDetailResponse creates an AccountDetailResponse with properly initialized slices.
func NewAccountDetailResponse() AccountDetailResponse {
	return AccountDetailResponse{
		Controls: make([]ControlResponse, 0),
		SCPs:     make([]SCPSummaryResponse, 0),
	}
}

// ControlResponse matches Python's ControlResponse.
type ControlResponse struct {
	ARN         string `json:"arn"`
	ControlID   string `json:"control_id"`
	Name        string `json:"name"`
	ControlType string `json:"control_type"`
	Enforcement string `json:"enforcement"`
	TargetID    string `json:"target_id"`
	Description string `json:"description"`
}

// SCPSummaryResponse matches Python's ScpSummaryResponse.
type SCPSummaryResponse struct {
	ID          string   `json:"id"`
	ARN         string   `json:"arn"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	TargetIDs   []string `json:"target_ids"`
}

// NewSCPSummaryResponse creates an SCPSummaryResponse with properly initialized slices.
func NewSCPSummaryResponse() SCPSummaryResponse {
	return SCPSummaryResponse{
		TargetIDs: make([]string, 0),
	}
}

// PolicyResponse matches Python's PolicyResponse.
type PolicyResponse struct {
	Version   string           `json:"version"`
	Statement []map[string]any `json:"statement"`
}

// NewPolicyResponse creates a PolicyResponse with properly initialized slices.
func NewPolicyResponse() PolicyResponse {
	return PolicyResponse{
		Statement: make([]map[string]any, 0),
	}
}

// SCPDetailResponse matches Python's ScpDetailResponse (extends SCPSummaryResponse).
type SCPDetailResponse struct {
	ID             string          `json:"id"`
	ARN            string          `json:"arn"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	TargetIDs      []string        `json:"target_ids"`
	PolicyDocument string          `json:"policy_document"`
	Policy         *PolicyResponse `json:"policy"`
}

// NewSCPDetailResponse creates an SCPDetailResponse with properly initialized slices.
func NewSCPDetailResponse() SCPDetailResponse {
	return SCPDetailResponse{
		TargetIDs: make([]string, 0),
	}
}

package models

// OrgUnit represents an AWS Organizations organizational unit.
type OrgUnit struct {
	ID          string   `json:"id"`
	Name        string   `json:"name"`
	ARN         string   `json:"arn"`
	ParentID    string   `json:"parent_id"`
	ChildrenIDs []string `json:"children_ids"`
	AccountIDs  []string `json:"account_ids"`
}

// NewOrgUnit creates an OrgUnit with properly initialized slices.
func NewOrgUnit() OrgUnit {
	return OrgUnit{
		ChildrenIDs: make([]string, 0),
		AccountIDs:  make([]string, 0),
	}
}

// Account represents an AWS account within an organization.
type Account struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	ARN    string `json:"arn"`
	Email  string `json:"email"`
	Status string `json:"status"`
	OUID   string `json:"ou_id"`
}

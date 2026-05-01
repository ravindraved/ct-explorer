package models

// CatalogBehavior represents the behavior type of a catalog control.
type CatalogBehavior string

const (
	BehaviorPreventive CatalogBehavior = "PREVENTIVE"
	BehaviorDetective  CatalogBehavior = "DETECTIVE"
	BehaviorProactive  CatalogBehavior = "PROACTIVE"
)

// CatalogSeverity represents the severity level of a catalog control.
type CatalogSeverity string

const (
	SeverityLow      CatalogSeverity = "LOW"
	SeverityMedium   CatalogSeverity = "MEDIUM"
	SeverityHigh     CatalogSeverity = "HIGH"
	SeverityCritical CatalogSeverity = "CRITICAL"
)

// CatalogDomain represents a control catalog domain.
type CatalogDomain struct {
	ARN         string `json:"arn"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CatalogObjective represents a control catalog objective.
type CatalogObjective struct {
	ARN         string `json:"arn"`
	Name        string `json:"name"`
	Description string `json:"description"`
	DomainARN   string `json:"domain_arn"`
}

// CatalogCommonControl represents a control catalog common control.
type CatalogCommonControl struct {
	ARN          string `json:"arn"`
	Name         string `json:"name"`
	Description  string `json:"description"`
	ObjectiveARN string `json:"objective_arn"`
	DomainARN    string `json:"domain_arn"`
}

// CatalogControl represents a control from the AWS Control Catalog.
type CatalogControl struct {
	ARN                      string          `json:"arn"`
	Name                     string          `json:"name"`
	Description              string          `json:"description"`
	Behavior                 CatalogBehavior `json:"behavior"`
	Severity                 CatalogSeverity `json:"severity"`
	Aliases                  []string        `json:"aliases"`
	GovernedResources        []string        `json:"governed_resources"`
	ImplementationType       string          `json:"implementation_type"`
	ImplementationIdentifier string          `json:"implementation_identifier"`
	CreateTime               string          `json:"create_time"`
	CommonControlARN         string          `json:"common_control_arn"`
}

// NewCatalogControl creates a CatalogControl with properly initialized slices.
func NewCatalogControl() CatalogControl {
	return CatalogControl{
		Aliases:           make([]string, 0),
		GovernedResources: make([]string, 0),
	}
}

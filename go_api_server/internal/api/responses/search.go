package responses

// SearchResultItem matches Python's SearchResultItem.
type SearchResultItem struct {
	EntityType  string         `json:"entity_type"`
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	ARN         *string        `json:"arn,omitempty"`
	Description *string        `json:"description,omitempty"`
	Metadata    map[string]any `json:"metadata"`
}

// NewSearchResultItem creates a SearchResultItem with properly initialized maps.
func NewSearchResultItem() SearchResultItem {
	return SearchResultItem{
		Metadata: make(map[string]any),
	}
}

// GroupedResults matches Python's GroupedResults.
type GroupedResults struct {
	EntityType string             `json:"entity_type"`
	Count      int                `json:"count"`
	Items      []SearchResultItem `json:"items"`
}

// NewGroupedResults creates a GroupedResults with properly initialized slices.
func NewGroupedResults() GroupedResults {
	return GroupedResults{
		Items: make([]SearchResultItem, 0),
	}
}

// PaginatedSearchResponse matches Python's PaginatedSearchResponse.
type PaginatedSearchResponse struct {
	Items     []SearchResultItem `json:"items"`
	Groups    []GroupedResults   `json:"groups"`
	Total     int                `json:"total"`
	Page      int                `json:"page"`
	PageCount int                `json:"page_count"`
	Mode      string             `json:"mode"`
}

// NewPaginatedSearchResponse creates a PaginatedSearchResponse with properly initialized slices.
func NewPaginatedSearchResponse() PaginatedSearchResponse {
	return PaginatedSearchResponse{
		Items:  make([]SearchResultItem, 0),
		Groups: make([]GroupedResults, 0),
	}
}

// CoverageItem matches Python's CoverageItemResponse.
type CoverageItem struct {
	ARN        string         `json:"arn"`
	Name       string         `json:"name"`
	EntityType string         `json:"entity_type"`
	IsCovered  bool           `json:"is_covered"`
	Metadata   map[string]any `json:"metadata"`
}

// NewCoverageItem creates a CoverageItem with properly initialized maps.
func NewCoverageItem() CoverageItem {
	return CoverageItem{
		Metadata: make(map[string]any),
	}
}

// CoverageResponse matches Python's CoverageResponse.
type CoverageResponse struct {
	TargetID      string         `json:"target_id"`
	TargetType    string         `json:"target_type"`
	TargetName    string         `json:"target_name"`
	TotalControls int            `json:"total_controls"`
	EnabledCount  int            `json:"enabled_count"`
	GapCount      int            `json:"gap_count"`
	SCPCount      int            `json:"scp_count"`
	AccountCount  int            `json:"account_count"`
	Items         []CoverageItem `json:"items"`
	Mode          string         `json:"mode"`
}

// NewCoverageResponse creates a CoverageResponse with properly initialized slices.
func NewCoverageResponse() CoverageResponse {
	return CoverageResponse{
		Items: make([]CoverageItem, 0),
		Mode:  "coverage",
	}
}

// PathNode matches Python's PathNodeResponse.
type PathNode struct {
	OUID     string             `json:"ou_id"`
	OUName   string             `json:"ou_name"`
	OUARN    string             `json:"ou_arn"`
	Controls []SearchResultItem `json:"controls"`
	SCPs     []SearchResultItem `json:"scps"`
}

// NewPathNode creates a PathNode with properly initialized slices.
func NewPathNode() PathNode {
	return PathNode{
		Controls: make([]SearchResultItem, 0),
		SCPs:     make([]SearchResultItem, 0),
	}
}

// PathResponse matches Python's PathResponse.
type PathResponse struct {
	AccountID     string     `json:"account_id"`
	AccountName   string     `json:"account_name"`
	Chain         []PathNode `json:"chain"`
	TotalControls int        `json:"total_controls"`
	TotalSCPs     int        `json:"total_scps"`
	Mode          string     `json:"mode"`
}

// NewPathResponse creates a PathResponse with properly initialized slices.
func NewPathResponse() PathResponse {
	return PathResponse{
		Chain: make([]PathNode, 0),
		Mode:  "path",
	}
}

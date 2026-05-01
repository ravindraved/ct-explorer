package handlers

import (
	"math"
	"net/http"
	"strconv"
	"strings"

	"go_api_server/internal/api"
	"go_api_server/internal/api/responses"
	"go_api_server/internal/search"
	"go_api_server/internal/store"
)

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func parseIntDefault(s string, def int) int {
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return def
	}
	return n
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		t := strings.TrimSpace(p)
		if t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func toSearchResultItem(item search.FindResultItem) responses.SearchResultItem {
	arn := item.ARN
	desc := item.Description
	metadata := item.Metadata
	if metadata == nil {
		metadata = make(map[string]any)
	}
	return responses.SearchResultItem{
		EntityType:  item.EntityType,
		ID:          item.ID,
		Name:        item.Name,
		ARN:         &arn,
		Description: &desc,
		Metadata:    metadata,
	}
}

func toGroupedResults(items []responses.SearchResultItem, groups map[string]int) []responses.GroupedResults {
	result := make([]responses.GroupedResults, 0, len(groups))
	for et, count := range groups {
		groupItems := make([]responses.SearchResultItem, 0)
		for _, item := range items {
			if item.EntityType == et {
				groupItems = append(groupItems, item)
			}
		}
		result = append(result, responses.GroupedResults{
			EntityType: et,
			Count:      count,
			Items:      groupItems,
		})
	}
	return result
}

func toCoverageItem(item search.CoverageItem) responses.CoverageItem {
	metadata := item.Metadata
	if metadata == nil {
		metadata = make(map[string]any)
	}
	return responses.CoverageItem{
		ARN:        item.ARN,
		Name:       item.Name,
		EntityType: item.EntityType,
		IsCovered:  item.IsCovered,
		Metadata:   metadata,
	}
}

func toPathNode(node search.PathNode) responses.PathNode {
	controls := make([]responses.SearchResultItem, 0, len(node.Controls))
	for _, c := range node.Controls {
		controls = append(controls, toSearchResultItem(c))
	}
	scps := make([]responses.SearchResultItem, 0, len(node.SCPs))
	for _, s := range node.SCPs {
		scps = append(scps, toSearchResultItem(s))
	}
	return responses.PathNode{
		OUID:     node.OUID,
		OUName:   node.OUName,
		OUARN:    node.OUARN,
		Controls: controls,
		SCPs:     scps,
	}
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// SearchFind godoc
// @Summary      Search across all entities
// @Description  Case-insensitive text search with facet filters and pagination
// @Tags         search
// @Produce      json
// @Param        q             query  string  false  "Search query"
// @Param        entity_types  query  string  false  "Comma-separated entity type filter"
// @Param        behaviors     query  string  false  "Comma-separated behavior filter"
// @Param        severities    query  string  false  "Comma-separated severity filter"
// @Param        page          query  int     false  "Page number"  default(1)
// @Param        page_size     query  int     false  "Page size"    default(50)
// @Success      200  {object}  responses.PaginatedSearchResponse
// @Failure      400  {object}  responses.ErrorResponse
// @Router       /api/search/find [get]
func SearchFind(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		query := q.Get("q")
		entityTypes := splitCSV(q.Get("entity_types"))
		behaviors := splitCSV(q.Get("behaviors"))
		severities := splitCSV(q.Get("severities"))
		page := parseIntDefault(q.Get("page"), 1)
		pageSize := parseIntDefault(q.Get("page_size"), 50)

		if page < 1 || pageSize < 1 {
			api.WriteError(w, http.StatusBadRequest, "page and page_size must be positive integers")
			return
		}

		engine := search.NewEngine(s)
		result := engine.Find(query, entityTypes, behaviors, severities, page, pageSize)

		items := make([]responses.SearchResultItem, 0, len(result.Items))
		for _, item := range result.Items {
			items = append(items, toSearchResultItem(item))
		}

		groups := toGroupedResults(items, result.Groups)

		pageCount := 1
		if result.Total > 0 {
			pageCount = int(math.Ceil(float64(result.Total) / float64(pageSize)))
			if pageCount < 1 {
				pageCount = 1
			}
		}

		resp := responses.PaginatedSearchResponse{
			Items:     items,
			Groups:    groups,
			Total:     result.Total,
			Page:      page,
			PageCount: pageCount,
			Mode:      "find",
		}
		api.WriteJSON(w, http.StatusOK, resp)
	}
}

// SearchCoverage godoc
// @Summary      Get posture coverage
// @Description  Returns posture coverage for an OU, account, or ontology node
// @Tags         search
// @Produce      json
// @Param        target_type  query  string  true  "Target type (ou, account, domain, objective, common_control)"
// @Param        target_id    query  string  true  "Target ID or ARN"
// @Success      200  {object}  responses.CoverageResponse
// @Failure      400  {object}  responses.ErrorResponse
// @Router       /api/search/coverage [get]
func SearchCoverage(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		targetType := r.URL.Query().Get("target_type")
		targetID := r.URL.Query().Get("target_id")

		if targetID == "" {
			api.WriteError(w, http.StatusBadRequest, "target_id is required")
			return
		}

		engine := search.NewEngine(s)
		result, err := engine.Coverage(targetType, targetID)
		if err != nil {
			msg := err.Error()
			if strings.Contains(msg, "invalid target_type") || strings.Contains(msg, "Invalid target_type") {
				api.WriteError(w, http.StatusBadRequest, msg)
				return
			}
			api.WriteError(w, http.StatusNotFound, msg)
			return
		}

		items := make([]responses.CoverageItem, 0, len(result.Items))
		for _, item := range result.Items {
			items = append(items, toCoverageItem(item))
		}

		resp := responses.CoverageResponse{
			TargetID:      result.TargetID,
			TargetType:    result.TargetType,
			TargetName:    result.TargetName,
			TotalControls: result.TotalControls,
			EnabledCount:  result.EnabledCount,
			GapCount:      result.GapCount,
			SCPCount:      result.SCPCount,
			AccountCount:  result.AccountCount,
			Items:         items,
			Mode:          "coverage",
		}
		api.WriteJSON(w, http.StatusOK, resp)
	}
}

// SearchPath godoc
// @Summary      Get OU inheritance path
// @Description  Returns the OU inheritance chain for an account from root to leaf
// @Tags         search
// @Produce      json
// @Param        account_id  query  string  true  "Account ID"
// @Success      200  {object}  responses.PathResponse
// @Failure      400  {object}  responses.ErrorResponse
// @Router       /api/search/path [get]
func SearchPath(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		accountID := r.URL.Query().Get("account_id")
		if accountID == "" {
			api.WriteError(w, http.StatusBadRequest, "account_id is required")
			return
		}

		engine := search.NewEngine(s)
		result, err := engine.Path(accountID)
		if err != nil {
			api.WriteError(w, http.StatusNotFound, err.Error())
			return
		}

		chain := make([]responses.PathNode, 0, len(result.Chain))
		for _, node := range result.Chain {
			chain = append(chain, toPathNode(node))
		}

		resp := responses.PathResponse{
			AccountID:     result.AccountID,
			AccountName:   result.AccountName,
			Chain:         chain,
			TotalControls: result.TotalControls,
			TotalSCPs:     result.TotalSCPs,
			Mode:          "path",
		}
		api.WriteJSON(w, http.StatusOK, resp)
	}
}

// SearchQuickQuery godoc
// @Summary      Execute preset compliance query
// @Description  Runs a preset compliance query with pagination
// @Tags         search
// @Produce      json
// @Param        preset     query  string  true   "Preset name"
// @Param        page       query  int     false  "Page number"  default(1)
// @Param        page_size  query  int     false  "Page size"    default(50)
// @Success      200  {object}  responses.PaginatedSearchResponse
// @Failure      400  {object}  responses.ErrorResponse
// @Router       /api/search/quick-query [get]
func SearchQuickQuery(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		preset := q.Get("preset")
		page := parseIntDefault(q.Get("page"), 1)
		pageSize := parseIntDefault(q.Get("page_size"), 50)

		if page < 1 || pageSize < 1 {
			api.WriteError(w, http.StatusBadRequest, "page and page_size must be positive integers")
			return
		}

		engine := search.NewEngine(s)
		result, err := engine.QuickQuery(preset, page, pageSize)
		if err != nil {
			api.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}

		items := make([]responses.SearchResultItem, 0, len(result.Items))
		for _, item := range result.Items {
			items = append(items, toSearchResultItem(item))
		}

		pageCount := 1
		if result.Total > 0 {
			pageCount = int(math.Ceil(float64(result.Total) / float64(pageSize)))
			if pageCount < 1 {
				pageCount = 1
			}
		}

		resp := responses.PaginatedSearchResponse{
			Items:     items,
			Groups:    make([]responses.GroupedResults, 0),
			Total:     result.Total,
			Page:      page,
			PageCount: pageCount,
			Mode:      "quick_query",
		}
		api.WriteJSON(w, http.StatusOK, resp)
	}
}

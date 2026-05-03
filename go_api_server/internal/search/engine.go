// Package search provides a 4-mode search engine operating against the Store.
// All methods are pure reads with no side effects.
package search

import (
	"fmt"
	"math"
	"sort"
	"strings"

	"go_api_server/internal/models"
	"go_api_server/internal/store"
)

// ---------------------------------------------------------------------------
// Result types
// ---------------------------------------------------------------------------

// FindResultItem represents a single search result item.
type FindResultItem struct {
	EntityType  string         `json:"entity_type"`
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	ARN         string         `json:"arn"`
	Description string         `json:"description"`
	Metadata    map[string]any `json:"metadata"`
}

// GroupCount holds a count for an entity type group.
type GroupCount struct {
	EntityType string `json:"entity_type"`
	Count      int    `json:"count"`
}

// FindResult holds the result of a Find search.
type FindResult struct {
	Items     []FindResultItem  `json:"items"`
	Groups    map[string]int    `json:"groups"`
	Total     int               `json:"total"`
}

// CoverageItem represents a single item in a coverage result.
type CoverageItem struct {
	ARN        string         `json:"arn"`
	Name       string         `json:"name"`
	EntityType string         `json:"entity_type"`
	IsCovered  bool           `json:"is_covered"`
	Metadata   map[string]any `json:"metadata"`
}

// CoverageResult holds the result of a Coverage search.
type CoverageResult struct {
	TargetID      string         `json:"target_id"`
	TargetType    string         `json:"target_type"`
	TargetName    string         `json:"target_name"`
	Items         []CoverageItem `json:"items"`
	TotalControls int            `json:"total_controls"`
	EnabledCount  int            `json:"enabled_count"`
	GapCount      int            `json:"gap_count"`
	SCPCount      int            `json:"scp_count"`
	AccountCount  int            `json:"account_count"`
}

// PathNode represents a single OU in the inheritance chain.
type PathNode struct {
	OUID     string           `json:"ou_id"`
	OUName   string           `json:"ou_name"`
	OUARN    string           `json:"ou_arn"`
	Controls []FindResultItem `json:"controls"`
	SCPs     []FindResultItem `json:"scps"`
}

// PathResult holds the result of a Path search.
type PathResult struct {
	AccountID     string     `json:"account_id"`
	AccountName   string     `json:"account_name"`
	Chain         []PathNode `json:"chain"`
	TotalControls int        `json:"total_controls"`
	TotalSCPs     int        `json:"total_scps"`
}

// QuickQueryResult holds the result of a QuickQuery search.
type QuickQueryResult struct {
	Items []FindResultItem `json:"items"`
	Total int              `json:"total"`
}

// ---------------------------------------------------------------------------
// Engine
// ---------------------------------------------------------------------------

// Engine provides multi-mode search against the Store.
type Engine struct {
	store *store.Store
}

// NewEngine creates a new search Engine.
func NewEngine(s *store.Store) *Engine {
	return &Engine{store: s}
}

// validPresets lists the allowed quick-query preset names.
var validPresets = map[string]bool{
	"unenabled_critical_controls":      true,
	"ous_without_detective_controls":   true,
	"accounts_without_scps":            true,
	"uncovered_ontology_objectives":    true,
}

// ValidPresetNames returns the sorted list of valid preset names.
func ValidPresetNames() []string {
	names := make([]string, 0, len(validPresets))
	for k := range validPresets {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// ---------------------------------------------------------------------------
// Find mode
// ---------------------------------------------------------------------------

// Find performs a case-insensitive text search across all entity collections.
func (e *Engine) Find(q string, entityTypes, behaviors, severities []string, page, pageSize int) FindResult {
	if strings.TrimSpace(q) == "" {
		return FindResult{
			Items:  make([]FindResultItem, 0),
			Groups: make(map[string]int),
			Total:  0,
		}
	}

	query := strings.ToLower(q)
	allItems := make([]FindResultItem, 0)

	allItems = append(allItems, e.findOUs(query)...)
	allItems = append(allItems, e.findAccounts(query)...)
	allItems = append(allItems, e.findControls(query)...)
	allItems = append(allItems, e.findSCPs(query)...)
	allItems = append(allItems, e.findCatalogControls(query)...)
	allItems = append(allItems, e.findCatalogDomains(query)...)
	allItems = append(allItems, e.findCatalogObjectives(query)...)
	allItems = append(allItems, e.findCatalogCommonControls(query)...)

	// Apply facet filters
	allItems = applyFilters(allItems, entityTypes, behaviors, severities)

	// Group counts (full set, before pagination)
	groups := make(map[string]int)
	for _, item := range allItems {
		groups[item.EntityType]++
	}

	total := len(allItems)

	// Paginate
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	pageItems := make([]FindResultItem, len(allItems[start:end]))
	copy(pageItems, allItems[start:end])

	return FindResult{
		Items:  pageItems,
		Groups: groups,
		Total:  total,
	}
}

// ---------------------------------------------------------------------------
// Find helpers — per-entity matching
// ---------------------------------------------------------------------------

func matches(query string, fields []string) bool {
	for _, f := range fields {
		if f != "" && strings.Contains(strings.ToLower(f), query) {
			return true
		}
	}
	return false
}

func (e *Engine) findOUs(query string) []FindResultItem {
	items := make([]FindResultItem, 0)
	for _, ou := range e.store.AllOUs() {
		if matches(query, []string{ou.ID, ou.Name, ou.ARN}) {
			items = append(items, FindResultItem{
				EntityType:  "ou",
				ID:          ou.ID,
				Name:        ou.Name,
				ARN:         ou.ARN,
				Description: "",
				Metadata:    map[string]any{"parent_id": ou.ParentID},
			})
		}
	}
	return items
}

func (e *Engine) findAccounts(query string) []FindResultItem {
	items := make([]FindResultItem, 0)
	for _, acct := range e.store.AllAccounts() {
		if matches(query, []string{acct.ID, acct.Name, acct.ARN, acct.Email}) {
			items = append(items, FindResultItem{
				EntityType:  "account",
				ID:          acct.ID,
				Name:        acct.Name,
				ARN:         acct.ARN,
				Description: "",
				Metadata:    map[string]any{"email": acct.Email, "status": acct.Status, "ou_id": acct.OUID},
			})
		}
	}
	return items
}

func (e *Engine) findControls(query string) []FindResultItem {
	items := make([]FindResultItem, 0)
	for _, ctrl := range e.store.AllControls() {
		fields := []string{
			ctrl.ARN, ctrl.ControlID, ctrl.Name, ctrl.Description,
			string(ctrl.ControlType), string(ctrl.Enforcement), ctrl.TargetID,
		}
		if matches(query, fields) {
			items = append(items, FindResultItem{
				EntityType:  "control",
				ID:          ctrl.ControlID,
				Name:        ctrl.Name,
				ARN:         ctrl.ARN,
				Description: ctrl.Description,
				Metadata: map[string]any{
					"control_type": string(ctrl.ControlType),
					"enforcement":  string(ctrl.Enforcement),
					"target_id":    ctrl.TargetID,
				},
			})
		}
	}
	return items
}

func (e *Engine) findSCPs(query string) []FindResultItem {
	items := make([]FindResultItem, 0)
	for _, scp := range e.store.AllSCPs() {
		if matches(query, []string{scp.ID, scp.ARN, scp.Name, scp.Description}) {
			items = append(items, FindResultItem{
				EntityType:  "scp",
				ID:          scp.ID,
				Name:        scp.Name,
				ARN:         scp.ARN,
				Description: scp.Description,
				Metadata:    map[string]any{"target_ids": scp.TargetIDs},
			})
		}
	}
	return items
}

func (e *Engine) findCatalogControls(query string) []FindResultItem {
	items := make([]FindResultItem, 0)
	for _, cc := range e.store.AllCatalogControls() {
		fields := []string{
			cc.ARN, cc.Name, cc.Description,
			string(cc.Behavior), string(cc.Severity),
			strings.Join(cc.Aliases, " "), cc.ImplementationType,
		}
		if matches(query, fields) {
			items = append(items, FindResultItem{
				EntityType:  "catalog_control",
				ID:          cc.ARN,
				Name:        cc.Name,
				ARN:         cc.ARN,
				Description: cc.Description,
				Metadata: map[string]any{
					"behavior":            string(cc.Behavior),
					"severity":            string(cc.Severity),
					"implementation_type": cc.ImplementationType,
				},
			})
		}
	}
	return items
}

func (e *Engine) findCatalogDomains(query string) []FindResultItem {
	items := make([]FindResultItem, 0)
	for _, d := range e.store.AllCatalogDomains() {
		if matches(query, []string{d.ARN, d.Name, d.Description}) {
			items = append(items, FindResultItem{
				EntityType:  "catalog_domain",
				ID:          d.ARN,
				Name:        d.Name,
				ARN:         d.ARN,
				Description: d.Description,
				Metadata:    make(map[string]any),
			})
		}
	}
	return items
}

func (e *Engine) findCatalogObjectives(query string) []FindResultItem {
	items := make([]FindResultItem, 0)
	for _, o := range e.store.AllCatalogObjectives() {
		if matches(query, []string{o.ARN, o.Name, o.Description}) {
			items = append(items, FindResultItem{
				EntityType:  "catalog_objective",
				ID:          o.ARN,
				Name:        o.Name,
				ARN:         o.ARN,
				Description: o.Description,
				Metadata:    map[string]any{"domain_arn": o.DomainARN},
			})
		}
	}
	return items
}

func (e *Engine) findCatalogCommonControls(query string) []FindResultItem {
	items := make([]FindResultItem, 0)
	for _, cc := range e.store.AllCatalogCommonControls() {
		if matches(query, []string{cc.ARN, cc.Name, cc.Description}) {
			items = append(items, FindResultItem{
				EntityType:  "catalog_common_control",
				ID:          cc.ARN,
				Name:        cc.Name,
				ARN:         cc.ARN,
				Description: cc.Description,
				Metadata:    map[string]any{"objective_arn": cc.ObjectiveARN},
			})
		}
	}
	return items
}

// ---------------------------------------------------------------------------
// Find helpers — filtering
// ---------------------------------------------------------------------------

func applyFilters(items []FindResultItem, entityTypes, behaviors, severities []string) []FindResultItem {
	result := items

	if len(entityTypes) > 0 {
		etSet := make(map[string]bool, len(entityTypes))
		for _, t := range entityTypes {
			etSet[strings.ToLower(t)] = true
		}
		filtered := make([]FindResultItem, 0)
		for _, item := range result {
			if etSet[item.EntityType] {
				filtered = append(filtered, item)
			}
		}
		result = filtered
	}

	if len(behaviors) > 0 {
		behSet := make(map[string]bool, len(behaviors))
		for _, b := range behaviors {
			behSet[strings.ToUpper(b)] = true
		}
		filtered := make([]FindResultItem, 0)
		for _, item := range result {
			switch item.EntityType {
			case "control":
				ct, _ := item.Metadata["control_type"].(string)
				if behSet[ct] {
					filtered = append(filtered, item)
				}
			case "catalog_control":
				beh, _ := item.Metadata["behavior"].(string)
				if behSet[beh] {
					filtered = append(filtered, item)
				}
			}
			// Other entity types are excluded when behavior filter is active
		}
		result = filtered
	}

	if len(severities) > 0 {
		sevSet := make(map[string]bool, len(severities))
		for _, s := range severities {
			sevSet[strings.ToUpper(s)] = true
		}
		filtered := make([]FindResultItem, 0)
		for _, item := range result {
			if item.EntityType == "catalog_control" {
				sev, _ := item.Metadata["severity"].(string)
				if sevSet[sev] {
					filtered = append(filtered, item)
				}
			}
			// Only catalog_control has severity — other types excluded
		}
		result = filtered
	}

	return result
}

// ---------------------------------------------------------------------------
// Coverage mode
// ---------------------------------------------------------------------------

// Coverage computes posture coverage for a given entity.
func (e *Engine) Coverage(targetType, targetID string) (CoverageResult, error) {
	switch targetType {
	case "ou":
		return e.coverageOU(targetID)
	case "account":
		return e.coverageAccount(targetID)
	case "domain", "objective", "common_control":
		return e.coverageOntology(targetID, targetType)
	default:
		return CoverageResult{}, fmt.Errorf("invalid target_type: must be one of ou, account, domain, objective, common_control")
	}
}

func (e *Engine) coverageOU(ouID string) (CoverageResult, error) {
	ou, ok := e.store.GetOU(ouID)
	if !ok {
		return CoverageResult{}, fmt.Errorf("entity not found: %s", ouID)
	}

	controls := e.store.ControlsForTarget(ouID)
	scps := e.store.SCPsForTarget(ouID)
	accounts := e.store.AccountsForOU(ouID)

	items := make([]CoverageItem, 0, len(controls)+len(scps))
	for _, c := range controls {
		items = append(items, CoverageItem{
			ARN:        c.ARN,
			Name:       c.Name,
			EntityType: "control",
			IsCovered:  true,
			Metadata: map[string]any{
				"control_type": string(c.ControlType),
				"target_id":    c.TargetID,
			},
		})
	}
	for _, s := range scps {
		items = append(items, CoverageItem{
			ARN:        s.ARN,
			Name:       s.Name,
			EntityType: "scp",
			IsCovered:  true,
			Metadata:   map[string]any{"target_ids": s.TargetIDs},
		})
	}

	totalControls := len(controls) + len(scps)
	return CoverageResult{
		TargetID:      ouID,
		TargetType:    "ou",
		TargetName:    ou.Name,
		Items:         items,
		TotalControls: totalControls,
		EnabledCount:  totalControls,
		GapCount:      0,
		SCPCount:      len(scps),
		AccountCount:  len(accounts),
	}, nil
}

func (e *Engine) coverageAccount(accountID string) (CoverageResult, error) {
	account, ok := e.store.GetAccount(accountID)
	if !ok {
		return CoverageResult{}, fmt.Errorf("entity not found: %s", accountID)
	}

	items := make([]CoverageItem, 0)
	totalSCPs := 0

	// Walk OU ancestry to root, collecting controls and SCPs
	currentOUID := account.OUID
	for currentOUID != "" {
		ou, ok := e.store.GetOU(currentOUID)
		if !ok {
			break
		}
		for _, c := range e.store.ControlsForTarget(currentOUID) {
			items = append(items, CoverageItem{
				ARN:        c.ARN,
				Name:       c.Name,
				EntityType: "control",
				IsCovered:  true,
				Metadata: map[string]any{
					"control_type": string(c.ControlType),
					"target_id":    c.TargetID,
				},
			})
		}
		for _, s := range e.store.SCPsForTarget(currentOUID) {
			items = append(items, CoverageItem{
				ARN:        s.ARN,
				Name:       s.Name,
				EntityType: "scp",
				IsCovered:  true,
				Metadata:   map[string]any{"target_ids": s.TargetIDs},
			})
			totalSCPs++
		}
		currentOUID = ou.ParentID
	}

	controlItems := 0
	for _, item := range items {
		if item.EntityType == "control" {
			controlItems++
		}
	}

	return CoverageResult{
		TargetID:      accountID,
		TargetType:    "account",
		TargetName:    account.Name,
		Items:         items,
		TotalControls: controlItems + totalSCPs,
		EnabledCount:  controlItems + totalSCPs,
		GapCount:      0,
		SCPCount:      totalSCPs,
		AccountCount:  0,
	}, nil
}

func (e *Engine) coverageOntology(targetID, targetType string) (CoverageResult, error) {
	catalogControls := e.store.AllCatalogControls()
	enabledControls := e.store.AllControls()
	controlMappings := e.store.GetControlMappings()

	reverseMap := buildReverseMapping(catalogControls, controlMappings)
	enabledOUMap := buildEnabledOUMap(catalogControls, enabledControls)

	var implementing []models.CatalogControl
	var targetName string

	switch targetType {
	case "common_control":
		cc, ok := e.store.GetCatalogCommonControl(targetID)
		if !ok {
			return CoverageResult{}, fmt.Errorf("entity not found: %s", targetID)
		}
		targetName = cc.Name
		implementing = reverseMap[cc.ARN]

	case "objective":
		obj, ok := e.store.GetCatalogObjective(targetID)
		if !ok {
			return CoverageResult{}, fmt.Errorf("entity not found: %s", targetID)
		}
		targetName = obj.Name
		childCCs := e.store.CatalogCommonControlsForObjective(obj.ARN)
		implementing = make([]models.CatalogControl, 0)
		for _, cc := range childCCs {
			implementing = append(implementing, reverseMap[cc.ARN]...)
		}

	case "domain":
		domain, ok := e.store.GetCatalogDomain(targetID)
		if !ok {
			return CoverageResult{}, fmt.Errorf("entity not found: %s", targetID)
		}
		targetName = domain.Name
		childObjs := e.store.CatalogObjectivesForDomain(domain.ARN)
		implementing = make([]models.CatalogControl, 0)
		for _, obj := range childObjs {
			childCCs := e.store.CatalogCommonControlsForObjective(obj.ARN)
			for _, cc := range childCCs {
				implementing = append(implementing, reverseMap[cc.ARN]...)
			}
		}
	}

	if implementing == nil {
		implementing = make([]models.CatalogControl, 0)
	}

	items := make([]CoverageItem, 0, len(implementing))
	enabledCount := 0
	gapCount := 0
	for _, cc := range implementing {
		_, isCovered := enabledOUMap[cc.ARN]
		items = append(items, CoverageItem{
			ARN:        cc.ARN,
			Name:       cc.Name,
			EntityType: "catalog_control",
			IsCovered:  isCovered,
			Metadata: map[string]any{
				"behavior": string(cc.Behavior),
				"severity": string(cc.Severity),
			},
		})
		if isCovered {
			enabledCount++
		} else {
			gapCount++
		}
	}

	return CoverageResult{
		TargetID:      targetID,
		TargetType:    targetType,
		TargetName:    targetName,
		Items:         items,
		TotalControls: len(implementing),
		EnabledCount:  enabledCount,
		GapCount:      gapCount,
		SCPCount:      0,
		AccountCount:  0,
	}, nil
}

// ---------------------------------------------------------------------------
// Path mode
// ---------------------------------------------------------------------------

// Path computes the OU inheritance chain for an account from root to leaf.
func (e *Engine) Path(accountID string) (PathResult, error) {
	account, ok := e.store.GetAccount(accountID)
	if !ok {
		return PathResult{}, fmt.Errorf("entity not found: %s", accountID)
	}

	chain := make([]PathNode, 0)
	currentOUID := account.OUID

	for currentOUID != "" {
		ou, ok := e.store.GetOU(currentOUID)
		if !ok {
			break
		}

		controls := e.store.ControlsForTarget(currentOUID)
		scps := e.store.SCPsForTarget(currentOUID)

		ctrlItems := make([]FindResultItem, 0, len(controls))
		for _, c := range controls {
			ctrlItems = append(ctrlItems, FindResultItem{
				EntityType:  "control",
				ID:          c.ControlID,
				Name:        c.Name,
				ARN:         c.ARN,
				Description: c.Description,
				Metadata: map[string]any{
					"control_type": string(c.ControlType),
					"enforcement":  string(c.Enforcement),
					"target_id":    c.TargetID,
				},
			})
		}

		scpItems := make([]FindResultItem, 0, len(scps))
		for _, s := range scps {
			scpItems = append(scpItems, FindResultItem{
				EntityType:  "scp",
				ID:          s.ID,
				Name:        s.Name,
				ARN:         s.ARN,
				Description: s.Description,
				Metadata:    map[string]any{"target_ids": s.TargetIDs},
			})
		}

		chain = append(chain, PathNode{
			OUID:     ou.ID,
			OUName:   ou.Name,
			OUARN:    ou.ARN,
			Controls: ctrlItems,
			SCPs:     scpItems,
		})
		currentOUID = ou.ParentID
	}

	// Reverse so chain is root → leaf
	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}

	totalControls := 0
	totalSCPs := 0
	for _, n := range chain {
		totalControls += len(n.Controls)
		totalSCPs += len(n.SCPs)
	}

	return PathResult{
		AccountID:     accountID,
		AccountName:   account.Name,
		Chain:         chain,
		TotalControls: totalControls,
		TotalSCPs:     totalSCPs,
	}, nil
}

// ---------------------------------------------------------------------------
// Quick query mode
// ---------------------------------------------------------------------------

// QuickQuery executes a preset compliance query.
func (e *Engine) QuickQuery(preset string, page, pageSize int) (QuickQueryResult, error) {
	if !validPresets[preset] {
		return QuickQueryResult{}, fmt.Errorf(
			"unknown preset: %s. Valid presets: %s",
			preset, strings.Join(ValidPresetNames(), ", "),
		)
	}

	var allItems []FindResultItem

	switch preset {
	case "unenabled_critical_controls":
		allItems = e.unenabledCriticalControls()
	case "ous_without_detective_controls":
		allItems = e.ousWithoutDetectiveControls()
	case "accounts_without_scps":
		allItems = e.accountsWithoutSCPs()
	case "uncovered_ontology_objectives":
		allItems = e.uncoveredOntologyObjectives()
	}

	total := len(allItems)
	start := (page - 1) * pageSize
	end := start + pageSize
	if start > total {
		start = total
	}
	if end > total {
		end = total
	}
	pageItems := make([]FindResultItem, len(allItems[start:end]))
	copy(pageItems, allItems[start:end])

	return QuickQueryResult{
		Items: pageItems,
		Total: total,
	}, nil
}

// ---------------------------------------------------------------------------
// Quick query helpers
// ---------------------------------------------------------------------------

func (e *Engine) unenabledCriticalControls() []FindResultItem {
	catalogControls := e.store.AllCatalogControls()
	enabledOUMap := buildEnabledOUMap(catalogControls, e.store.AllControls())

	items := make([]FindResultItem, 0)
	for _, cc := range catalogControls {
		if cc.Severity != models.SeverityCritical {
			continue
		}
		if _, ok := enabledOUMap[cc.ARN]; ok {
			continue
		}
		items = append(items, FindResultItem{
			EntityType:  "catalog_control",
			ID:          cc.ARN,
			Name:        cc.Name,
			ARN:         cc.ARN,
			Description: cc.Description,
			Metadata: map[string]any{
				"behavior": string(cc.Behavior),
				"severity": string(cc.Severity),
			},
		})
	}
	return items
}

func (e *Engine) ousWithoutDetectiveControls() []FindResultItem {
	items := make([]FindResultItem, 0)
	for _, ou := range e.store.AllOUs() {
		controls := e.store.ControlsForTarget(ou.ID)
		hasDetective := false
		for _, c := range controls {
			if c.ControlType == models.ControlTypeDetective {
				hasDetective = true
				break
			}
		}
		if !hasDetective {
			items = append(items, FindResultItem{
				EntityType:  "ou",
				ID:          ou.ID,
				Name:        ou.Name,
				ARN:         ou.ARN,
				Description: "",
				Metadata:    map[string]any{"parent_id": ou.ParentID},
			})
		}
	}
	return items
}

func (e *Engine) accountsWithoutSCPs() []FindResultItem {
	items := make([]FindResultItem, 0)
	for _, acct := range e.store.AllAccounts() {
		hasSCP := false
		currentOUID := acct.OUID
		for currentOUID != "" {
			ou, ok := e.store.GetOU(currentOUID)
			if !ok {
				break
			}
			if len(e.store.SCPsForTarget(currentOUID)) > 0 {
				hasSCP = true
				break
			}
			currentOUID = ou.ParentID
		}
		if !hasSCP {
			items = append(items, FindResultItem{
				EntityType:  "account",
				ID:          acct.ID,
				Name:        acct.Name,
				ARN:         acct.ARN,
				Description: "",
				Metadata:    map[string]any{"email": acct.Email, "status": acct.Status, "ou_id": acct.OUID},
			})
		}
	}
	return items
}

func (e *Engine) uncoveredOntologyObjectives() []FindResultItem {
	catalogControls := e.store.AllCatalogControls()
	enabledControls := e.store.AllControls()
	controlMappings := e.store.GetControlMappings()

	reverseMap := buildReverseMapping(catalogControls, controlMappings)
	enabledOUMap := buildEnabledOUMap(catalogControls, enabledControls)

	items := make([]FindResultItem, 0)
	for _, obj := range e.store.AllCatalogObjectives() {
		childCCs := e.store.CatalogCommonControlsForObjective(obj.ARN)
		hasEnabled := false
		for _, cc := range childCCs {
			impl := reverseMap[cc.ARN]
			for _, catCtrl := range impl {
				if _, ok := enabledOUMap[catCtrl.ARN]; ok {
					hasEnabled = true
					break
				}
			}
			if hasEnabled {
				break
			}
		}
		if !hasEnabled {
			items = append(items, FindResultItem{
				EntityType:  "catalog_objective",
				ID:          obj.ARN,
				Name:        obj.Name,
				ARN:         obj.ARN,
				Description: obj.Description,
				Metadata:    map[string]any{"domain_arn": obj.DomainARN},
			})
		}
	}
	return items
}

// ---------------------------------------------------------------------------
// Posture helpers (ported from Python posture/aggregator.py)
// These are local to the search package to avoid circular imports.
// ---------------------------------------------------------------------------

// buildReverseMapping groups catalog controls by common_control_arn.
// Uses controlMappings when available, falls back to CatalogControl.CommonControlARN.
func buildReverseMapping(catalogControls []models.CatalogControl, controlMappings map[string][]string) map[string][]models.CatalogControl {
	mapping := make(map[string][]models.CatalogControl)

	if len(controlMappings) > 0 {
		arnToControl := make(map[string]models.CatalogControl, len(catalogControls))
		for _, c := range catalogControls {
			arnToControl[c.ARN] = c
		}
		for ctrlARN, ccARNs := range controlMappings {
			ctrl, ok := arnToControl[ctrlARN]
			if !ok {
				continue
			}
			for _, ccARN := range ccARNs {
				mapping[ccARN] = append(mapping[ccARN], ctrl)
			}
		}
	} else {
		for _, c := range catalogControls {
			if c.CommonControlARN != "" {
				mapping[c.CommonControlARN] = append(mapping[c.CommonControlARN], c)
			}
		}
	}

	return mapping
}

// buildEnabledOUMap returns a mapping of catalog control ARN → list of OU IDs where enabled.
func buildEnabledOUMap(catalogControls []models.CatalogControl, enabledControls []models.Control) map[string][]string {
	idToOUs := buildIdentifierToOUs(enabledControls)
	result := make(map[string][]string)
	for _, cc := range catalogControls {
		ous := matchCatalogControl(cc, idToOUs)
		if len(ous) > 0 {
			result[cc.ARN] = ous
		}
	}
	return result
}

// buildIdentifierToOUs builds a lookup: control identifier → list of OU IDs.
func buildIdentifierToOUs(enabledControls []models.Control) map[string][]string {
	mapping := make(map[string][]string)
	for _, ec := range enabledControls {
		parts := strings.Split(ec.ARN, "/")
		if len(parts) >= 2 {
			identifier := parts[len(parts)-1]
			mapping[identifier] = append(mapping[identifier], ec.TargetID)
		}
	}
	return mapping
}

// matchCatalogControl returns the sorted list of OU IDs where this catalog control is enabled.
func matchCatalogControl(cc models.CatalogControl, identifierToOUs map[string][]string) []string {
	ousSet := make(map[string]bool)

	// Match by catalog control identifier (last segment of ARN)
	parts := strings.Split(cc.ARN, "/")
	if len(parts) >= 2 {
		identifier := parts[len(parts)-1]
		for _, ouID := range identifierToOUs[identifier] {
			ousSet[ouID] = true
		}
	}

	// Match by aliases
	for _, alias := range cc.Aliases {
		aliasParts := strings.Split(alias, "/")
		aliasIdent := aliasParts[len(aliasParts)-1]
		for _, ouID := range identifierToOUs[aliasIdent] {
			ousSet[ouID] = true
		}
	}

	ous := make([]string, 0, len(ousSet))
	for ouID := range ousSet {
		ous = append(ous, ouID)
	}
	sort.Strings(ous)
	return ous
}

// pageCount computes the number of pages for a given total and page size.
func pageCount(total, pageSize int) int {
	if pageSize <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(pageSize)))
}

// PageCount returns the number of pages for a FindResult given a page size.
func (r FindResult) PageCount(pageSize int) int {
	return pageCount(r.Total, pageSize)
}

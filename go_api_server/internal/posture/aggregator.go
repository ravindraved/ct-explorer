// Package posture provides pure functions for computing ontology posture metrics.
// All results are computed on-the-fly from existing store data.
package posture

import (
	"fmt"
	"sort"
	"strings"

	"go_api_server/internal/models"
	"go_api_server/internal/store"
)

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

// PostureNode represents a node in the ontology posture tree.
type PostureNode struct {
	ARN                string            `json:"arn"`
	Name               string            `json:"name"`
	Description        string            `json:"description"`
	NodeType           string            `json:"node_type"`
	Number             string            `json:"number"`
	ControlCount       int               `json:"control_count"`
	BehaviorBreakdown  map[string]int    `json:"behavior_breakdown"`
	CoverageStatus     string            `json:"coverage_status"`
	EnabledCount       int               `json:"enabled_count"`
	TotalCount         int               `json:"total_count"`
	Children           []PostureNode     `json:"children"`
}

// NodeControl represents a single implementing control with enabled OUs.
type NodeControl struct {
	ARN        string   `json:"arn"`
	Name       string   `json:"name"`
	Behavior   string   `json:"behavior"`
	Severity   string   `json:"severity"`
	Aliases    []string `json:"aliases"`
	EnabledOUs []string `json:"enabled_ous"`
}

// NodeControlsResult holds the full response for a node's controls.
type NodeControlsResult struct {
	ARN               string         `json:"arn"`
	Name              string         `json:"name"`
	Description       string         `json:"description"`
	NodeType          string         `json:"node_type"`
	Number            string         `json:"number"`
	Breadcrumb        []string       `json:"breadcrumb"`
	Controls          []NodeControl  `json:"controls"`
	ControlCount      int            `json:"control_count"`
	EnabledCount      int            `json:"enabled_count"`
	DistinctOUCount   int            `json:"distinct_ou_count"`
	BehaviorBreakdown map[string]int `json:"behavior_breakdown"`
	CoverageStatus    string         `json:"coverage_status"`
}

// ---------------------------------------------------------------------------
// Public functions
// ---------------------------------------------------------------------------

// BuildReverseMapping groups catalog controls by common_control_arn.
// Uses controlMappings when available, falls back to CatalogControl.CommonControlARN.
func BuildReverseMapping(catalogControls []models.CatalogControl, controlMappings map[string][]string) map[string][]models.CatalogControl {
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

// BuildEnabledOUMap returns a mapping of catalog control ARN → list of OU IDs where enabled.
func BuildEnabledOUMap(catalogControls []models.CatalogControl, enabledControls []models.Control) map[string][]string {
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

// ComputeCoverageStatus returns "green"/"yellow"/"red" based on control behaviors.
func ComputeCoverageStatus(enabledCount, totalCount int) string {
	if totalCount == 0 {
		return "red"
	}
	if enabledCount == 0 {
		return "red"
	}
	if enabledCount == totalCount {
		return "green"
	}
	return "yellow"
}

// computeCoverageStatusFromControls returns coverage status based on control behaviors.
func computeCoverageStatusFromControls(controls []models.CatalogControl) string {
	if len(controls) == 0 {
		return "red"
	}
	behaviors := make(map[string]bool)
	for _, c := range controls {
		behaviors[string(c.Behavior)] = true
	}
	if behaviors["PREVENTIVE"] {
		return "green"
	}
	return "yellow"
}

// ComputeBehaviorBreakdown counts controls by behavior type.
func ComputeBehaviorBreakdown(controls []models.CatalogControl) map[string]int {
	breakdown := map[string]int{
		"PREVENTIVE": 0,
		"DETECTIVE":  0,
		"PROACTIVE":  0,
	}
	for _, c := range controls {
		key := string(c.Behavior)
		breakdown[key]++
	}
	return breakdown
}

// BuildPostureTree builds the full ontology hierarchy with aggregated metrics.
func BuildPostureTree(s *store.Store) []PostureNode {
	catalogControls := s.AllCatalogControls()
	enabledControls := s.AllControls()
	domains := s.AllCatalogDomains()
	objectives := s.AllCatalogObjectives()
	commonControls := s.AllCatalogCommonControls()
	controlMappings := s.GetControlMappings()

	reverseMap := BuildReverseMapping(catalogControls, controlMappings)
	enabledOUMap := BuildEnabledOUMap(catalogControls, enabledControls)

	// Index objectives by domain, common controls by objective
	objsByDomain := make(map[string][]models.CatalogObjective)
	for _, o := range objectives {
		objsByDomain[o.DomainARN] = append(objsByDomain[o.DomainARN], o)
	}

	ccsByObjective := make(map[string][]models.CatalogCommonControl)
	for _, cc := range commonControls {
		ccsByObjective[cc.ObjectiveARN] = append(ccsByObjective[cc.ObjectiveARN], cc)
	}

	// Sort domains by name
	sort.Slice(domains, func(i, j int) bool { return domains[i].Name < domains[j].Name })

	tree := make([]PostureNode, 0, len(domains))
	for di, domain := range domains {
		dNum := fmt.Sprintf("%d", di+1)
		domainControls := make([]models.CatalogControl, 0)
		domainEnabled := 0

		domainObjs := objsByDomain[domain.ARN]
		sort.Slice(domainObjs, func(i, j int) bool { return domainObjs[i].Name < domainObjs[j].Name })

		objNodes := make([]PostureNode, 0, len(domainObjs))
		for oi, obj := range domainObjs {
			oNum := fmt.Sprintf("%s.%d", dNum, oi+1)
			objControls := make([]models.CatalogControl, 0)
			objEnabled := 0

			objCCs := ccsByObjective[obj.ARN]
			sort.Slice(objCCs, func(i, j int) bool { return objCCs[i].Name < objCCs[j].Name })

			ccNodes := make([]PostureNode, 0, len(objCCs))
			for ci, cc := range objCCs {
				impl := reverseMap[cc.ARN]
				if impl == nil {
					impl = make([]models.CatalogControl, 0)
				}
				enCount := 0
				for _, c := range impl {
					if _, ok := enabledOUMap[c.ARN]; ok {
						enCount++
					}
				}
				ccNodes = append(ccNodes, PostureNode{
					ARN:               cc.ARN,
					Name:              cc.Name,
					Description:       cc.Description,
					NodeType:          "common_control",
					Number:            fmt.Sprintf("%s.%d", oNum, ci+1),
					ControlCount:      len(impl),
					BehaviorBreakdown: ComputeBehaviorBreakdown(impl),
					CoverageStatus:    computeCoverageStatusFromControls(impl),
					EnabledCount:      enCount,
					TotalCount:        len(impl),
					Children:          make([]PostureNode, 0),
				})
				objControls = append(objControls, impl...)
				objEnabled += enCount
			}

			objNodes = append(objNodes, PostureNode{
				ARN:               obj.ARN,
				Name:              obj.Name,
				Description:       obj.Description,
				NodeType:          "objective",
				Number:            oNum,
				ControlCount:      len(objControls),
				BehaviorBreakdown: ComputeBehaviorBreakdown(objControls),
				CoverageStatus:    computeCoverageStatusFromControls(objControls),
				EnabledCount:      objEnabled,
				TotalCount:        len(objControls),
				Children:          ccNodes,
			})
			domainControls = append(domainControls, objControls...)
			domainEnabled += objEnabled
		}

		tree = append(tree, PostureNode{
			ARN:               domain.ARN,
			Name:              domain.Name,
			Description:       domain.Description,
			NodeType:          "domain",
			Number:            dNum,
			ControlCount:      len(domainControls),
			BehaviorBreakdown: ComputeBehaviorBreakdown(domainControls),
			CoverageStatus:    computeCoverageStatusFromControls(domainControls),
			EnabledCount:      domainEnabled,
			TotalCount:        len(domainControls),
			Children:          objNodes,
		})
	}

	return tree
}

// GetNodeControls returns implementing controls for a node with metadata.
// Returns nil if the ARN is not found.
func GetNodeControls(s *store.Store, arn string) (*NodeControlsResult, error) {
	catalogControls := s.AllCatalogControls()
	enabledControls := s.AllControls()
	controlMappings := s.GetControlMappings()

	enabledOUMap := BuildEnabledOUMap(catalogControls, enabledControls)
	reverseMap := BuildReverseMapping(catalogControls, controlMappings)

	controlToDict := func(c models.CatalogControl) NodeControl {
		aliases := c.Aliases
		if aliases == nil {
			aliases = make([]string, 0)
		}
		enabledOUs := enabledOUMap[c.ARN]
		if enabledOUs == nil {
			enabledOUs = make([]string, 0)
		}
		return NodeControl{
			ARN:        c.ARN,
			Name:       c.Name,
			Behavior:   string(c.Behavior),
			Severity:   string(c.Severity),
			Aliases:    aliases,
			EnabledOUs: enabledOUs,
		}
	}

	var nodeType string
	var impl []models.CatalogControl
	var name, description string
	var breadcrumb []string
	var number string

	// Check common control
	if cc, ok := s.GetCatalogCommonControl(arn); ok {
		nodeType = "common_control"
		impl = reverseMap[cc.ARN]
		name = cc.Name
		description = cc.Description

		obj, objOK := s.GetCatalogObjective(cc.ObjectiveARN)
		if objOK {
			dom, domOK := s.GetCatalogDomain(obj.DomainARN)
			if domOK {
				breadcrumb = []string{dom.Name, obj.Name, cc.Name}
				number = computeNumber(s, dom.ARN, obj.ARN, cc.ARN)
			} else {
				breadcrumb = []string{"Unknown", obj.Name, cc.Name}
			}
		} else {
			breadcrumb = []string{"Unknown", "Unknown", cc.Name}
		}
	} else if obj, ok := s.GetCatalogObjective(arn); ok {
		nodeType = "objective"
		name = obj.Name
		description = obj.Description
		childCCs := s.CatalogCommonControlsForObjective(obj.ARN)
		impl = make([]models.CatalogControl, 0)
		for _, cc := range childCCs {
			impl = append(impl, reverseMap[cc.ARN]...)
		}
		dom, domOK := s.GetCatalogDomain(obj.DomainARN)
		if domOK {
			breadcrumb = []string{dom.Name, obj.Name}
			number = computeNumberObjective(s, dom.ARN, obj.ARN)
		} else {
			breadcrumb = []string{"Unknown", obj.Name}
		}
	} else if dom, ok := s.GetCatalogDomain(arn); ok {
		nodeType = "domain"
		name = dom.Name
		description = dom.Description
		childObjs := s.CatalogObjectivesForDomain(dom.ARN)
		impl = make([]models.CatalogControl, 0)
		for _, obj := range childObjs {
			childCCs := s.CatalogCommonControlsForObjective(obj.ARN)
			for _, cc := range childCCs {
				impl = append(impl, reverseMap[cc.ARN]...)
			}
		}
		breadcrumb = []string{dom.Name}
		number = computeNumberDomain(s, dom.ARN)
	} else {
		return nil, fmt.Errorf("ontology node not found")
	}

	if impl == nil {
		impl = make([]models.CatalogControl, 0)
	}

	controls := make([]NodeControl, 0, len(impl))
	for _, c := range impl {
		controls = append(controls, controlToDict(c))
	}

	enabledCount := 0
	distinctOUs := make(map[string]struct{})
	for _, c := range controls {
		if len(c.EnabledOUs) > 0 {
			enabledCount++
		}
		for _, ou := range c.EnabledOUs {
			distinctOUs[ou] = struct{}{}
		}
	}

	// Build dummy CatalogControl list for behavior breakdown / coverage
	dummyControls := make([]models.CatalogControl, 0, len(controls))
	for _, c := range controls {
		dummyControls = append(dummyControls, models.CatalogControl{
			ARN:      c.ARN,
			Name:     c.Name,
			Behavior: models.CatalogBehavior(c.Behavior),
			Severity: models.CatalogSeverity(c.Severity),
		})
	}

	return &NodeControlsResult{
		ARN:               arn,
		Name:              name,
		Description:       description,
		NodeType:          nodeType,
		Number:            number,
		Breadcrumb:        breadcrumb,
		Controls:          controls,
		ControlCount:      len(controls),
		EnabledCount:      enabledCount,
		DistinctOUCount:   len(distinctOUs),
		BehaviorBreakdown: ComputeBehaviorBreakdown(dummyControls),
		CoverageStatus:    computeCoverageStatusFromControls(dummyControls),
	}, nil
}

// ---------------------------------------------------------------------------
// Internal helpers
// ---------------------------------------------------------------------------

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

func matchCatalogControl(cc models.CatalogControl, identifierToOUs map[string][]string) []string {
	ousSet := make(map[string]bool)

	parts := strings.Split(cc.ARN, "/")
	if len(parts) >= 2 {
		identifier := parts[len(parts)-1]
		for _, ouID := range identifierToOUs[identifier] {
			ousSet[ouID] = true
		}
	}

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

func computeNumber(s *store.Store, domainARN, objectiveARN, ccARN string) string {
	oi := computeNumberObjective(s, domainARN, objectiveARN)
	// CC index within objective
	ccs := s.CatalogCommonControlsForObjective(objectiveARN)
	sort.Slice(ccs, func(i, j int) bool { return ccs[i].Name < ccs[j].Name })
	ci := 0
	for i, cc := range ccs {
		if cc.ARN == ccARN {
			ci = i + 1
			break
		}
	}
	return fmt.Sprintf("%s.%d", oi, ci)
}

func computeNumberDomain(s *store.Store, domainARN string) string {
	domains := s.AllCatalogDomains()
	sort.Slice(domains, func(i, j int) bool { return domains[i].Name < domains[j].Name })
	for i, d := range domains {
		if d.ARN == domainARN {
			return fmt.Sprintf("%d", i+1)
		}
	}
	return "0"
}

func computeNumberObjective(s *store.Store, domainARN, objectiveARN string) string {
	di := computeNumberDomain(s, domainARN)
	objs := s.CatalogObjectivesForDomain(domainARN)
	sort.Slice(objs, func(i, j int) bool { return objs[i].Name < objs[j].Name })
	for i, o := range objs {
		if o.ARN == objectiveARN {
			return fmt.Sprintf("%s.%d", di, i+1)
		}
	}
	return fmt.Sprintf("%s.0", di)
}

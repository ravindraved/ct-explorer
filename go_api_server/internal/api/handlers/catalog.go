package handlers

import (
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"

	"go_api_server/internal/api"
	"go_api_server/internal/models"
	"go_api_server/internal/store"
)

// catalogControlResponse is the JSON shape for a catalog control.
type catalogControlResponse struct {
	ARN                      string   `json:"arn"`
	Name                     string   `json:"name"`
	Description              string   `json:"description"`
	Behavior                 string   `json:"behavior"`
	Severity                 string   `json:"severity"`
	Aliases                  []string `json:"aliases"`
	GovernedResources        []string `json:"governed_resources"`
	Services                 []string `json:"services"`
	ImplementationType       string   `json:"implementation_type"`
	ImplementationIdentifier string   `json:"implementation_identifier"`
	CreateTime               string   `json:"create_time"`
	CommonControlARN         string   `json:"common_control_arn"`
}

// catalogServiceResponse is the JSON shape for a service entry.
type catalogServiceResponse struct {
	Service string `json:"service"`
	Count   int    `json:"count"`
}

func toCatalogControlResponse(c models.CatalogControl) catalogControlResponse {
	aliases := c.Aliases
	if aliases == nil {
		aliases = make([]string, 0)
	}
	governed := c.GovernedResources
	if governed == nil {
		governed = make([]string, 0)
	}
	return catalogControlResponse{
		ARN:                      c.ARN,
		Name:                     c.Name,
		Description:              c.Description,
		Behavior:                 string(c.Behavior),
		Severity:                 string(c.Severity),
		Aliases:                  aliases,
		GovernedResources:        governed,
		Services:                 extractServices(governed),
		ImplementationType:       c.ImplementationType,
		ImplementationIdentifier: c.ImplementationIdentifier,
		CreateTime:               c.CreateTime,
		CommonControlARN:         c.CommonControlARN,
	}
}

// extractServices derives unique AWS service names from CloudFormation resource types.
// "AWS::S3::Bucket" → "S3". Returns ["Cross-Service"] if empty.
func extractServices(governedResources []string) []string {
	seen := make(map[string]struct{})
	for _, r := range governedResources {
		parts := strings.Split(r, "::")
		if len(parts) >= 2 {
			seen[parts[1]] = struct{}{}
		}
	}
	if len(seen) == 0 {
		return []string{"Cross-Service"}
	}
	services := make([]string, 0, len(seen))
	for s := range seen {
		services = append(services, s)
	}
	sort.Strings(services)
	return services
}

// CatalogDomains godoc
// @Summary      List catalog domains
// @Description  Returns all catalog domains
// @Tags         catalog
// @Produce      json
// @Success      200  {array}  models.CatalogDomain
// @Router       /api/catalog/domains [get]
func CatalogDomains(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		domains := s.AllCatalogDomains()
		if domains == nil {
			domains = make([]models.CatalogDomain, 0)
		}
		api.WriteJSON(w, http.StatusOK, domains)
	}
}

// CatalogObjectives godoc
// @Summary      List catalog objectives
// @Description  Returns all catalog objectives, optionally filtered by domain_arn
// @Tags         catalog
// @Produce      json
// @Param        domain_arn  query  string  false  "Domain ARN filter"
// @Success      200  {array}  models.CatalogObjective
// @Router       /api/catalog/objectives [get]
func CatalogObjectives(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		domainARN := r.URL.Query().Get("domain_arn")
		var objectives []models.CatalogObjective
		if domainARN != "" {
			objectives = s.CatalogObjectivesForDomain(domainARN)
		} else {
			objectives = s.AllCatalogObjectives()
		}
		if objectives == nil {
			objectives = make([]models.CatalogObjective, 0)
		}
		api.WriteJSON(w, http.StatusOK, objectives)
	}
}

// CatalogCommonControls godoc
// @Summary      List catalog common controls
// @Description  Returns all catalog common controls, optionally filtered by objective_arn
// @Tags         catalog
// @Produce      json
// @Param        objective_arn  query  string  false  "Objective ARN filter"
// @Success      200  {array}  models.CatalogCommonControl
// @Router       /api/catalog/common-controls [get]
func CatalogCommonControls(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		objectiveARN := r.URL.Query().Get("objective_arn")
		var ccs []models.CatalogCommonControl
		if objectiveARN != "" {
			ccs = s.CatalogCommonControlsForObjective(objectiveARN)
		} else {
			ccs = s.AllCatalogCommonControls()
		}
		if ccs == nil {
			ccs = make([]models.CatalogCommonControl, 0)
		}
		api.WriteJSON(w, http.StatusOK, ccs)
	}
}

// CatalogControls godoc
// @Summary      List catalog controls
// @Description  Returns catalog controls with optional filters (AND logic)
// @Tags         catalog
// @Produce      json
// @Param        behavior            query  string  false  "Behavior filter"
// @Param        severity            query  string  false  "Severity filter"
// @Param        implementation_type query  string  false  "Implementation type filter"
// @Param        service             query  string  false  "Service filter"
// @Param        domain_arn          query  string  false  "Domain ARN filter"
// @Param        objective_arn       query  string  false  "Objective ARN filter"
// @Param        q                   query  string  false  "Search text"
// @Success      200  {array}  catalogControlResponse
// @Router       /api/catalog/controls [get]
func CatalogControls(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		behavior := q.Get("behavior")
		severity := q.Get("severity")
		implType := q.Get("implementation_type")
		service := q.Get("service")
		domainARN := q.Get("domain_arn")
		objectiveARN := q.Get("objective_arn")
		search := q.Get("q")

		controls := s.AllCatalogControls()

		if behavior != "" {
			filtered := make([]models.CatalogControl, 0)
			for _, c := range controls {
				if string(c.Behavior) == behavior {
					filtered = append(filtered, c)
				}
			}
			controls = filtered
		}

		if severity != "" {
			filtered := make([]models.CatalogControl, 0)
			for _, c := range controls {
				if string(c.Severity) == severity {
					filtered = append(filtered, c)
				}
			}
			controls = filtered
		}

		if implType != "" {
			filtered := make([]models.CatalogControl, 0)
			for _, c := range controls {
				if c.ImplementationType == implType {
					filtered = append(filtered, c)
				}
			}
			controls = filtered
		}

		if service != "" {
			filtered := make([]models.CatalogControl, 0)
			for _, c := range controls {
				if service == "Cross-Service" {
					if len(c.GovernedResources) == 0 {
						filtered = append(filtered, c)
					}
				} else {
					for _, res := range c.GovernedResources {
						parts := strings.Split(res, "::")
						if len(parts) >= 2 && parts[1] == service {
							filtered = append(filtered, c)
							break
						}
					}
				}
			}
			controls = filtered
		}

		if domainARN != "" {
			objectiveARNs := make(map[string]struct{})
			for _, o := range s.CatalogObjectivesForDomain(domainARN) {
				objectiveARNs[o.ARN] = struct{}{}
			}
			ccARNs := make(map[string]struct{})
			for oa := range objectiveARNs {
				for _, cc := range s.CatalogCommonControlsForObjective(oa) {
					ccARNs[cc.ARN] = struct{}{}
				}
			}
			filtered := make([]models.CatalogControl, 0)
			for _, c := range controls {
				if _, ok := ccARNs[c.CommonControlARN]; ok {
					filtered = append(filtered, c)
				}
			}
			controls = filtered
		}

		if objectiveARN != "" {
			ccARNs := make(map[string]struct{})
			for _, cc := range s.CatalogCommonControlsForObjective(objectiveARN) {
				ccARNs[cc.ARN] = struct{}{}
			}
			filtered := make([]models.CatalogControl, 0)
			for _, c := range controls {
				if _, ok := ccARNs[c.CommonControlARN]; ok {
					filtered = append(filtered, c)
				}
			}
			controls = filtered
		}

		if search != "" {
			term := strings.ToLower(search)
			filtered := make([]models.CatalogControl, 0)
			for _, c := range controls {
				if strings.Contains(strings.ToLower(c.Name), term) ||
					strings.Contains(strings.ToLower(c.Description), term) {
					filtered = append(filtered, c)
					continue
				}
				for _, alias := range c.Aliases {
					if strings.Contains(strings.ToLower(alias), term) {
						filtered = append(filtered, c)
						break
					}
				}
			}
			controls = filtered
		}

		result := make([]catalogControlResponse, 0, len(controls))
		for _, c := range controls {
			result = append(result, toCatalogControlResponse(c))
		}
		api.WriteJSON(w, http.StatusOK, result)
	}
}

// CatalogControlByARN godoc
// @Summary      Get catalog control by ARN
// @Description  Returns a single catalog control by ARN. 404 if not found.
// @Tags         catalog
// @Produce      json
// @Param        arn  path  string  true  "Catalog control ARN"
// @Success      200  {object}  catalogControlResponse
// @Failure      404  {object}  responses.ErrorResponse
// @Router       /api/catalog/controls/{arn} [get]
func CatalogControlByARN(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		arn := chi.URLParam(r, "*")

		cc, ok := s.GetCatalogControl(arn)
		if !ok {
			api.WriteError(w, http.StatusNotFound, "Catalog control not found")
			return
		}
		api.WriteJSON(w, http.StatusOK, toCatalogControlResponse(cc))
	}
}

// CatalogServices godoc
// @Summary      List catalog services
// @Description  Returns distinct service names with control counts from governed_resources
// @Tags         catalog
// @Produce      json
// @Success      200  {array}  catalogServiceResponse
// @Router       /api/catalog/services [get]
func CatalogServices(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		controls := s.AllCatalogControls()
		serviceCounts := make(map[string]int)
		for _, c := range controls {
			for _, svc := range extractServices(c.GovernedResources) {
				serviceCounts[svc]++
			}
		}

		// Sort by service name
		names := make([]string, 0, len(serviceCounts))
		for name := range serviceCounts {
			names = append(names, name)
		}
		sort.Strings(names)

		result := make([]catalogServiceResponse, 0, len(names))
		for _, name := range names {
			result = append(result, catalogServiceResponse{
				Service: name,
				Count:   serviceCounts[name],
			})
		}
		api.WriteJSON(w, http.StatusOK, result)
	}
}

// CatalogEnabledMap godoc
// @Summary      Get enabled map
// @Description  Returns catalog control ARN → OU IDs mapping using identifier-based matching
// @Tags         catalog
// @Produce      json
// @Success      200  {object}  map[string][]string
// @Router       /api/catalog/enabled-map [get]
func CatalogEnabledMap(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		enabledControls := s.AllControls()
		catalogControls := s.AllCatalogControls()

		// Build lookup: identifier → list of OU IDs from enabled controls
		identifierToOUs := make(map[string][]string)
		for _, ec := range enabledControls {
			parts := strings.SplitN(ec.ARN, "/", -1)
			if len(parts) >= 2 {
				ident := parts[len(parts)-1]
				identifierToOUs[ident] = append(identifierToOUs[ident], ec.TargetID)
			}
		}

		// Match catalog controls by identifier or aliases
		result := make(map[string][]string)
		for _, cc := range catalogControls {
			ous := make(map[string]struct{})

			// Match by catalog control identifier (last segment of ARN)
			ccParts := strings.SplitN(cc.ARN, "/", -1)
			if len(ccParts) >= 2 {
				ccIdent := ccParts[len(ccParts)-1]
				if targets, ok := identifierToOUs[ccIdent]; ok {
					for _, t := range targets {
						ous[t] = struct{}{}
					}
				}
			}

			// Match by aliases
			for _, alias := range cc.Aliases {
				aliasParts := strings.SplitN(alias, "/", -1)
				aliasIdent := alias
				if len(aliasParts) >= 2 {
					aliasIdent = aliasParts[len(aliasParts)-1]
				}
				if targets, ok := identifierToOUs[aliasIdent]; ok {
					for _, t := range targets {
						ous[t] = struct{}{}
					}
				}
			}

			if len(ous) > 0 {
				sorted := make([]string, 0, len(ous))
				for ou := range ous {
					sorted = append(sorted, ou)
				}
				sort.Strings(sorted)
				result[cc.ARN] = sorted
			}
		}

		api.WriteJSON(w, http.StatusOK, result)
	}
}

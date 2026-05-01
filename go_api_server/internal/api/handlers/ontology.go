package handlers

import (
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"

	"go_api_server/internal/api"
	"go_api_server/internal/posture"
	"go_api_server/internal/store"
)

// PostureTree godoc
// @Summary      Get ontology posture tree
// @Description  Returns full ontology hierarchy with aggregated posture metrics
// @Tags         ontology
// @Produce      json
// @Success      200  {array}  posture.PostureNode
// @Router       /api/ontology/posture-tree [get]
func PostureTree(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tree := posture.BuildPostureTree(s)
		if tree == nil {
			tree = make([]posture.PostureNode, 0)
		}
		api.WriteJSON(w, http.StatusOK, tree)
	}
}

// NodeControls godoc
// @Summary      Get node controls
// @Description  Returns implementing controls for a specific ontology node
// @Tags         ontology
// @Produce      json
// @Param        arn  path  string  true  "Ontology node ARN"
// @Success      200  {object}  posture.NodeControlsResult
// @Failure      404  {object}  responses.ErrorResponse
// @Router       /api/ontology/node/{arn}/controls [get]
func NodeControls(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Wildcard captures everything after /node/ — strip trailing /controls
		arn := chi.URLParam(r, "*")
		arn = strings.TrimSuffix(arn, "/controls")

		result, err := posture.GetNodeControls(s, arn)
		if err != nil {
			api.WriteError(w, http.StatusNotFound, "Ontology node not found")
			return
		}
		api.WriteJSON(w, http.StatusOK, result)
	}
}

// controlMapRef is a single ontology reference for the control map.
type controlMapRef struct {
	ARN    string `json:"arn"`
	Number string `json:"number"`
	Name   string `json:"name"`
}

// ControlMap godoc
// @Summary      Get control ontology map
// @Description  Returns mapping of catalog control ARN to ontology references
// @Tags         ontology
// @Produce      json
// @Success      200  {object}  map[string][]object
// @Router       /api/ontology/control-map [get]
func ControlMap(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		controlMappings := s.GetControlMappings()
		if len(controlMappings) == 0 {
			api.WriteJSON(w, http.StatusOK, make(map[string][]controlMapRef))
			return
		}

		// Build number lookup from posture tree (common controls only)
		tree := posture.BuildPostureTree(s)
		ccARNToRef := make(map[string]controlMapRef)
		for _, domain := range tree {
			for _, obj := range domain.Children {
				for _, cc := range obj.Children {
					ccARNToRef[cc.ARN] = controlMapRef{
						ARN:    cc.ARN,
						Number: cc.Number,
						Name:   cc.Name,
					}
				}
			}
		}

		// Reverse: catalog_control_arn → list of {arn, number, name}
		result := make(map[string][]controlMapRef)
		for ctrlARN, ccARNs := range controlMappings {
			refs := make([]controlMapRef, 0)
			for _, ccARN := range ccARNs {
				if ref, ok := ccARNToRef[ccARN]; ok {
					refs = append(refs, ref)
				}
			}
			if len(refs) > 0 {
				sort.Slice(refs, func(i, j int) bool { return refs[i].Number < refs[j].Number })
				result[ctrlARN] = refs
			}
		}

		api.WriteJSON(w, http.StatusOK, result)
	}
}

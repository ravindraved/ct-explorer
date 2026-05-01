package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	"go_api_server/internal/api"
	"go_api_server/internal/api/responses"
	"go_api_server/internal/store"
)

// ListSCPs godoc
// @Summary      List SCPs
// @Description  Returns all SCPs in summary format
// @Tags         scps
// @Produce      json
// @Success      200  {array}  responses.SCPSummaryResponse
// @Router       /api/scps [get]
func ListSCPs(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scps := s.AllSCPs()
		result := make([]responses.SCPSummaryResponse, 0, len(scps))
		for _, scp := range scps {
			resp := responses.NewSCPSummaryResponse()
			resp.ID = scp.ID
			resp.ARN = scp.ARN
			resp.Name = scp.Name
			resp.Description = scp.Description
			if scp.TargetIDs != nil {
				resp.TargetIDs = scp.TargetIDs
			}
			result = append(result, resp)
		}
		api.WriteJSON(w, http.StatusOK, result)
	}
}

// GetSCP godoc
// @Summary      Get SCP detail
// @Description  Returns full SCP detail with policy document and parsed policy. 404 if not found.
// @Tags         scps
// @Produce      json
// @Param        scpID  path  string  true  "SCP ID"
// @Success      200  {object}  responses.SCPDetailResponse
// @Failure      404  {object}  responses.ErrorResponse
// @Router       /api/scps/{scpID} [get]
func GetSCP(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		scpID := chi.URLParam(r, "scpID")

		scp, ok := s.GetSCP(scpID)
		if !ok {
			api.WriteError(w, http.StatusNotFound, "SCP not found")
			return
		}

		resp := responses.NewSCPDetailResponse()
		resp.ID = scp.ID
		resp.ARN = scp.ARN
		resp.Name = scp.Name
		resp.Description = scp.Description
		if scp.TargetIDs != nil {
			resp.TargetIDs = scp.TargetIDs
		}
		resp.PolicyDocument = scp.PolicyDocument

		if scp.Policy != nil {
			policy := responses.NewPolicyResponse()
			policy.Version = scp.Policy.Version
			if scp.Policy.Statement != nil {
				policy.Statement = scp.Policy.Statement
			}
			resp.Policy = &policy
		}

		// Also try to parse policy_document into policy if Policy is nil
		if resp.Policy == nil && scp.PolicyDocument != "" {
			var raw map[string]json.RawMessage
			if err := json.Unmarshal([]byte(scp.PolicyDocument), &raw); err == nil {
				policy := responses.NewPolicyResponse()
				if v, ok := raw["Version"]; ok {
					json.Unmarshal(v, &policy.Version)
				}
				if s, ok := raw["Statement"]; ok {
					json.Unmarshal(s, &policy.Statement)
				}
				resp.Policy = &policy
			}
		}

		api.WriteJSON(w, http.StatusOK, resp)
	}
}

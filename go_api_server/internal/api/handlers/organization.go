package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"go_api_server/internal/api"
	"go_api_server/internal/api/responses"
	"go_api_server/internal/store"
)

// buildOrgTree recursively builds an OrgTreeNode for the given OU ID.
func buildOrgTree(s *store.Store, ouID string, resolver *api.AccountIDResolver) responses.OrgTreeNode {
	node := responses.NewOrgTreeNode()

	ou, ok := s.GetOU(ouID)
	if !ok {
		node.ID = ouID
		node.Type = "ou"
		return node
	}

	node.ID = ou.ID
	node.Name = ou.Name
	node.Type = "ou"
	node.ARN = ou.ARN

	for _, childID := range ou.ChildrenIDs {
		childNode := buildOrgTree(s, childID, resolver)
		node.Children = append(node.Children, childNode)
	}

	for _, acctID := range ou.AccountIDs {
		acct, ok := s.GetAccount(acctID)
		if !ok {
			continue
		}
		maskedID := resolver.Register(acct.ID)
		leaf := responses.NewOrgTreeNode()
		leaf.ID = maskedID
		leaf.Name = acct.Name
		leaf.Type = "account"
		leaf.ARN = api.MaskARN(acct.ARN)
		leaf.Email = &acct.Email
		leaf.Status = &acct.Status
		node.Children = append(node.Children, leaf)
	}

	return node
}

// OrgTree godoc
// @Summary      Get organization tree
// @Description  Returns the full organization tree rooted at root_ou_id
// @Tags         organization
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /api/organization/tree [get]
func OrgTree(s *store.Store, resolver *api.AccountIDResolver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rootID := s.GetRootOUID()
		if rootID == "" {
			api.WriteJSON(w, http.StatusOK, map[string]interface{}{
				"tree":    nil,
				"message": "No organization data available",
			})
			return
		}

		_, ok := s.GetOU(rootID)
		if !ok {
			api.WriteJSON(w, http.StatusOK, map[string]interface{}{
				"tree":    nil,
				"message": "No organization data available",
			})
			return
		}

		tree := buildOrgTree(s, rootID, resolver)
		api.WriteJSON(w, http.StatusOK, map[string]interface{}{
			"tree":    tree,
			"message": "OK",
		})
	}
}

// AccountDetail godoc
// @Summary      Get account details
// @Description  Returns account info, inherited controls, and SCPs
// @Tags         organization
// @Produce      json
// @Param        accountID  path  string  true  "Account ID"
// @Success      200  {object}  responses.AccountDetailResponse
// @Failure      404  {object}  responses.ErrorResponse
// @Router       /api/organization/account/{accountID} [get]
func AccountDetail(s *store.Store, resolver *api.AccountIDResolver) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inputID := chi.URLParam(r, "accountID")
		// Resolve masked ID to real ID for store lookup
		accountID := resolver.Resolve(inputID)

		account, ok := s.GetAccount(accountID)
		if !ok {
			api.WriteError(w, http.StatusNotFound, "Account not found")
			return
		}

		maskedID := resolver.Register(account.ID)

		ouName := "Unknown"
		if parentOU, ok := s.GetOU(account.OUID); ok {
			ouName = parentOU.Name
		}

		// Controls: inherited from parent OU + direct account-level
		inheritedControls := s.ControlsForTarget(account.OUID)
		directControls := s.ControlsForTarget(accountID)
		seenControls := make(map[string]struct{})
		controlResponses := make([]responses.ControlResponse, 0)
		for _, c := range append(inheritedControls, directControls...) {
			key := c.ARN + "::" + c.TargetID
			if _, exists := seenControls[key]; exists {
				continue
			}
			seenControls[key] = struct{}{}
			controlResponses = append(controlResponses, responses.ControlResponse{
				ARN:         c.ARN,
				ControlID:   c.ControlID,
				Name:        c.Name,
				ControlType: string(c.ControlType),
				Enforcement: string(c.Enforcement),
				TargetID:    c.TargetID,
				Description: c.Description,
			})
		}

		// SCPs targeting this account directly or its parent OU
		scpsDirect := s.SCPsForTarget(accountID)
		seenSCPs := make(map[string]struct{})
		scpResponses := make([]responses.SCPSummaryResponse, 0)
		for _, scp := range scpsDirect {
			if _, exists := seenSCPs[scp.ID]; exists {
				continue
			}
			seenSCPs[scp.ID] = struct{}{}
			resp := responses.NewSCPSummaryResponse()
			resp.ID = scp.ID
			resp.ARN = scp.ARN
			resp.Name = scp.Name
			resp.Description = scp.Description
			if scp.TargetIDs != nil {
				resp.TargetIDs = scp.TargetIDs
			}
			scpResponses = append(scpResponses, resp)
		}
		if account.OUID != "" {
			for _, scp := range s.SCPsForTarget(account.OUID) {
				if _, exists := seenSCPs[scp.ID]; exists {
					continue
				}
				seenSCPs[scp.ID] = struct{}{}
				resp := responses.NewSCPSummaryResponse()
				resp.ID = scp.ID
				resp.ARN = scp.ARN
				resp.Name = scp.Name
				resp.Description = scp.Description
				if scp.TargetIDs != nil {
					resp.TargetIDs = scp.TargetIDs
				}
				scpResponses = append(scpResponses, resp)
			}
		}

		result := responses.AccountDetailResponse{
			ID:       maskedID,
			Name:     account.Name,
			ARN:      api.MaskARN(account.ARN),
			Email:    account.Email,
			Status:   account.Status,
			OUID:     account.OUID,
			OUName:   ouName,
			Controls: controlResponses,
			SCPs:     scpResponses,
		}

		api.WriteJSON(w, http.StatusOK, result)
	}
}

package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"go_api_server/internal/api"
	"go_api_server/internal/api/responses"
	"go_api_server/internal/store"
)

// ListControls godoc
// @Summary      List controls
// @Description  Returns all enabled controls, optionally filtered by ou_id
// @Tags         controls
// @Produce      json
// @Param        ou_id  query  string  false  "OU ID filter"
// @Success      200  {array}  responses.ControlResponse
// @Router       /api/controls [get]
func ListControls(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ouID := r.URL.Query().Get("ou_id")

		var result []responses.ControlResponse
		if ouID != "" {
			controls := s.ControlsForTarget(ouID)
			result = make([]responses.ControlResponse, 0, len(controls))
			for _, c := range controls {
				result = append(result, responses.ControlResponse{
					ARN:         c.ARN,
					ControlID:   c.ControlID,
					Name:        c.Name,
					ControlType: string(c.ControlType),
					Enforcement: string(c.Enforcement),
					TargetID:    c.TargetID,
					Description: c.Description,
				})
			}
		} else {
			controls := s.AllControls()
			result = make([]responses.ControlResponse, 0, len(controls))
			for _, c := range controls {
				result = append(result, responses.ControlResponse{
					ARN:         c.ARN,
					ControlID:   c.ControlID,
					Name:        c.Name,
					ControlType: string(c.ControlType),
					Enforcement: string(c.Enforcement),
					TargetID:    c.TargetID,
					Description: c.Description,
				})
			}
		}

		api.WriteJSON(w, http.StatusOK, result)
	}
}

// GetControl godoc
// @Summary      Get controls by ARN
// @Description  Returns all controls matching the given ARN. 404 if none found.
// @Tags         controls
// @Produce      json
// @Param        arn  path  string  true  "Control ARN"
// @Success      200  {array}   responses.ControlResponse
// @Failure      404  {object}  responses.ErrorResponse
// @Router       /api/controls/{arn} [get]
func GetControl(s *store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		arn := chi.URLParam(r, "*")

		controls := s.ControlsByARN(arn)
		if len(controls) == 0 {
			api.WriteError(w, http.StatusNotFound, "Control not found")
			return
		}

		result := make([]responses.ControlResponse, 0, len(controls))
		for _, c := range controls {
			result = append(result, responses.ControlResponse{
				ARN:         c.ARN,
				ControlID:   c.ControlID,
				Name:        c.Name,
				ControlType: string(c.ControlType),
				Enforcement: string(c.Enforcement),
				TargetID:    c.TargetID,
				Description: c.Description,
			})
		}

		api.WriteJSON(w, http.StatusOK, result)
	}
}

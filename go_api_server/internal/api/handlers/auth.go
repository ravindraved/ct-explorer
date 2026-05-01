package handlers

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"go_api_server/internal/api"
	"go_api_server/internal/api/responses"
	"go_api_server/internal/fetchers"
)

// strPtr returns a pointer to s, or nil if s is empty.
func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// AuthConfig returns the auth configuration for the SPA.
// When auth is enabled, returns Cognito details so the SPA can redirect to login.
// When auth is disabled, returns enabled=false so the SPA skips login.
func AuthConfig() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		config := map[string]any{
			"enabled": api.AuthEnabled,
		}
		if api.AuthEnabled {
			config["cognito_domain"] = api.CognitoDomain
			config["client_id"] = api.CognitoClientID
			config["region"] = api.CognitoRegion
		}
		api.WriteJSON(w, http.StatusOK, config)
	}
}

// AuthStatus godoc
// @Summary      Get authentication status
// @Description  Returns the current authentication state. No credentials exposed.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  responses.AuthStatusResponse
// @Router       /api/auth/status [get]
func AuthStatus(authState *fetchers.AuthState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := authState.Get()
		authMode := snap.AuthMode
		if authMode == "" {
			authMode = "instance_metadata"
		}
		resp := responses.AuthStatusResponse{
			Authenticated: snap.Authenticated,
			AccountID:     strPtr(api.MaskAccountID(snap.AccountID)),
			Region:        strPtr(snap.Region),
			Error:         strPtr(snap.Error),
			AuthMode:      authMode,
		}
		api.WriteJSON(w, http.StatusOK, resp)
	}
}

// AuthMetadata godoc
// @Summary      Get EC2 instance metadata
// @Description  Returns EC2 instance metadata via IMDSv2, or unavailable marker.
// @Tags         auth
// @Produce      json
// @Success      200  {object}  responses.AuthMetadataResponse
// @Router       /api/auth/metadata [get]
func AuthMetadata(authState *fetchers.AuthState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := authState.Get()
		authMode := snap.AuthMode
		if authMode == "" {
			authMode = "instance_metadata"
		}

		metadata, err := fetchers.GetIMDSMetadata(r.Context())
		if err != nil || !metadata.Available {
			resp := responses.AuthMetadataResponse{
				Available:        false,
				InstanceID:       nil,
				InstanceType:     nil,
				AvailabilityZone: nil,
				Region:           nil,
				IAMRole:          nil,
				AccountID:        nil,
				AuthMode:         authMode,
			}
			api.WriteJSON(w, http.StatusOK, resp)
			return
		}

		resp := responses.AuthMetadataResponse{
			Available:        true,
			InstanceID:       strPtr(metadata.InstanceID),
			InstanceType:     strPtr(metadata.InstanceType),
			AvailabilityZone: strPtr(metadata.AvailabilityZone),
			Region:           strPtr(metadata.Region),
			IAMRole:          strPtr(metadata.IAMRole),
			AccountID:        strPtr(api.MaskAccountID(metadata.AccountID)),
			AuthMode:         authMode,
		}
		api.WriteJSON(w, http.StatusOK, resp)
	}
}

// AuthConfigure godoc
// @Summary      Configure authentication
// @Description  Switch auth mode or set manual credentials. Never returns secrets.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  responses.AuthConfigureRequest  true  "Auth configuration"
// @Success      200  {object}  responses.AuthConfigureResponse
// @Failure      400  {object}  responses.ErrorResponse
// @Router       /api/auth/configure [post]
func AuthConfigure(authState *fetchers.AuthState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body responses.AuthConfigureRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			api.WriteError(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if body.AuthMode == "instance_metadata" {
			cfg, accountID, err := fetchers.ValidateCredentials(r.Context(), "")
			if err != nil {
				slog.Error("auth_configure_failed", "auth_mode", "instance_metadata")
				errMsg := err.Error()
				resp := responses.AuthConfigureResponse{
					Success:   false,
					AccountID: nil,
					Region:    nil,
					AuthMode:  "instance_metadata",
					Error:     &errMsg,
				}
				api.WriteJSON(w, http.StatusOK, resp)
				return
			}

			region := cfg.Region
			if region == "" {
				region = "us-east-1"
			}

			authState.Set(fetchers.AuthStateSnapshot{
				Authenticated: true,
				AccountID:     accountID,
				Region:        region,
				AuthMode:      "instance_metadata",
				Error:         "",
				AWSConfig:     cfg,
			})

			slog.Info("auth_configured", "auth_mode", "instance_metadata", "account_id", accountID)
			maskedID := api.MaskAccountID(accountID)
			resp := responses.AuthConfigureResponse{
				Success:   true,
				AccountID: &maskedID,
				Region:    &region,
				AuthMode:  "instance_metadata",
				Error:     nil,
			}
			api.WriteJSON(w, http.StatusOK, resp)
		} else {
			// manual mode
			opts := fetchers.ManualCredentialOpts{}
			if body.AccessKeyID != nil {
				opts.AccessKeyID = *body.AccessKeyID
			}
			if body.SecretAccessKey != nil {
				opts.SecretAccessKey = *body.SecretAccessKey
			}
			if body.SessionToken != nil {
				opts.SessionToken = *body.SessionToken
			}
			if body.ProfileName != nil {
				opts.ProfileName = *body.ProfileName
			}
			if body.Region != nil {
				opts.Region = *body.Region
			}

			cfg, accountID, err := fetchers.ValidateManualCredentials(r.Context(), opts)
			if err != nil {
				slog.Error("auth_configure_failed", "auth_mode", "manual")
				errMsg := err.Error()
				resp := responses.AuthConfigureResponse{
					Success:   false,
					AccountID: nil,
					Region:    nil,
					AuthMode:  "manual",
					Error:     &errMsg,
				}
				api.WriteJSON(w, http.StatusOK, resp)
				return
			}

			region := opts.Region
			if region == "" {
				region = cfg.Region
			}
			if region == "" {
				region = "us-east-1"
			}

			authState.Set(fetchers.AuthStateSnapshot{
				Authenticated: true,
				AccountID:     accountID,
				Region:        region,
				AuthMode:      "manual",
				Error:         "",
				AWSConfig:     cfg,
			})

			slog.Info("auth_configured", "auth_mode", "manual", "account_id", accountID)
			maskedID := api.MaskAccountID(accountID)
			resp := responses.AuthConfigureResponse{
				Success:   true,
				AccountID: &maskedID,
				Region:    &region,
				AuthMode:  "manual",
				Error:     nil,
			}
			api.WriteJSON(w, http.StatusOK, resp)
		}
	}
}

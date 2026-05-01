package handlers

import (
	"context"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"

	"go_api_server/internal/api"
	"go_api_server/internal/api/responses"
	"go_api_server/internal/fetchers"
	"go_api_server/internal/store"
)

// RefreshState is a thread-safe container for refresh progress tracking.
type RefreshState struct {
	mu              sync.Mutex
	Status          string
	Phase           string
	Error           string
	LastRefreshedAt string
	running         bool
}

// NewRefreshState creates a RefreshState with idle status.
func NewRefreshState() *RefreshState {
	return &RefreshState{Status: "idle"}
}

// Get returns a snapshot of the current refresh state as a response.
func (rs *RefreshState) Get() responses.RefreshStatusResponse {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	resp := responses.RefreshStatusResponse{
		Status: rs.Status,
	}
	if rs.Phase != "" {
		resp.Phase = &rs.Phase
	}
	if rs.Error != "" {
		resp.Error = &rs.Error
	}
	if rs.LastRefreshedAt != "" {
		resp.LastRefreshedAt = &rs.LastRefreshedAt
	}
	return resp
}

// SetRunning attempts to mark the state as running. Returns false if already running.
func (rs *RefreshState) SetRunning() bool {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if rs.running {
		return false
	}
	rs.running = true
	rs.Status = "in_progress"
	rs.Phase = "starting"
	rs.Error = ""
	return true
}

// SetPhase updates the current phase.
func (rs *RefreshState) SetPhase(phase string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.Phase = phase
}

// SetCompleted marks the refresh as completed.
func (rs *RefreshState) SetCompleted() {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.running = false
	rs.Status = "completed"
	rs.Phase = ""
	rs.Error = ""
	rs.LastRefreshedAt = time.Now().UTC().Format(time.RFC3339)
}

// SetFailed marks the refresh as failed with an error message.
func (rs *RefreshState) SetFailed(errMsg string) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.running = false
	rs.Status = "failed"
	rs.Phase = ""
	rs.Error = errMsg
}

// Refresh godoc
// @Summary      Trigger full data refresh
// @Description  Starts a background refresh of org, controls, and SCPs. 409 if already running, 400 if no session.
// @Tags         refresh
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  responses.ErrorResponse
// @Failure      409  {object}  responses.ErrorResponse
// @Router       /api/refresh [post]
func Refresh(s *store.Store, authState *fetchers.AuthState, refreshState *RefreshState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := authState.Get()
		if !snap.Authenticated || snap.AWSConfig == nil {
			api.WriteError(w, http.StatusBadRequest, "No AWS session available")
			return
		}

		if !refreshState.SetRunning() {
			api.WriteError(w, http.StatusConflict, "Refresh already in progress")
			return
		}

		slog.Info("refresh_started", "component", "refresh")

		cfg := *snap.AWSConfig
		go runOrgRefresh(context.Background(), s, cfg, refreshState)

		api.WriteJSON(w, http.StatusOK, map[string]string{
			"status":  "in_progress",
			"message": "Refresh started",
		})
	}
}

// runOrgRefresh performs the full org data refresh in a background goroutine.
func runOrgRefresh(ctx context.Context, s *store.Store, cfg aws.Config, refreshState *RefreshState) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("refresh_panic", "component", "refresh", "error", rec)
			refreshState.SetFailed("internal panic during refresh")
		}
	}()

	refreshState.SetPhase("validating_credentials")

	// Fetch into a temporary store
	tempStore := store.New()

	refreshState.SetPhase("fetching_organization")
	if err := fetchers.FetchOrganization(ctx, cfg, tempStore); err != nil {
		slog.Error("refresh_failed", "component", "refresh", "phase", "fetching_organization", "error", err.Error())
		refreshState.SetFailed(err.Error())
		return
	}

	refreshState.SetPhase("fetching_controls")
	if err := fetchers.FetchControls(ctx, cfg, tempStore); err != nil {
		slog.Error("refresh_failed", "component", "refresh", "phase", "fetching_controls", "error", err.Error())
		refreshState.SetFailed(err.Error())
		return
	}

	refreshState.SetPhase("fetching_scps")
	if err := fetchers.FetchSCPs(ctx, cfg, tempStore); err != nil {
		slog.Error("refresh_failed", "component", "refresh", "phase", "fetching_scps", "error", err.Error())
		refreshState.SetFailed(err.Error())
		return
	}

	// Atomic swap org data into the live store
	orgData := tempStore.ExportOrgData()
	s.SwapOrgData(orgData.OUs, orgData.Accounts, orgData.Controls, orgData.SCPs, orgData.RootOUID)

	refreshState.SetPhase("saving_cache")
	cacheDir := store.CacheDir()
	if err := store.SaveOrgCache(s, cacheDir); err != nil {
		slog.Warn("cache_save_failed", "component", "refresh", "error", err.Error())
	}

	refreshState.SetCompleted()
	slog.Info("refresh_complete", "component", "refresh")
}

// RefreshStatus godoc
// @Summary      Get refresh status
// @Description  Returns the current refresh state
// @Tags         refresh
// @Produce      json
// @Success      200  {object}  responses.RefreshStatusResponse
// @Router       /api/refresh/status [get]
func RefreshStatus(refreshState *RefreshState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, refreshState.Get())
	}
}

// CatalogRefresh godoc
// @Summary      Trigger catalog refresh
// @Description  Starts a background catalog-only refresh. 409 if already running, 400 if no session.
// @Tags         catalog
// @Produce      json
// @Success      200  {object}  map[string]string
// @Failure      400  {object}  responses.ErrorResponse
// @Failure      409  {object}  responses.ErrorResponse
// @Router       /api/catalog/refresh [post]
func CatalogRefresh(s *store.Store, authState *fetchers.AuthState, catalogRefreshState *RefreshState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		snap := authState.Get()
		if !snap.Authenticated || snap.AWSConfig == nil {
			api.WriteError(w, http.StatusBadRequest, "No AWS session available")
			return
		}

		if !catalogRefreshState.SetRunning() {
			api.WriteError(w, http.StatusConflict, "Catalog refresh already in progress")
			return
		}

		slog.Info("catalog_refresh_started", "component", "catalog_refresh")

		cfg := *snap.AWSConfig
		go runCatalogRefresh(context.Background(), s, cfg, catalogRefreshState)

		api.WriteJSON(w, http.StatusOK, map[string]string{
			"status":  "in_progress",
			"message": "Catalog refresh started",
		})
	}
}

// runCatalogRefresh performs the catalog-only refresh in a background goroutine.
func runCatalogRefresh(ctx context.Context, s *store.Store, cfg aws.Config, refreshState *RefreshState) {
	defer func() {
		if rec := recover(); rec != nil {
			slog.Error("catalog_refresh_panic", "component", "catalog_refresh", "error", rec)
			refreshState.SetFailed("internal panic during catalog refresh")
		}
	}()

	// Fetch into a temporary store
	tempStore := store.New()

	if err := fetchers.FetchCatalog(ctx, cfg, tempStore); err != nil {
		slog.Error("catalog_refresh_failed", "component", "catalog_refresh", "error", err.Error())
		refreshState.SetFailed(err.Error())
		return
	}

	// Atomic swap catalog data into the live store
	catData := tempStore.ExportCatalogData()
	s.SwapCatalogData(
		catData.CatalogControls,
		catData.CatalogDomains,
		catData.CatalogObjectives,
		catData.CatalogCommonControls,
		catData.ControlMappings,
	)

	cacheDir := store.CacheDir()
	if err := store.SaveCatalogCache(s, cacheDir); err != nil {
		slog.Warn("catalog_cache_save_failed", "component", "catalog_refresh", "error", err.Error())
	}

	refreshState.SetCompleted()
	slog.Info("catalog_refresh_complete", "component", "catalog_refresh")
}

// CatalogRefreshStatus godoc
// @Summary      Get catalog refresh status
// @Description  Returns the current catalog refresh state
// @Tags         catalog
// @Produce      json
// @Success      200  {object}  responses.RefreshStatusResponse
// @Router       /api/catalog/refresh/status [get]
func CatalogRefreshStatus(catalogRefreshState *RefreshState) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		api.WriteJSON(w, http.StatusOK, catalogRefreshState.Get())
	}
}

package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go_api_server/internal/api/handlers"
	"go_api_server/internal/api/router"
	"go_api_server/internal/fetchers"
	"go_api_server/internal/logging"
	"go_api_server/internal/store"
)

func main() {
	// ── 1. Parse environment variables ──
	port := envOr("PORT", "8000")
	logLevel := parseLogLevel(envOr("LOG_LEVEL", "INFO"))

	// ── 2. Initialize broadcaster and slog ──
	broadcaster := logging.NewBroadcaster()
	logging.Setup(logLevel, broadcaster)

	slog.Info("server_init", "port", port)

	// ── 3. Create store, load caches ──
	s := store.New()
	cacheDir := store.CacheDir()

	if err := store.LoadOrgCache(s, cacheDir); err != nil {
		slog.Warn("org_cache_load_error", "error", err)
	}

	if err := store.LoadCatalogCache(s, cacheDir); err != nil {
		slog.Warn("catalog_cache_load_error", "error", err)
	}

	// ── 4. Validate credentials (best-effort, degraded mode on failure) ──
	authState := &fetchers.AuthState{}
	ctx := context.Background()

	cfg, accountID, err := fetchers.ValidateCredentials(ctx, "")
	if err != nil {
		slog.Warn("credential_validation_failed", "error", err.Error())
		authState.Set(fetchers.AuthStateSnapshot{
			Authenticated: false,
			Error:         err.Error(),
			AuthMode:      "instance_metadata",
		})
	} else {
		region := cfg.Region
		if region == "" {
			region = "us-east-1"
		}
		slog.Info("credentials_validated", "account_id", accountID, "region", region)
		authState.Set(fetchers.AuthStateSnapshot{
			Authenticated: true,
			AccountID:     accountID,
			Region:        region,
			AuthMode:      "instance_metadata",
			AWSConfig:     cfg,
		})
	}

	// ── 5. If store is empty and session available: fetch all data, save cache ──
	snap := authState.Get()
	if s.IsEmpty() && snap.Authenticated && snap.AWSConfig != nil {
		slog.Info("store_empty_fetching_data", "component", "startup")
		awsCfg := *snap.AWSConfig

		if fetchErr := fetchers.FetchOrganization(ctx, awsCfg, s); fetchErr != nil {
			slog.Error("fetch_organization_failed", "error", fetchErr.Error())
		}
		if fetchErr := fetchers.FetchControls(ctx, awsCfg, s); fetchErr != nil {
			slog.Error("fetch_controls_failed", "error", fetchErr.Error())
		}
		if fetchErr := fetchers.FetchSCPs(ctx, awsCfg, s); fetchErr != nil {
			slog.Error("fetch_scps_failed", "error", fetchErr.Error())
		}

		if saveErr := store.SaveOrgCache(s, cacheDir); saveErr != nil {
			slog.Warn("org_cache_save_failed", "error", saveErr.Error())
		}
	}

	// ── 6. If catalog data empty and session available: fetch catalog, save cache ──
	if catalogEmpty(s) && snap.Authenticated && snap.AWSConfig != nil {
		slog.Info("catalog_empty_fetching_data", "component", "startup")
		awsCfg := *snap.AWSConfig

		if fetchErr := fetchers.FetchCatalog(ctx, awsCfg, s); fetchErr != nil {
			slog.Error("fetch_catalog_failed", "error", fetchErr.Error())
		}

		if saveErr := store.SaveCatalogCache(s, cacheDir); saveErr != nil {
			slog.Warn("catalog_cache_save_failed", "error", saveErr.Error())
		}
	}

	// ── 7. Create refresh states ──
	refreshState := handlers.NewRefreshState()
	catalogRefreshState := handlers.NewRefreshState()

	// ── 8. Build router ──
	r := router.NewRouter(s, authState, refreshState, catalogRefreshState, broadcaster)

	// ── 9. Start HTTP server with graceful shutdown ──
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		slog.Info("server_starting", "port", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server_error", "error", err)
			os.Exit(1)
		}
	}()

	<-done
	slog.Info("server_shutting_down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown_error", "error", err)
	}

	slog.Info("server_stopped")
}

// catalogEmpty returns true if the store has no catalog data.
func catalogEmpty(s *store.Store) bool {
	return len(s.AllCatalogDomains()) == 0 &&
		len(s.AllCatalogControls()) == 0 &&
		len(s.AllCatalogObjectives()) == 0
}

// envOr returns the value of the environment variable or the fallback.
func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// parseLogLevel converts a string log level to slog.Level.
func parseLogLevel(level string) slog.Level {
	switch strings.ToUpper(level) {
	case "DEBUG":
		return slog.LevelDebug
	case "WARN", "WARNING":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

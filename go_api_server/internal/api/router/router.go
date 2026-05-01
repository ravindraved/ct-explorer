// Package router provides the Chi v5 HTTP router for the CT Explorer Go API server.
package router

import (
	"encoding/json"
	"io/fs"
	"net/http"

	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	"github.com/go-chi/chi/v5"

	"go_api_server/internal/api"
	"go_api_server/internal/api/handlers"
	"go_api_server/internal/api/middleware"
	"go_api_server/internal/fetchers"
	"go_api_server/internal/logging"
	"go_api_server/internal/static"
	"go_api_server/internal/store"
)

// @title CT Explorer API
// @version 0.1.0
// @description JSON REST API for AWS Control Tower Explorer
// @host localhost:8000
// @BasePath /

// NewRouter builds the Chi router with all route groups, middleware, and handlers.
func NewRouter(
	s *store.Store,
	authState *fetchers.AuthState,
	refreshState *handlers.RefreshState,
	catalogRefreshState *handlers.RefreshState,
	broadcaster *logging.Broadcaster,
) *chi.Mux {
	r := chi.NewRouter()
	resolver := api.NewAccountIDResolver()

	// Global middleware
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.Security)
	r.Use(cors.Handler(middleware.CORSOptions()))
	r.Use(middleware.JWTAuth)

	// ── Auth config (no auth required — SPA needs this before login) ──
	r.Get("/api/auth/config", handlers.AuthConfig())

	// ── Auth routes ──
	r.Route("/api/auth", func(r chi.Router) {
		r.Get("/status", handlers.AuthStatus(authState))
		r.Get("/metadata", handlers.AuthMetadata(authState))
		r.Post("/configure", handlers.AuthConfigure(authState))
	})

	// ── Organization routes ──
	r.Route("/api/organization", func(r chi.Router) {
		r.Get("/tree", handlers.OrgTree(s, resolver))
		r.Get("/account/{accountID}", handlers.AccountDetail(s, resolver))
	})

	// ── Controls routes (wildcard for ARN paths containing slashes) ──
	r.Route("/api/controls", func(r chi.Router) {
		r.Get("/", handlers.ListControls(s))
		r.Get("/*", handlers.GetControl(s))
	})

	// ── SCPs routes ──
	r.Route("/api/scps", func(r chi.Router) {
		r.Get("/", handlers.ListSCPs(s))
		r.Get("/{scpID}", handlers.GetSCP(s))
	})

	// ── Catalog routes ──
	r.Route("/api/catalog", func(r chi.Router) {
		r.Get("/domains", handlers.CatalogDomains(s))
		r.Get("/objectives", handlers.CatalogObjectives(s))
		r.Get("/common-controls", handlers.CatalogCommonControls(s))
		r.Get("/controls", handlers.CatalogControls(s))
		r.Get("/controls/*", handlers.CatalogControlByARN(s))
		r.Get("/services", handlers.CatalogServices(s))
		r.Get("/enabled-map", handlers.CatalogEnabledMap(s))
		r.Post("/refresh", handlers.CatalogRefresh(s, authState, catalogRefreshState))
		r.Get("/refresh/status", handlers.CatalogRefreshStatus(catalogRefreshState))
	})

	// ── Ontology routes ──
	r.Route("/api/ontology", func(r chi.Router) {
		r.Get("/posture-tree", handlers.PostureTree(s))
		r.Get("/node/*", handlers.NodeControls(s)) // matches /node/{arn}/controls — handler strips trailing /controls
		r.Get("/control-map", handlers.ControlMap(s))
	})

	// ── Search routes ──
	r.Route("/api/search", func(r chi.Router) {
		r.Get("/find", handlers.SearchFind(s))
		r.Get("/coverage", handlers.SearchCoverage(s))
		r.Get("/path", handlers.SearchPath(s))
		r.Get("/quick-query", handlers.SearchQuickQuery(s))
	})

	// ── Refresh routes ──
	r.Post("/api/refresh", handlers.Refresh(s, authState, refreshState))
	r.Get("/api/refresh/status", handlers.RefreshStatus(refreshState))

	// ── WebSocket ──
	r.Get("/ws/logs", handlers.WSLogs(broadcaster))

	// ── Health check ──
	r.Get("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})

	// ── Embedded React frontend (SPA) ──
	webFS, err := static.FS()
	if err == nil {
		fileServer := http.FileServer(http.FS(webFS))
		r.Get("/*", func(w http.ResponseWriter, r *http.Request) {
			// Try to serve the file directly; if not found, serve index.html (SPA fallback)
			path := r.URL.Path
			if path == "/" {
				path = "/index.html"
			}
			// Check if file exists
			f, openErr := webFS.(fs.ReadFileFS).ReadFile(path[1:]) // strip leading /
			if openErr != nil {
				// SPA fallback: serve index.html for any unknown path
				r.URL.Path = "/"
				fileServer.ServeHTTP(w, r)
				return
			}
			_ = f
			fileServer.ServeHTTP(w, r)
		})
	}

	return r
}

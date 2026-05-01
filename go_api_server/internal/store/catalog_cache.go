package store

import (
	"bytes"
	"encoding/gob"
	"log/slog"
	"os"
	"path/filepath"

	"go_api_server/internal/models"
)

// CatalogCacheData holds all catalog-related data for Gob cache persistence.
type CatalogCacheData struct {
	CatalogControls       map[string]models.CatalogControl
	CatalogDomains        map[string]models.CatalogDomain
	CatalogObjectives     map[string]models.CatalogObjective
	CatalogCommonControls map[string]models.CatalogCommonControl
	ControlMappings       map[string][]string
}

// ExportCatalogData returns a snapshot of all catalog data under a read lock.
func (s *Store) ExportCatalogData() CatalogCacheData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	catalogControls := make(map[string]models.CatalogControl, len(s.catalogControls))
	for k, v := range s.catalogControls {
		catalogControls[k] = v
	}
	catalogDomains := make(map[string]models.CatalogDomain, len(s.catalogDomains))
	for k, v := range s.catalogDomains {
		catalogDomains[k] = v
	}
	catalogObjectives := make(map[string]models.CatalogObjective, len(s.catalogObjectives))
	for k, v := range s.catalogObjectives {
		catalogObjectives[k] = v
	}
	catalogCommonControls := make(map[string]models.CatalogCommonControl, len(s.catalogCommonControls))
	for k, v := range s.catalogCommonControls {
		catalogCommonControls[k] = v
	}
	controlMappings := make(map[string][]string, len(s.controlMappings))
	for k, v := range s.controlMappings {
		cp := make([]string, len(v))
		copy(cp, v)
		controlMappings[k] = cp
	}

	return CatalogCacheData{
		CatalogControls:       catalogControls,
		CatalogDomains:        catalogDomains,
		CatalogObjectives:     catalogObjectives,
		CatalogCommonControls: catalogCommonControls,
		ControlMappings:       controlMappings,
	}
}

// SaveCatalogCache serializes the store's catalog data to {dir}/catalog_cache.gob.
func SaveCatalogCache(store *Store, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data := store.ExportCatalogData()
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(data); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "catalog_cache.gob"), buf.Bytes(), 0o644)
}

// LoadCatalogCache reads {dir}/catalog_cache.gob and populates the store via
// SwapCatalogData. On missing or corrupt file it logs a warning and returns
// nil — the store remains empty.
func LoadCatalogCache(store *Store, dir string) error {
	path := filepath.Join(dir, "catalog_cache.gob")
	b, err := os.ReadFile(path)
	if err != nil {
		slog.Warn("catalog_cache_load_skip", "path", path, "error", err)
		return nil
	}

	var data CatalogCacheData
	if err := gob.NewDecoder(bytes.NewReader(b)).Decode(&data); err != nil {
		slog.Warn("catalog_cache_decode_error", "path", path, "error", err)
		return nil
	}

	// Ensure maps are non-nil
	if data.CatalogControls == nil {
		data.CatalogControls = make(map[string]models.CatalogControl)
	}
	if data.CatalogDomains == nil {
		data.CatalogDomains = make(map[string]models.CatalogDomain)
	}
	if data.CatalogObjectives == nil {
		data.CatalogObjectives = make(map[string]models.CatalogObjective)
	}
	if data.CatalogCommonControls == nil {
		data.CatalogCommonControls = make(map[string]models.CatalogCommonControl)
	}
	if data.ControlMappings == nil {
		data.ControlMappings = make(map[string][]string)
	}

	store.SwapCatalogData(
		data.CatalogControls,
		data.CatalogDomains,
		data.CatalogObjectives,
		data.CatalogCommonControls,
		data.ControlMappings,
	)
	return nil
}

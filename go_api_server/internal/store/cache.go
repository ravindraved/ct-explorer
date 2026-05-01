package store

import (
	"bytes"
	"encoding/gob"
	"log/slog"
	"os"
	"path/filepath"

	"go_api_server/internal/models"
)

// OrgCacheData holds all organization-related data for Gob cache persistence.
type OrgCacheData struct {
	OUs      map[string]models.OrgUnit
	Accounts map[string]models.Account
	Controls map[string]models.Control
	SCPs     map[string]models.SCP
	RootOUID string
}

func init() {
	// Register concrete types that may appear inside map[string]any (SCP Policy.Statement)
	gob.Register(map[string]any{})
	gob.Register([]any{})
	gob.Register([]map[string]any{})
}

// ExportOrgData returns a snapshot of all organization data under a read lock.
func (s *Store) ExportOrgData() OrgCacheData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ous := make(map[string]models.OrgUnit, len(s.ous))
	for k, v := range s.ous {
		ous[k] = v
	}
	accounts := make(map[string]models.Account, len(s.accounts))
	for k, v := range s.accounts {
		accounts[k] = v
	}
	controls := make(map[string]models.Control, len(s.controls))
	for k, v := range s.controls {
		controls[k] = v
	}
	scps := make(map[string]models.SCP, len(s.scps))
	for k, v := range s.scps {
		scps[k] = v
	}

	return OrgCacheData{
		OUs:      ous,
		Accounts: accounts,
		Controls: controls,
		SCPs:     scps,
		RootOUID: s.rootOUID,
	}
}

// CacheDir returns the cache directory path. It reads the CT_CACHE_DIR
// environment variable, falling back to ~/.aws_control_tower_explorer/.
func CacheDir() string {
	if dir := os.Getenv("CT_CACHE_DIR"); dir != "" {
		return dir
	}
	home, err := os.UserHomeDir()
	if err != nil {
		slog.Warn("cache_dir_home_error", "error", err)
		return ".aws_control_tower_explorer"
	}
	return filepath.Join(home, ".aws_control_tower_explorer")
}

// SaveOrgCache serializes the store's organization data to {dir}/cache.gob.
func SaveOrgCache(store *Store, dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	data := store.ExportOrgData()
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(data); err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(dir, "cache.gob"), buf.Bytes(), 0o644)
}

// LoadOrgCache reads {dir}/cache.gob and populates the store via SwapOrgData.
// On missing or corrupt file it logs a warning and returns nil — the store
// remains empty.
func LoadOrgCache(store *Store, dir string) error {
	path := filepath.Join(dir, "cache.gob")
	b, err := os.ReadFile(path)
	if err != nil {
		slog.Warn("org_cache_load_skip", "path", path, "error", err)
		return nil
	}

	var data OrgCacheData
	if err := gob.NewDecoder(bytes.NewReader(b)).Decode(&data); err != nil {
		slog.Warn("org_cache_decode_error", "path", path, "error", err)
		return nil
	}

	// Ensure maps are non-nil
	if data.OUs == nil {
		data.OUs = make(map[string]models.OrgUnit)
	}
	if data.Accounts == nil {
		data.Accounts = make(map[string]models.Account)
	}
	if data.Controls == nil {
		data.Controls = make(map[string]models.Control)
	}
	if data.SCPs == nil {
		data.SCPs = make(map[string]models.SCP)
	}

	store.SwapOrgData(data.OUs, data.Accounts, data.Controls, data.SCPs, data.RootOUID)
	return nil
}

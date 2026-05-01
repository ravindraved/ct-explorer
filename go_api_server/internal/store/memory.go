// Package store provides a thread-safe in-memory data store for AWS
// Control Tower Explorer data. All read methods return copies (not
// references to internal maps) and list methods return empty slices
// (not nil) when no items exist.
package store

import (
	"sync"

	"go_api_server/internal/models"
)

// Store is the central in-memory store for all AWS data.
// Read methods acquire RLock; write methods acquire Lock.
type Store struct {
	mu sync.RWMutex

	ous      map[string]models.OrgUnit
	accounts map[string]models.Account
	controls map[string]models.Control // keyed by "{arn}::{target_id}"
	scps     map[string]models.SCP

	catalogControls       map[string]models.CatalogControl
	catalogDomains        map[string]models.CatalogDomain
	catalogObjectives     map[string]models.CatalogObjective
	catalogCommonControls map[string]models.CatalogCommonControl
	controlMappings       map[string][]string // control_arn → []common_control_arn

	errors   []models.ErrorInfo
	rootOUID string
}

// New creates a Store with all maps initialized.
func New() *Store {
	return &Store{
		ous:                   make(map[string]models.OrgUnit),
		accounts:              make(map[string]models.Account),
		controls:              make(map[string]models.Control),
		scps:                  make(map[string]models.SCP),
		catalogControls:       make(map[string]models.CatalogControl),
		catalogDomains:        make(map[string]models.CatalogDomain),
		catalogObjectives:     make(map[string]models.CatalogObjective),
		catalogCommonControls: make(map[string]models.CatalogCommonControl),
		controlMappings:       make(map[string][]string),
		errors:                make([]models.ErrorInfo, 0),
	}
}

// ---------------------------------------------------------------------------
// OU operations
// ---------------------------------------------------------------------------

// GetOU returns the OrgUnit with the given ID and true, or a zero value and
// false if not found.
func (s *Store) GetOU(id string) (models.OrgUnit, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ou, ok := s.ous[id]
	return ou, ok
}

// AllOUs returns a copy of all OrgUnits in the store.
func (s *Store) AllOUs() []models.OrgUnit {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.OrgUnit, 0, len(s.ous))
	for _, ou := range s.ous {
		out = append(out, ou)
	}
	return out
}

// PutOU inserts or replaces an OrgUnit keyed by its ID.
func (s *Store) PutOU(ou models.OrgUnit) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ous[ou.ID] = ou
}

// ---------------------------------------------------------------------------
// Account operations
// ---------------------------------------------------------------------------

// GetAccount returns the Account with the given ID and true, or a zero value
// and false if not found.
func (s *Store) GetAccount(id string) (models.Account, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	acct, ok := s.accounts[id]
	return acct, ok
}

// AllAccounts returns a copy of all Accounts in the store.
func (s *Store) AllAccounts() []models.Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Account, 0, len(s.accounts))
	for _, a := range s.accounts {
		out = append(out, a)
	}
	return out
}

// AccountsForOU returns all accounts whose OUID matches the given OU ID.
func (s *Store) AccountsForOU(ouID string) []models.Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Account, 0)
	for _, a := range s.accounts {
		if a.OUID == ouID {
			out = append(out, a)
		}
	}
	return out
}

// PutAccount inserts or replaces an Account keyed by its ID.
func (s *Store) PutAccount(acct models.Account) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.accounts[acct.ID] = acct
}

// ---------------------------------------------------------------------------
// Control operations
// ---------------------------------------------------------------------------

// GetControl returns the Control with the given composite key
// ("{arn}::{target_id}") and true, or a zero value and false if not found.
func (s *Store) GetControl(key string) (models.Control, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ctrl, ok := s.controls[key]
	return ctrl, ok
}

// AllControls returns a copy of all Controls in the store.
func (s *Store) AllControls() []models.Control {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Control, 0, len(s.controls))
	for _, c := range s.controls {
		out = append(out, c)
	}
	return out
}

// ControlsForTarget returns all controls whose TargetID matches.
func (s *Store) ControlsForTarget(targetID string) []models.Control {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Control, 0)
	for _, c := range s.controls {
		if c.TargetID == targetID {
			out = append(out, c)
		}
	}
	return out
}

// ControlsByARN returns all controls whose ARN matches.
func (s *Store) ControlsByARN(arn string) []models.Control {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.Control, 0)
	for _, c := range s.controls {
		if c.ARN == arn {
			out = append(out, c)
		}
	}
	return out
}

// PutControl inserts or replaces a Control keyed by "{arn}::{target_id}".
func (s *Store) PutControl(ctrl models.Control) {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := ctrl.ARN + "::" + ctrl.TargetID
	s.controls[key] = ctrl
}

// ---------------------------------------------------------------------------
// SCP operations
// ---------------------------------------------------------------------------

// GetSCP returns the SCP with the given ID and true, or a zero value and
// false if not found.
func (s *Store) GetSCP(id string) (models.SCP, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	scp, ok := s.scps[id]
	return scp, ok
}

// AllSCPs returns a copy of all SCPs in the store.
func (s *Store) AllSCPs() []models.SCP {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.SCP, 0, len(s.scps))
	for _, scp := range s.scps {
		out = append(out, scp)
	}
	return out
}

// SCPsForTarget returns all SCPs whose TargetIDs contains the given target.
func (s *Store) SCPsForTarget(targetID string) []models.SCP {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.SCP, 0)
	for _, scp := range s.scps {
		for _, tid := range scp.TargetIDs {
			if tid == targetID {
				out = append(out, scp)
				break
			}
		}
	}
	return out
}

// PutSCP inserts or replaces an SCP keyed by its ID.
func (s *Store) PutSCP(scp models.SCP) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.scps[scp.ID] = scp
}

// ---------------------------------------------------------------------------
// Catalog Control operations
// ---------------------------------------------------------------------------

// GetCatalogControl returns the CatalogControl with the given ARN and true,
// or a zero value and false if not found.
func (s *Store) GetCatalogControl(arn string) (models.CatalogControl, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cc, ok := s.catalogControls[arn]
	return cc, ok
}

// AllCatalogControls returns a copy of all CatalogControls in the store.
func (s *Store) AllCatalogControls() []models.CatalogControl {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.CatalogControl, 0, len(s.catalogControls))
	for _, cc := range s.catalogControls {
		out = append(out, cc)
	}
	return out
}

// PutCatalogControl inserts or replaces a CatalogControl keyed by its ARN.
func (s *Store) PutCatalogControl(cc models.CatalogControl) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.catalogControls[cc.ARN] = cc
}

// ---------------------------------------------------------------------------
// Catalog Domain operations
// ---------------------------------------------------------------------------

// GetCatalogDomain returns the CatalogDomain with the given ARN and true,
// or a zero value and false if not found.
func (s *Store) GetCatalogDomain(arn string) (models.CatalogDomain, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cd, ok := s.catalogDomains[arn]
	return cd, ok
}

// AllCatalogDomains returns a copy of all CatalogDomains in the store.
func (s *Store) AllCatalogDomains() []models.CatalogDomain {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.CatalogDomain, 0, len(s.catalogDomains))
	for _, cd := range s.catalogDomains {
		out = append(out, cd)
	}
	return out
}

// PutCatalogDomain inserts or replaces a CatalogDomain keyed by its ARN.
func (s *Store) PutCatalogDomain(cd models.CatalogDomain) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.catalogDomains[cd.ARN] = cd
}

// ---------------------------------------------------------------------------
// Catalog Objective operations
// ---------------------------------------------------------------------------

// GetCatalogObjective returns the CatalogObjective with the given ARN and
// true, or a zero value and false if not found.
func (s *Store) GetCatalogObjective(arn string) (models.CatalogObjective, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	co, ok := s.catalogObjectives[arn]
	return co, ok
}

// AllCatalogObjectives returns a copy of all CatalogObjectives in the store.
func (s *Store) AllCatalogObjectives() []models.CatalogObjective {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.CatalogObjective, 0, len(s.catalogObjectives))
	for _, co := range s.catalogObjectives {
		out = append(out, co)
	}
	return out
}

// CatalogObjectivesForDomain returns all objectives whose DomainARN matches.
func (s *Store) CatalogObjectivesForDomain(domainARN string) []models.CatalogObjective {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.CatalogObjective, 0)
	for _, co := range s.catalogObjectives {
		if co.DomainARN == domainARN {
			out = append(out, co)
		}
	}
	return out
}

// PutCatalogObjective inserts or replaces a CatalogObjective keyed by ARN.
func (s *Store) PutCatalogObjective(co models.CatalogObjective) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.catalogObjectives[co.ARN] = co
}

// ---------------------------------------------------------------------------
// Catalog Common Control operations
// ---------------------------------------------------------------------------

// GetCatalogCommonControl returns the CatalogCommonControl with the given ARN
// and true, or a zero value and false if not found.
func (s *Store) GetCatalogCommonControl(arn string) (models.CatalogCommonControl, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ccc, ok := s.catalogCommonControls[arn]
	return ccc, ok
}

// AllCatalogCommonControls returns a copy of all CatalogCommonControls.
func (s *Store) AllCatalogCommonControls() []models.CatalogCommonControl {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.CatalogCommonControl, 0, len(s.catalogCommonControls))
	for _, ccc := range s.catalogCommonControls {
		out = append(out, ccc)
	}
	return out
}

// CatalogCommonControlsForObjective returns all common controls whose
// ObjectiveARN matches.
func (s *Store) CatalogCommonControlsForObjective(objectiveARN string) []models.CatalogCommonControl {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.CatalogCommonControl, 0)
	for _, ccc := range s.catalogCommonControls {
		if ccc.ObjectiveARN == objectiveARN {
			out = append(out, ccc)
		}
	}
	return out
}

// PutCatalogCommonControl inserts or replaces a CatalogCommonControl keyed
// by its ARN.
func (s *Store) PutCatalogCommonControl(ccc models.CatalogCommonControl) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.catalogCommonControls[ccc.ARN] = ccc
}

// ---------------------------------------------------------------------------
// Control Mappings
// ---------------------------------------------------------------------------

// GetControlMappings returns a deep copy of the control_arn → []common_control_arn
// mapping.
func (s *Store) GetControlMappings() map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string][]string, len(s.controlMappings))
	for k, v := range s.controlMappings {
		cp := make([]string, len(v))
		copy(cp, v)
		out[k] = cp
	}
	return out
}

// PutControlMapping sets the common control ARNs for a given control ARN.
func (s *Store) PutControlMapping(controlARN string, commonControlARNs []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	cp := make([]string, len(commonControlARNs))
	copy(cp, commonControlARNs)
	s.controlMappings[controlARN] = cp
}

// ---------------------------------------------------------------------------
// Root OU ID
// ---------------------------------------------------------------------------

// GetRootOUID returns the root OU ID.
func (s *Store) GetRootOUID() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.rootOUID
}

// SetRootOUID sets the root OU ID.
func (s *Store) SetRootOUID(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.rootOUID = id
}

// ---------------------------------------------------------------------------
// Errors
// ---------------------------------------------------------------------------

// AllErrors returns a copy of all recorded errors.
func (s *Store) AllErrors() []models.ErrorInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]models.ErrorInfo, len(s.errors))
	copy(out, s.errors)
	return out
}

// AddError appends an error to the store.
func (s *Store) AddError(err models.ErrorInfo) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.errors = append(s.errors, err)
}

// ---------------------------------------------------------------------------
// IsEmpty
// ---------------------------------------------------------------------------

// IsEmpty returns true if the store contains no OUs, accounts, controls, or
// SCPs.
func (s *Store) IsEmpty() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.ous) == 0 && len(s.accounts) == 0 &&
		len(s.controls) == 0 && len(s.scps) == 0
}

// ---------------------------------------------------------------------------
// Atomic bulk swap
// ---------------------------------------------------------------------------

// SwapOrgData atomically replaces all organization-related maps under a write
// lock. This ensures readers never see partial state during a refresh.
func (s *Store) SwapOrgData(
	ous map[string]models.OrgUnit,
	accounts map[string]models.Account,
	controls map[string]models.Control,
	scps map[string]models.SCP,
	rootOUID string,
) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ous = ous
	s.accounts = accounts
	s.controls = controls
	s.scps = scps
	s.rootOUID = rootOUID
}

// SwapCatalogData atomically replaces all catalog-related maps under a write
// lock.
func (s *Store) SwapCatalogData(
	catalogControls map[string]models.CatalogControl,
	catalogDomains map[string]models.CatalogDomain,
	catalogObjectives map[string]models.CatalogObjective,
	catalogCommonControls map[string]models.CatalogCommonControl,
	controlMappings map[string][]string,
) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.catalogControls = catalogControls
	s.catalogDomains = catalogDomains
	s.catalogObjectives = catalogObjectives
	s.catalogCommonControls = catalogCommonControls
	s.controlMappings = controlMappings
}



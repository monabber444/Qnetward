package msp

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	ErrPartnerNotFound = errors.New("partner not found")
	ErrOrgNotFound     = errors.New("organization not found")
)

// Manager coordinates multi-tenant MSP hierarchy and co-managed access control.
type Manager struct {
	mu           sync.RWMutex
	partners     map[string]*Partner
	orgs         map[string]*Organization
	environments map[string][]*Environment
}

// NewManager creates an MSP multi-tenant manager.
func NewManager() *Manager {
	return &Manager{
		partners:     make(map[string]*Partner),
		orgs:         make(map[string]*Organization),
		environments: make(map[string][]*Environment),
	}
}

// RegisterPartner registers a top-level MSP or reseller partner.
func (m *Manager) RegisterPartner(p Partner) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	m.partners[p.ID] = &p
}

// RegisterOrganization registers a customer enterprise under an MSP partner.
func (m *Manager) RegisterOrganization(org Organization) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.partners[org.PartnerID]; !exists {
		return fmt.Errorf("%w: partner ID %s", ErrPartnerNotFound, org.PartnerID)
	}

	if org.CreatedAt.IsZero() {
		org.CreatedAt = time.Now()
	}
	m.orgs[org.ID] = &org
	return nil
}

// RegisterEnvironment adds an environment (Production, Staging, DR) to an organization.
func (m *Manager) RegisterEnvironment(env Environment) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.orgs[env.OrgID]; !exists {
		return fmt.Errorf("%w: organization ID %s", ErrOrgNotFound, env.OrgID)
	}

	m.environments[env.OrgID] = append(m.environments[env.OrgID], &env)
	return nil
}

// Authorize checks whether a co-managed user has permission to execute an action in an organization.
func (m *Manager) Authorize(user CoManagedUser, orgID string, action string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 1. Super Admins have global access
	if user.Role == RoleMSPSuperAdmin {
		return true
	}

	org, exists := m.orgs[orgID]
	if !exists {
		return false
	}

	// 2. Partner Admins have full access across all orgs under their partner account
	if user.Role == RolePartnerAdmin {
		if org.PartnerID != user.PartnerID {
			return false
		}
		return actionAllowed(user.AllowedActions, action)
	}

	// 3. Client Admins and ReadOnly Operators must be explicitly assigned to the Org
	foundOrg := false
	for _, id := range user.AllowedOrgIDs {
		if id == orgID {
			foundOrg = true
			break
		}
	}
	if !foundOrg {
		return false
	}

	if user.Role == RoleReadOnlyOperator && action != "read" {
		return false
	}

	return actionAllowed(user.AllowedActions, action)
}

func actionAllowed(allowed []string, requested string) bool {
	if len(allowed) == 0 {
		return true
	}
	for _, a := range allowed {
		if a == "*" || a == requested {
			return true
		}
	}
	return false
}

// GetPartner returns partner details.
func (m *Manager) GetPartner(id string) (*Partner, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.partners[id]
	return p, ok
}

// GetPartnerOrgs returns all organizations belonging to a partner.
func (m *Manager) GetPartnerOrgs(partnerID string) []*Organization {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var res []*Organization
	for _, o := range m.orgs {
		if o.PartnerID == partnerID {
			res = append(res, o)
		}
	}
	return res
}

// GetEnvironments returns all environments under an organization.
func (m *Manager) GetEnvironments(orgID string) []*Environment {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.environments[orgID]
}

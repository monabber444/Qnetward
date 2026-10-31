package msp

import (
	"testing"
)

func TestMSPManager_HierarchyAndAuthorization(t *testing.T) {
	mgr := NewManager()

	// 1. Register Partner with White-Label Branding
	partner := Partner{
		ID:   "partner-apex",
		Name: "Apex Cloud Managed Services",
		Tier: "PLATINUM",
		WhiteLabel: WhiteLabelConfig{
			BrandName:       "ApexNet",
			CustomDomain:    "net.apexcloud.io",
			PrimaryColorHex: "#0ea5e9",
			LogoURL:         "https://cdn.apexcloud.io/logo.svg",
			ReportHeader:    "Confidential Security & Network Audit Report",
		},
	}
	mgr.RegisterPartner(partner)

	p, ok := mgr.GetPartner("partner-apex")
	if !ok || p.WhiteLabel.BrandName != "ApexNet" {
		t.Fatalf("Failed to retrieve partner or white-label config")
	}

	// 2. Register Organizations under Partner
	orgFin := Organization{
		ID:        "org-fintech",
		PartnerID: "partner-apex",
		Name:      "Global Fintech Ltd",
		Industry:  "Financial Services",
		SLA:       "99.999%",
		Status:    "ACTIVE",
	}
	if err := mgr.RegisterOrganization(orgFin); err != nil {
		t.Fatalf("Failed to register orgFin: %v", err)
	}

	orgHealth := Organization{
		ID:        "org-health",
		PartnerID: "partner-apex",
		Name:      "Northwest Hospital Network",
		Industry:  "Healthcare",
		SLA:       "99.99%",
		Status:    "ACTIVE",
	}
	if err := mgr.RegisterOrganization(orgHealth); err != nil {
		t.Fatalf("Failed to register orgHealth: %v", err)
	}

	// 3. Register Environment
	envProd := Environment{
		ID:        "env-fintech-prod",
		OrgID:     "org-fintech",
		Name:      "Production Core",
		Region:    "us-east-1",
		NodeCount: 16,
	}
	if err := mgr.RegisterEnvironment(envProd); err != nil {
		t.Fatalf("Failed to register env: %v", err)
	}

	envs := mgr.GetEnvironments("org-fintech")
	if len(envs) != 1 || envs[0].ID != "env-fintech-prod" {
		t.Errorf("Expected 1 environment for org-fintech, got %d", len(envs))
	}

	// 4. Test Authorization Matrix
	// User A: MSP Super Admin (can do anything everywhere)
	userSuper := CoManagedUser{
		ID:   "user-super",
		Role: RoleMSPSuperAdmin,
	}
	if !mgr.Authorize(userSuper, "org-fintech", "remediate") {
		t.Errorf("Super Admin should be authorized for any action")
	}

	// User B: Partner Admin for Apex (can manage fintech & health, but not an org from another partner)
	userPartnerAdmin := CoManagedUser{
		ID:        "user-partner",
		PartnerID: "partner-apex",
		Role:      RolePartnerAdmin,
	}
	if !mgr.Authorize(userPartnerAdmin, "org-fintech", "configure") {
		t.Errorf("Partner admin should be authorized on client orgs under its partner ID")
	}
	if mgr.Authorize(userPartnerAdmin, "org-unrelated-foreign", "configure") {
		t.Errorf("Partner admin should NOT be authorized on unrelated foreign org")
	}

	// User C: Client Admin (restricted to Fintech org only)
	userClientAdmin := CoManagedUser{
		ID:            "user-client",
		Role:          RoleClientAdmin,
		AllowedOrgIDs: []string{"org-fintech"},
	}
	if !mgr.Authorize(userClientAdmin, "org-fintech", "remediate") {
		t.Errorf("Client admin should be authorized on org-fintech")
	}
	if mgr.Authorize(userClientAdmin, "org-health", "remediate") {
		t.Errorf("Client admin should NOT be authorized on org-health")
	}

	// User D: Read-Only Operator (cannot perform write/remediate actions)
	userReadOnly := CoManagedUser{
		ID:            "user-readonly",
		Role:          RoleReadOnlyOperator,
		AllowedOrgIDs: []string{"org-fintech"},
	}
	if !mgr.Authorize(userReadOnly, "org-fintech", "read") {
		t.Errorf("Read-only operator should be allowed to read")
	}
	if mgr.Authorize(userReadOnly, "org-fintech", "remediate") {
		t.Errorf("Read-only operator should NOT be allowed to remediate")
	}
}

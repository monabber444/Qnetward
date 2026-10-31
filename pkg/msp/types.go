package msp

import "time"

// Role defines co-managed access privileges across the MSP hierarchy.
type Role string

const (
	RoleMSPSuperAdmin     Role = "MSP_SUPER_ADMIN"
	RolePartnerAdmin      Role = "PARTNER_ADMIN"
	RoleClientAdmin       Role = "CLIENT_ADMIN"
	RoleReadOnlyOperator  Role = "READ_ONLY_OPERATOR"
)

// WhiteLabelConfig configures custom branding for an MSP partner.
type WhiteLabelConfig struct {
	BrandName       string `json:"brand_name"`
	CustomDomain    string `json:"custom_domain"`
	PrimaryColorHex string `json:"primary_color_hex"`
	LogoURL         string `json:"logo_url"`
	ReportHeader    string `json:"report_header"`
}

// Partner represents an MSP or reseller partner managing multiple organizations.
type Partner struct {
	ID          string           `json:"id"`
	Name        string           `json:"name"`
	Tier        string           `json:"tier"` // "PLATINUM", "GOLD", "SILVER"
	WhiteLabel  WhiteLabelConfig `json:"white_label"`
	CreatedAt   time.Time        `json:"created_at"`
}

// Organization represents a customer enterprise under an MSP partner.
type Organization struct {
	ID          string    `json:"id"`
	PartnerID   string    `json:"partner_id"`
	Name        string    `json:"name"`
	Industry    string    `json:"industry"`
	SLA         string    `json:"sla"` // "99.9%", "99.99%", "99.999%"
	Status      string    `json:"status"` // "ACTIVE", "SUSPENDED"
	CreatedAt   time.Time `json:"created_at"`
}

// Environment represents an infrastructure silo (e.g. Production, Staging, DR).
type Environment struct {
	ID        string `json:"id"`
	OrgID     string `json:"org_id"`
	Name      string `json:"name"`
	Region    string `json:"region"`
	NodeCount int    `json:"node_count"`
}

// CoManagedUser represents an operator with scoped multi-tenant permissions.
type CoManagedUser struct {
	ID             string   `json:"id"`
	Email          string   `json:"email"`
	PartnerID      string   `json:"partner_id"`
	Role           Role     `json:"role"`
	AllowedOrgIDs  []string `json:"allowed_org_ids"` // Empty slice if partner/super admin
	AllowedActions []string `json:"allowed_actions"` // e.g. ["read", "remediate", "configure"]
}

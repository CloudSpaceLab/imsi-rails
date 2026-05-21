package auth

import "time"

type Permission string

const (
	PermissionDashboardRead     Permission = "dashboard:read"
	PermissionTransactionsRead  Permission = "transactions:read"
	PermissionTransactionsTrace Permission = "transactions:trace"
	PermissionProvidersManage   Permission = "providers:manage"
	PermissionPolicyDraft       Permission = "policy:draft"
	PermissionPolicyApprove     Permission = "policy:approve"
	PermissionPolicyActivate    Permission = "policy:activate"
	PermissionIncidentsManage   Permission = "incidents:manage"
	PermissionReconciliation    Permission = "reconciliation:manage"
	PermissionFXRead            Permission = "fx:read"
	PermissionAuditRead         Permission = "audit:read"
	PermissionAuditExport       Permission = "audit:export"
	PermissionUsersManage       Permission = "users:manage"
	PermissionIdentityManage    Permission = "identity:manage"
)

type Role string

const (
	RolePlatformAdmin     Role = "platform_admin"
	RoleBankAdmin         Role = "bank_admin"
	RoleOpsLead           Role = "ops_lead"
	RoleOpsAnalyst        Role = "ops_analyst"
	RoleComplianceOfficer Role = "compliance_officer"
	RoleTreasuryFinance   Role = "treasury_finance"
	RoleAuditor           Role = "auditor"
	RoleViewer            Role = "viewer"
	RoleDemoViewer        Role = "demo_viewer"
)

var RolePermissions = map[Role][]Permission{
	RolePlatformAdmin: {
		PermissionDashboardRead, PermissionTransactionsRead, PermissionTransactionsTrace,
		PermissionProvidersManage, PermissionPolicyDraft, PermissionPolicyApprove, PermissionPolicyActivate,
		PermissionIncidentsManage, PermissionReconciliation, PermissionFXRead,
		PermissionAuditRead, PermissionAuditExport, PermissionUsersManage, PermissionIdentityManage,
	},
	RoleBankAdmin: {
		PermissionDashboardRead, PermissionTransactionsRead, PermissionTransactionsTrace,
		PermissionProvidersManage, PermissionPolicyDraft, PermissionPolicyApprove, PermissionPolicyActivate,
		PermissionIncidentsManage, PermissionReconciliation, PermissionFXRead,
		PermissionAuditRead, PermissionAuditExport, PermissionUsersManage, PermissionIdentityManage,
	},
	RoleOpsLead: {
		PermissionDashboardRead, PermissionTransactionsRead, PermissionTransactionsTrace,
		PermissionPolicyDraft, PermissionPolicyApprove, PermissionPolicyActivate,
		PermissionIncidentsManage, PermissionReconciliation, PermissionFXRead, PermissionAuditRead,
	},
	RoleOpsAnalyst: {
		PermissionDashboardRead, PermissionTransactionsRead, PermissionTransactionsTrace,
		PermissionPolicyDraft, PermissionIncidentsManage, PermissionReconciliation, PermissionFXRead, PermissionAuditRead,
	},
	RoleComplianceOfficer: {
		PermissionDashboardRead, PermissionTransactionsRead, PermissionTransactionsTrace,
		PermissionPolicyApprove, PermissionAuditRead, PermissionAuditExport,
	},
	RoleTreasuryFinance: {
		PermissionDashboardRead, PermissionTransactionsRead, PermissionFXRead, PermissionReconciliation, PermissionAuditRead,
	},
	RoleAuditor: {
		PermissionDashboardRead, PermissionTransactionsRead, PermissionAuditRead, PermissionAuditExport,
	},
	RoleViewer: {
		PermissionDashboardRead, PermissionTransactionsRead, PermissionFXRead,
	},
	RoleDemoViewer: {
		PermissionDashboardRead, PermissionTransactionsRead, PermissionFXRead,
	},
}

type SessionUser struct {
	ID           string       `json:"id"`
	BankID       string       `json:"bank_id"`
	Username     string       `json:"username"`
	Email        string       `json:"email"`
	DisplayName  string       `json:"display_name"`
	Roles        []Role       `json:"roles"`
	Permissions  []Permission `json:"permissions"`
	AuthProvider string       `json:"auth_provider"`
}

func (u SessionUser) HasPermission(permission Permission) bool {
	for _, candidate := range u.Permissions {
		if candidate == permission {
			return true
		}
	}
	return false
}

type UserStatus string

const (
	UserActive   UserStatus = "active"
	UserDisabled UserStatus = "disabled"
)

type User struct {
	ID           string     `json:"id"`
	BankID       string     `json:"bank_id"`
	Username     string     `json:"username"`
	Email        string     `json:"email"`
	DisplayName  string     `json:"display_name"`
	Roles        []Role     `json:"roles"`
	Status       UserStatus `json:"status"`
	AuthProvider string     `json:"auth_provider"`
	CreatedAt    time.Time  `json:"created_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

func (u User) SessionUser() SessionUser {
	return SessionUser{
		ID:           u.ID,
		BankID:       u.BankID,
		Username:     u.Username,
		Email:        u.Email,
		DisplayName:  u.DisplayName,
		Roles:        append([]Role(nil), u.Roles...),
		Permissions:  PermissionsForRoles(u.Roles),
		AuthProvider: u.AuthProvider,
	}
}

type ProviderType string

const (
	ProviderLocal      ProviderType = "local"
	ProviderLDAP       ProviderType = "ldap"
	ProviderGoogleOIDC ProviderType = "google_oidc"
)

type AuthProviderConfig struct {
	ID           string             `json:"id"`
	BankID       string             `json:"bank_id"`
	Type         ProviderType       `json:"type"`
	Name         string             `json:"name"`
	Enabled      bool               `json:"enabled"`
	LDAP         *LDAPConfig        `json:"ldap,omitempty"`
	GoogleOIDC   *GoogleOIDCConfig  `json:"google_oidc,omitempty"`
	DefaultRoles []Role             `json:"default_roles,omitempty"`
	GroupMapping []LDAPGroupMapping `json:"group_mapping,omitempty"`
	CreatedAt    time.Time          `json:"created_at"`
	UpdatedAt    time.Time          `json:"updated_at"`
}

type LDAPConfig struct {
	Host                 string `json:"host"`
	Port                 int    `json:"port"`
	TLSMode              string `json:"tls_mode"` // tls, starttls, or none for controlled dev only.
	BaseDN               string `json:"base_dn"`
	BindDN               string `json:"bind_dn,omitempty"`
	BindPasswordSecret   string `json:"bind_password_secret,omitempty"`
	UserFilter           string `json:"user_filter"`
	GroupFilter          string `json:"group_filter"`
	UsernameAttribute    string `json:"username_attribute"`
	EmailAttribute       string `json:"email_attribute"`
	DisplayNameAttribute string `json:"display_name_attribute"`
	GroupNameAttribute   string `json:"group_name_attribute"`
}

type GoogleOIDCConfig struct {
	Issuer         string   `json:"issuer"`
	ClientID       string   `json:"client_id"`
	ClientSecret   string   `json:"client_secret,omitempty"`
	RedirectURL    string   `json:"redirect_url"`
	AllowedDomains []string `json:"allowed_domains"`
}

type LDAPGroupMapping struct {
	ExternalGroup string `json:"external_group"`
	Roles         []Role `json:"roles"`
}

func PermissionsForRoles(roles []Role) []Permission {
	seen := map[Permission]bool{}
	permissions := make([]Permission, 0)
	for _, role := range roles {
		for _, permission := range RolePermissions[role] {
			if !seen[permission] {
				seen[permission] = true
				permissions = append(permissions, permission)
			}
		}
	}
	return permissions
}

func KnownRoles() []Role {
	return []Role{
		RolePlatformAdmin,
		RoleBankAdmin,
		RoleOpsLead,
		RoleOpsAnalyst,
		RoleComplianceOfficer,
		RoleTreasuryFinance,
		RoleAuditor,
		RoleViewer,
		RoleDemoViewer,
	}
}

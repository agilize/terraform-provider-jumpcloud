package organization

// Organization representa uma organização no JumpCloud
type Organization struct {
	ID             string            `json:"_id,omitempty"`
	Name           string            `json:"name"`
	DisplayName    string            `json:"displayName,omitempty"`
	LogoURL        string            `json:"logoUrl,omitempty"`
	Website        string            `json:"website,omitempty"`
	ContactName    string            `json:"contactName,omitempty"`
	ContactEmail   string            `json:"contactEmail,omitempty"`
	ContactPhone   string            `json:"contactPhone,omitempty"`
	Settings       map[string]string `json:"settings,omitempty"`
	ParentOrgID    string            `json:"parentOrgId,omitempty"`
	AllowedDomains []string          `json:"allowedDomains,omitempty"`
	Created        string            `json:"created,omitempty"`
	Updated        string            `json:"updated,omitempty"`
}

// OrganizationSettings representa a estrutura de configurações de uma organização no JumpCloud
type OrganizationSettings struct {
	ID                           string          `json:"_id,omitempty"`
	OrgID                        string          `json:"orgId"`
	PasswordPolicy               *PasswordPolicy `json:"passwordPolicy,omitempty"`
	SystemInsightsEnabled        bool            `json:"systemInsightsEnabled"`
	NewSystemUserStateManaged    bool            `json:"newSystemUserStateManaged"`
	NewUserEmailTemplate         string          `json:"newUserEmailTemplate,omitempty"`
	PasswordResetTemplate        string          `json:"passwordResetTemplate,omitempty"`
	DirectoryInsightsEnabled     bool            `json:"directoryInsightsEnabled"`
	LdapIntegrationEnabled       bool            `json:"ldapIntegrationEnabled"`
	AllowPublicKeyAuthentication bool            `json:"allowPublicKeyAuthentication"`
	AllowMultiFactorAuth         bool            `json:"allowMultiFactorAuth"`
	RequireMfa                   bool            `json:"requireMfa"`
	AllowedMfaMethods            []string        `json:"allowedMfaMethods,omitempty"`
	Created                      string          `json:"created,omitempty"`
	Updated                      string          `json:"updated,omitempty"`
}

// PasswordPolicy define as configurações de política de senha da organização
type PasswordPolicy struct {
	MinLength           int  `json:"minLength"`
	RequiresLowercase   bool `json:"requiresLowercase"`
	RequiresUppercase   bool `json:"requiresUppercase"`
	RequiresNumber      bool `json:"requiresNumber"`
	RequiresSpecialChar bool `json:"requiresSpecialChar"`
	ExpirationDays      int  `json:"expirationDays"`
	MaxHistory          int  `json:"maxHistory"`
}

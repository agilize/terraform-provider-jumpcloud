package organization

import (
	"testing"
)

func TestOrganization(t *testing.T) {
	// Test creating an Organization instance
	org := &Organization{
		ID:             "test-id",
		Name:           "Test Org",
		DisplayName:    "Test Organization",
		LogoURL:        "https://example.com/logo.png",
		Website:        "https://example.com",
		ContactName:    "Test Contact",
		ContactEmail:   "contact@example.com",
		ContactPhone:   "+1234567890",
		Settings:       map[string]string{"key": "value"},
		ParentOrgID:    "parent-id",
		AllowedDomains: []string{"example.com"},
		Created:        "2023-01-01T00:00:00Z",
		Updated:        "2023-01-01T01:00:00Z",
	}

	// Validate fields
	if org.ID != "test-id" {
		t.Errorf("Expected ID to be 'test-id', got '%s'", org.ID)
	}

	if org.Name != "Test Org" {
		t.Errorf("Expected Name to be 'Test Org', got '%s'", org.Name)
	}

	if org.DisplayName != "Test Organization" {
		t.Errorf("Expected DisplayName to be 'Test Organization', got '%s'", org.DisplayName)
	}

	if org.LogoURL != "https://example.com/logo.png" {
		t.Errorf("Expected LogoURL to be 'https://example.com/logo.png', got '%s'", org.LogoURL)
	}

	if org.Website != "https://example.com" {
		t.Errorf("Expected Website to be 'https://example.com', got '%s'", org.Website)
	}

	if org.ContactName != "Test Contact" {
		t.Errorf("Expected ContactName to be 'Test Contact', got '%s'", org.ContactName)
	}

	if org.ContactEmail != "contact@example.com" {
		t.Errorf("Expected ContactEmail to be 'contact@example.com', got '%s'", org.ContactEmail)
	}

	if org.ContactPhone != "+1234567890" {
		t.Errorf("Expected ContactPhone to be '+1234567890', got '%s'", org.ContactPhone)
	}

	if org.Settings["key"] != "value" {
		t.Errorf("Expected Settings['key'] to be 'value', got '%s'", org.Settings["key"])
	}

	if org.ParentOrgID != "parent-id" {
		t.Errorf("Expected ParentOrgID to be 'parent-id', got '%s'", org.ParentOrgID)
	}

	if len(org.AllowedDomains) != 1 || org.AllowedDomains[0] != "example.com" {
		t.Errorf("Expected AllowedDomains to contain 'example.com', got %v", org.AllowedDomains)
	}

	if org.Created != "2023-01-01T00:00:00Z" {
		t.Errorf("Expected Created to be '2023-01-01T00:00:00Z', got '%s'", org.Created)
	}

	if org.Updated != "2023-01-01T01:00:00Z" {
		t.Errorf("Expected Updated to be '2023-01-01T01:00:00Z', got '%s'", org.Updated)
	}
}

func TestOrganizationSettings(t *testing.T) {
	passwordPolicy := &PasswordPolicy{
		MinLength:           12,
		RequiresLowercase:   true,
		RequiresUppercase:   true,
		RequiresNumber:      true,
		RequiresSpecialChar: true,
		ExpirationDays:      90,
		MaxHistory:          5,
	}

	// Test creating an OrganizationSettings instance
	settings := &OrganizationSettings{
		ID:                           "settings-id",
		OrgID:                        "org-id",
		PasswordPolicy:               passwordPolicy,
		SystemInsightsEnabled:        true,
		NewSystemUserStateManaged:    true,
		NewUserEmailTemplate:         "Welcome!",
		PasswordResetTemplate:        "Reset your password",
		DirectoryInsightsEnabled:     true,
		LdapIntegrationEnabled:       false,
		AllowPublicKeyAuthentication: true,
		AllowMultiFactorAuth:         true,
		RequireMfa:                   false,
		AllowedMfaMethods:            []string{"totp", "push"},
		Created:                      "2023-01-01T00:00:00Z",
		Updated:                      "2023-01-01T01:00:00Z",
	}

	// Validate fields
	if settings.ID != "settings-id" {
		t.Errorf("Expected ID to be 'settings-id', got '%s'", settings.ID)
	}

	if settings.OrgID != "org-id" {
		t.Errorf("Expected OrgID to be 'org-id', got '%s'", settings.OrgID)
	}

	if settings.PasswordPolicy == nil {
		t.Errorf("Expected PasswordPolicy to be non-nil")
	} else {
		if settings.PasswordPolicy.MinLength != 12 {
			t.Errorf("Expected PasswordPolicy.MinLength to be 12, got %d", settings.PasswordPolicy.MinLength)
		}
		if !settings.PasswordPolicy.RequiresLowercase {
			t.Errorf("Expected PasswordPolicy.RequiresLowercase to be true")
		}
		if !settings.PasswordPolicy.RequiresUppercase {
			t.Errorf("Expected PasswordPolicy.RequiresUppercase to be true")
		}
		if !settings.PasswordPolicy.RequiresNumber {
			t.Errorf("Expected PasswordPolicy.RequiresNumber to be true")
		}
		if !settings.PasswordPolicy.RequiresSpecialChar {
			t.Errorf("Expected PasswordPolicy.RequiresSpecialChar to be true")
		}
		if settings.PasswordPolicy.ExpirationDays != 90 {
			t.Errorf("Expected PasswordPolicy.ExpirationDays to be 90, got %d", settings.PasswordPolicy.ExpirationDays)
		}
		if settings.PasswordPolicy.MaxHistory != 5 {
			t.Errorf("Expected PasswordPolicy.MaxHistory to be 5, got %d", settings.PasswordPolicy.MaxHistory)
		}
	}

	if !settings.SystemInsightsEnabled {
		t.Errorf("Expected SystemInsightsEnabled to be true")
	}

	if !settings.NewSystemUserStateManaged {
		t.Errorf("Expected NewSystemUserStateManaged to be true")
	}

	if settings.NewUserEmailTemplate != "Welcome!" {
		t.Errorf("Expected NewUserEmailTemplate to be 'Welcome!', got '%s'", settings.NewUserEmailTemplate)
	}

	if settings.PasswordResetTemplate != "Reset your password" {
		t.Errorf("Expected PasswordResetTemplate to be 'Reset your password', got '%s'", settings.PasswordResetTemplate)
	}

	if !settings.DirectoryInsightsEnabled {
		t.Errorf("Expected DirectoryInsightsEnabled to be true")
	}

	if settings.LdapIntegrationEnabled {
		t.Errorf("Expected LdapIntegrationEnabled to be false")
	}

	if !settings.AllowPublicKeyAuthentication {
		t.Errorf("Expected AllowPublicKeyAuthentication to be true")
	}

	if !settings.AllowMultiFactorAuth {
		t.Errorf("Expected AllowMultiFactorAuth to be true")
	}

	if settings.RequireMfa {
		t.Errorf("Expected RequireMfa to be false")
	}

	if len(settings.AllowedMfaMethods) != 2 {
		t.Errorf("Expected AllowedMfaMethods to have 2 items, got %d", len(settings.AllowedMfaMethods))
	} else {
		if settings.AllowedMfaMethods[0] != "totp" {
			t.Errorf("Expected AllowedMfaMethods[0] to be 'totp', got '%s'", settings.AllowedMfaMethods[0])
		}
		if settings.AllowedMfaMethods[1] != "push" {
			t.Errorf("Expected AllowedMfaMethods[1] to be 'push', got '%s'", settings.AllowedMfaMethods[1])
		}
	}

	if settings.Created != "2023-01-01T00:00:00Z" {
		t.Errorf("Expected Created to be '2023-01-01T00:00:00Z', got '%s'", settings.Created)
	}

	if settings.Updated != "2023-01-01T01:00:00Z" {
		t.Errorf("Expected Updated to be '2023-01-01T01:00:00Z', got '%s'", settings.Updated)
	}
}

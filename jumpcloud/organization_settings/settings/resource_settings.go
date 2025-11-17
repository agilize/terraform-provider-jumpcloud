package organization

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"

	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// ResourceSettings returns the resource schema for JumpCloud organization settings
func ResourceSettings() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceOrganizationSettingsCreate,
		ReadContext:   resourceOrganizationSettingsRead,
		UpdateContext: resourceOrganizationSettingsUpdate,
		DeleteContext: resourceOrganizationSettingsDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"password_policy": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"min_length": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      8,
							ValidateFunc: validation.IntBetween(8, 64),
						},
						"requires_lowercase": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"requires_uppercase": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"requires_number": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"requires_special_char": {
							Type:     schema.TypeBool,
							Optional: true,
							Default:  true,
						},
						"expiration_days": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      90,
							ValidateFunc: validation.IntBetween(0, 365),
						},
						"max_history": {
							Type:         schema.TypeInt,
							Optional:     true,
							Default:      5,
							ValidateFunc: validation.IntBetween(0, 24),
						},
					},
				},
			},
			"system_insights_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"new_system_user_state_managed": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"new_user_email_template": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"password_reset_template": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"directory_insights_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"ldap_integration_enabled": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"allow_public_key_authentication": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"allow_multi_factor_auth": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  true,
			},
			"require_mfa": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"allowed_mfa_methods": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceOrganizationSettingsCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Build the organization settings object
	orgSettings := buildOrganizationSettingsStruct(d)

	// Create the organization settings
	response, err := createOrganizationSettings(client, orgSettings)
	if err != nil {
		return diag.FromErr(err)
	}

	// Set the ID
	d.SetId(response.ID)

	// Read the organization settings to set the state
	return resourceOrganizationSettingsRead(ctx, d, meta)
}

func resourceOrganizationSettingsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Get the organization settings
	orgSettings, err := getOrganizationSettings(client, d.Id())
	if err != nil {
		// Check if it's a 404 error
		if common.IsNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Set the resource data
	if err := setOrganizationSettingsResourceData(d, orgSettings); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceOrganizationSettingsUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Build the organization settings object
	orgSettings := buildOrganizationSettingsStruct(d)

	// Update the organization settings
	if _, err := updateOrganizationSettings(client, d.Id(), orgSettings); err != nil {
		return diag.FromErr(err)
	}

	// Read the organization settings to set the state
	return resourceOrganizationSettingsRead(ctx, d, meta)
}

func resourceOrganizationSettingsDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	// Organization settings cannot be deleted, only reset to defaults
	d.SetId("")
	return nil
}

// Helper functions
func buildOrganizationSettingsStruct(d *schema.ResourceData) *OrganizationSettings {
	orgSettings := &OrganizationSettings{
		OrgID:                        d.Get("org_id").(string),
		SystemInsightsEnabled:        d.Get("system_insights_enabled").(bool),
		NewSystemUserStateManaged:    d.Get("new_system_user_state_managed").(bool),
		NewUserEmailTemplate:         d.Get("new_user_email_template").(string),
		PasswordResetTemplate:        d.Get("password_reset_template").(string),
		DirectoryInsightsEnabled:     d.Get("directory_insights_enabled").(bool),
		LdapIntegrationEnabled:       d.Get("ldap_integration_enabled").(bool),
		AllowPublicKeyAuthentication: d.Get("allow_public_key_authentication").(bool),
		AllowMultiFactorAuth:         d.Get("allow_multi_factor_auth").(bool),
		RequireMfa:                   d.Get("require_mfa").(bool),
	}

	// Convert password policy
	if v, ok := d.GetOk("password_policy"); ok && len(v.([]interface{})) > 0 {
		policyMap := v.([]interface{})[0].(map[string]interface{})
		orgSettings.PasswordPolicy = &PasswordPolicy{
			MinLength:           policyMap["min_length"].(int),
			RequiresLowercase:   policyMap["requires_lowercase"].(bool),
			RequiresUppercase:   policyMap["requires_uppercase"].(bool),
			RequiresNumber:      policyMap["requires_number"].(bool),
			RequiresSpecialChar: policyMap["requires_special_char"].(bool),
			ExpirationDays:      policyMap["expiration_days"].(int),
			MaxHistory:          policyMap["max_history"].(int),
		}
	}

	// Convert allowed MFA methods list
	if v, ok := d.GetOk("allowed_mfa_methods"); ok {
		methods := make([]string, 0)
		for _, method := range v.([]interface{}) {
			methods = append(methods, method.(string))
		}
		orgSettings.AllowedMfaMethods = methods
	}

	return orgSettings
}

func setOrganizationSettingsResourceData(d *schema.ResourceData, orgSettings *OrganizationSettings) error {
	if err := d.Set("org_id", orgSettings.OrgID); err != nil {
		return err
	}
	if err := d.Set("system_insights_enabled", orgSettings.SystemInsightsEnabled); err != nil {
		return err
	}
	if err := d.Set("new_system_user_state_managed", orgSettings.NewSystemUserStateManaged); err != nil {
		return err
	}
	if err := d.Set("new_user_email_template", orgSettings.NewUserEmailTemplate); err != nil {
		return err
	}
	if err := d.Set("password_reset_template", orgSettings.PasswordResetTemplate); err != nil {
		return err
	}
	if err := d.Set("directory_insights_enabled", orgSettings.DirectoryInsightsEnabled); err != nil {
		return err
	}
	if err := d.Set("ldap_integration_enabled", orgSettings.LdapIntegrationEnabled); err != nil {
		return err
	}
	if err := d.Set("allow_public_key_authentication", orgSettings.AllowPublicKeyAuthentication); err != nil {
		return err
	}
	if err := d.Set("allow_multi_factor_auth", orgSettings.AllowMultiFactorAuth); err != nil {
		return err
	}
	if err := d.Set("require_mfa", orgSettings.RequireMfa); err != nil {
		return err
	}
	if err := d.Set("allowed_mfa_methods", orgSettings.AllowedMfaMethods); err != nil {
		return err
	}
	if err := d.Set("created", orgSettings.Created); err != nil {
		return err
	}
	if err := d.Set("updated", orgSettings.Updated); err != nil {
		return err
	}

	// Set password policy
	if orgSettings.PasswordPolicy != nil {
		passwordPolicy := []map[string]interface{}{
			{
				"min_length":            orgSettings.PasswordPolicy.MinLength,
				"requires_lowercase":    orgSettings.PasswordPolicy.RequiresLowercase,
				"requires_uppercase":    orgSettings.PasswordPolicy.RequiresUppercase,
				"requires_number":       orgSettings.PasswordPolicy.RequiresNumber,
				"requires_special_char": orgSettings.PasswordPolicy.RequiresSpecialChar,
				"expiration_days":       orgSettings.PasswordPolicy.ExpirationDays,
				"max_history":           orgSettings.PasswordPolicy.MaxHistory,
			},
		}
		if err := d.Set("password_policy", passwordPolicy); err != nil {
			return err
		}
	}

	return nil
}

// API functions for organization settings
func createOrganizationSettings(client interface{}, orgSettings *OrganizationSettings) (*OrganizationSettings, error) {
	// Implementation depends on the actual JumpCloud API
	apiClient, err := common.ConvertToClientInterface(client)
	if err != nil {
		return nil, fmt.Errorf("error converting client: %w", err)
	}

	// Convert our internal struct to JSON
	body := map[string]interface{}{
		"orgId":                        orgSettings.OrgID,
		"systemInsightsEnabled":        orgSettings.SystemInsightsEnabled,
		"newSystemUserStateManaged":    orgSettings.NewSystemUserStateManaged,
		"newUserEmailTemplate":         orgSettings.NewUserEmailTemplate,
		"passwordResetTemplate":        orgSettings.PasswordResetTemplate,
		"directoryInsightsEnabled":     orgSettings.DirectoryInsightsEnabled,
		"ldapIntegrationEnabled":       orgSettings.LdapIntegrationEnabled,
		"allowPublicKeyAuthentication": orgSettings.AllowPublicKeyAuthentication,
		"allowMultiFactorAuth":         orgSettings.AllowMultiFactorAuth,
		"requireMfa":                   orgSettings.RequireMfa,
		"allowedMfaMethods":            orgSettings.AllowedMfaMethods,
	}

	// Add password policy if defined
	if orgSettings.PasswordPolicy != nil {
		body["passwordPolicy"] = map[string]interface{}{
			"minLength":           orgSettings.PasswordPolicy.MinLength,
			"requiresLowercase":   orgSettings.PasswordPolicy.RequiresLowercase,
			"requiresUppercase":   orgSettings.PasswordPolicy.RequiresUppercase,
			"requiresNumber":      orgSettings.PasswordPolicy.RequiresNumber,
			"requiresSpecialChar": orgSettings.PasswordPolicy.RequiresSpecialChar,
			"expirationDays":      orgSettings.PasswordPolicy.ExpirationDays,
			"maxHistory":          orgSettings.PasswordPolicy.MaxHistory,
		}
	}

	// Call the API
	resp, err := apiClient.DoRequest("POST", "/api/v2/organization-settings", body)
	if err != nil {
		return nil, fmt.Errorf("error creating organization settings: %w", err)
	}

	// Parse the response
	var result OrganizationSettings
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("error parsing organization settings response: %w", err)
	}

	return &result, nil
}

func getOrganizationSettings(client interface{}, id string) (*OrganizationSettings, error) {
	// Implementation depends on the actual JumpCloud API
	apiClient, err := common.ConvertToClientInterface(client)
	if err != nil {
		return nil, fmt.Errorf("error converting client: %w", err)
	}

	// Call the API
	resp, err := apiClient.DoRequest("GET", fmt.Sprintf("/api/v2/organization-settings/%s", id), nil)
	if err != nil {
		return nil, fmt.Errorf("error getting organization settings: %w", err)
	}

	// Parse the response
	var result OrganizationSettings
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("error parsing organization settings response: %w", err)
	}

	return &result, nil
}

func updateOrganizationSettings(client interface{}, id string, orgSettings *OrganizationSettings) (*OrganizationSettings, error) {
	// Implementation depends on the actual JumpCloud API
	apiClient, err := common.ConvertToClientInterface(client)
	if err != nil {
		return nil, fmt.Errorf("error converting client: %w", err)
	}

	// Convert our internal struct to JSON
	body := map[string]interface{}{
		"orgId":                        orgSettings.OrgID,
		"systemInsightsEnabled":        orgSettings.SystemInsightsEnabled,
		"newSystemUserStateManaged":    orgSettings.NewSystemUserStateManaged,
		"newUserEmailTemplate":         orgSettings.NewUserEmailTemplate,
		"passwordResetTemplate":        orgSettings.PasswordResetTemplate,
		"directoryInsightsEnabled":     orgSettings.DirectoryInsightsEnabled,
		"ldapIntegrationEnabled":       orgSettings.LdapIntegrationEnabled,
		"allowPublicKeyAuthentication": orgSettings.AllowPublicKeyAuthentication,
		"allowMultiFactorAuth":         orgSettings.AllowMultiFactorAuth,
		"requireMfa":                   orgSettings.RequireMfa,
		"allowedMfaMethods":            orgSettings.AllowedMfaMethods,
	}

	// Add password policy if defined
	if orgSettings.PasswordPolicy != nil {
		body["passwordPolicy"] = map[string]interface{}{
			"minLength":           orgSettings.PasswordPolicy.MinLength,
			"requiresLowercase":   orgSettings.PasswordPolicy.RequiresLowercase,
			"requiresUppercase":   orgSettings.PasswordPolicy.RequiresUppercase,
			"requiresNumber":      orgSettings.PasswordPolicy.RequiresNumber,
			"requiresSpecialChar": orgSettings.PasswordPolicy.RequiresSpecialChar,
			"expirationDays":      orgSettings.PasswordPolicy.ExpirationDays,
			"maxHistory":          orgSettings.PasswordPolicy.MaxHistory,
		}
	}

	// Call the API
	resp, err := apiClient.DoRequest("PUT", fmt.Sprintf("/api/v2/organization-settings/%s", id), body)
	if err != nil {
		return nil, fmt.Errorf("error updating organization settings: %w", err)
	}

	// Parse the response
	var result OrganizationSettings
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, fmt.Errorf("error parsing organization settings response: %w", err)
	}

	return &result, nil
}

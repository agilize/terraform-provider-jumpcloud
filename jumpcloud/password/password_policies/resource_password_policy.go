package password_policies

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
	"registry.terraform.io/agilize/jumpcloud/pkg/apiclient"
)

// PasswordPolicy represents a password policy in JumpCloud
type PasswordPolicy struct {
	ID                        string   `json:"_id,omitempty"`
	Name                      string   `json:"name"`
	Description               string   `json:"description,omitempty"`
	Status                    string   `json:"status,omitempty"` // active, inactive
	MinLength                 int      `json:"minLength"`
	MaxLength                 int      `json:"maxLength,omitempty"`
	RequireUppercase          bool     `json:"requireUppercase"`
	RequireLowercase          bool     `json:"requireLowercase"`
	RequireNumber             bool     `json:"requireNumber"`
	RequireSymbol             bool     `json:"requireSymbol"`
	MinimumAge                int      `json:"minimumAge,omitempty"`            // in days
	ExpirationTime            int      `json:"expirationTime,omitempty"`        // in days
	ExpirationWarningTime     int      `json:"expirationWarningTime,omitempty"` // in days
	DisallowPreviousPasswords int      `json:"disallowPreviousPasswords,omitempty"`
	DisallowCommonPasswords   bool     `json:"disallowCommonPasswords"`
	DisallowUsername          bool     `json:"disallowUsername"`
	DisallowNameAndEmail      bool     `json:"disallowNameAndEmail"`
	DisallowPasswordsFromList []string `json:"disallowPasswordsFromList,omitempty"`
	Scope                     string   `json:"scope,omitempty"` // organization, system_group
	TargetResources           []string `json:"targetResources,omitempty"`
	OrgID                     string   `json:"orgId,omitempty"`
	Created                   string   `json:"created,omitempty"`
	Updated                   string   `json:"updated,omitempty"`
}

// ResourcePasswordPolicy returns the resource schema for JumpCloud password policies
func ResourcePasswordPolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePasswordPolicyCreate,
		ReadContext:   resourcePasswordPolicyRead,
		UpdateContext: resourcePasswordPolicyUpdate,
		DeleteContext: resourcePasswordPolicyDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the password policy",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the password policy",
			},
			"status": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "active",
				ValidateFunc: validation.StringInSlice([]string{"active", "inactive"}, false),
				Description:  "Status of the policy: active or inactive",
			},
			"min_length": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(8, 64),
				Description:  "Minimum password length",
			},
			"max_length": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      64,
				ValidateFunc: validation.IntBetween(8, 128),
				Description:  "Maximum password length",
			},
			"require_uppercase": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Require at least one uppercase letter",
			},
			"require_lowercase": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Require at least one lowercase letter",
			},
			"require_number": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Require at least one number",
			},
			"require_symbol": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Require at least one special character",
			},
			"minimum_age": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 365),
				Description:  "Minimum password age in days before it can be changed",
			},
			"expiration_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntBetween(0, 365),
				Description:  "Password expiration time in days (0 means never expires)",
			},
			"expiration_warning_time": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      7,
				ValidateFunc: validation.IntBetween(1, 30),
				Description:  "Days before expiration to start warning",
			},
			"disallow_previous_passwords": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      5,
				ValidateFunc: validation.IntBetween(0, 100),
				Description:  "Number of previous passwords disallowed from reuse",
			},
			"disallow_common_passwords": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Prevent the use of common passwords",
			},
			"disallow_username": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Prevent the use of username in password",
			},
			"disallow_name_and_email": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Prevent the use of first name, last name or email in password",
			},
			"disallow_passwords_from_list": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "List of specific passwords to disallow",
			},
			"scope": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "organization",
				ValidateFunc: validation.StringInSlice([]string{"organization", "system_group"}, false),
				Description:  "Scope of the policy: organization or system_group",
			},
			"target_resources": {
				Type:        schema.TypeList,
				Optional:    true,
				Elem:        &schema.Schema{Type: schema.TypeString},
				Description: "Target resources for the policy when scope is system_group",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the policy",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update date of the policy",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

// resourcePasswordPolicyCreate creates a new password policy
func resourcePasswordPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(apiclient.Client)
	if !ok {
		return diag.FromErr(fmt.Errorf("invalid client configuration"))
	}

	// Build password policy
	policy := buildPasswordPolicyFromSchema(d)

	// Validate target_resources when scope is system_group
	if policy.Scope == "system_group" && len(policy.TargetResources) == 0 {
		return diag.FromErr(fmt.Errorf("target_resources is required when scope is 'system_group'"))
	}

	// Serialize to JSON
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing password policy: %v", err))
	}

	// Create policy via API
	tflog.Debug(ctx, fmt.Sprintf("Creating password policy: %s", policy.Name))
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/passwordpolicies", policyJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating password policy: %v", err))
	}

	// Deserialize response
	var createdPolicy PasswordPolicy
	if err := json.Unmarshal(resp, &createdPolicy); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	if createdPolicy.ID == "" {
		return diag.FromErr(fmt.Errorf("password policy created without ID"))
	}

	d.SetId(createdPolicy.ID)
	return resourcePasswordPolicyRead(ctx, d, meta)
}

// resourcePasswordPolicyRead reads a password policy
func resourcePasswordPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, ok := meta.(apiclient.Client)
	if !ok {
		return diag.FromErr(fmt.Errorf("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("password policy ID not provided"))
	}

	// Get policy via API
	tflog.Debug(ctx, fmt.Sprintf("Reading password policy: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/password-policies/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Password policy %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading password policy: %v", err))
	}

	// Deserialize response
	var policy PasswordPolicy
	if err := json.Unmarshal(resp, &policy); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set values in state
	d.Set("name", policy.Name)
	d.Set("description", policy.Description)
	d.Set("status", policy.Status)
	d.Set("min_length", policy.MinLength)
	d.Set("max_length", policy.MaxLength)
	d.Set("require_uppercase", policy.RequireUppercase)
	d.Set("require_lowercase", policy.RequireLowercase)
	d.Set("require_number", policy.RequireNumber)
	d.Set("require_symbol", policy.RequireSymbol)
	d.Set("minimum_age", policy.MinimumAge)
	d.Set("expiration_time", policy.ExpirationTime)
	d.Set("expiration_warning_time", policy.ExpirationWarningTime)
	d.Set("disallow_previous_passwords", policy.DisallowPreviousPasswords)
	d.Set("disallow_common_passwords", policy.DisallowCommonPasswords)
	d.Set("disallow_username", policy.DisallowUsername)
	d.Set("disallow_name_and_email", policy.DisallowNameAndEmail)
	d.Set("scope", policy.Scope)
	d.Set("created", policy.Created)
	d.Set("updated", policy.Updated)

	if policy.DisallowPasswordsFromList != nil {
		d.Set("disallow_passwords_from_list", policy.DisallowPasswordsFromList)
	}

	if policy.TargetResources != nil {
		d.Set("target_resources", policy.TargetResources)
	}

	if policy.OrgID != "" {
		d.Set("org_id", policy.OrgID)
	}

	return diags
}

// resourcePasswordPolicyUpdate updates an existing password policy
func resourcePasswordPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(apiclient.Client)
	if !ok {
		return diag.FromErr(fmt.Errorf("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("password policy ID not provided"))
	}

	// Build updated password policy
	policy := buildPasswordPolicyFromSchema(d)
	policy.ID = id

	// Validate target_resources when scope is system_group
	if policy.Scope == "system_group" && len(policy.TargetResources) == 0 {
		return diag.FromErr(fmt.Errorf("target_resources is required when scope is 'system_group'"))
	}

	// Serialize to JSON
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing password policy: %v", err))
	}

	// Update password policy via API
	tflog.Debug(ctx, fmt.Sprintf("Updating password policy: %s", id))
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/passwordpolicies/%s", id), policyJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating password policy: %v", err))
	}

	return resourcePasswordPolicyRead(ctx, d, meta)
}

// resourcePasswordPolicyDelete deletes a password policy
func resourcePasswordPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, ok := meta.(apiclient.Client)
	if !ok {
		return diag.FromErr(fmt.Errorf("invalid client configuration"))
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("password policy ID not provided"))
	}

	// Delete policy via API
	tflog.Debug(ctx, fmt.Sprintf("Deleting password policy: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/password-policies/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			// If the policy doesn't exist, consider the delete successful
			tflog.Warn(ctx, fmt.Sprintf("Password policy %s not found, assuming already deleted", id))
			return diag.Diagnostics{}
		}
		return diag.FromErr(fmt.Errorf("error deleting password policy: %v", err))
	}

	// Remove from state
	d.SetId("")
	return diag.Diagnostics{}
}

// buildPasswordPolicyFromSchema constructs a PasswordPolicy from schema data
func buildPasswordPolicyFromSchema(d *schema.ResourceData) *PasswordPolicy {
	policy := &PasswordPolicy{
		Name:                      d.Get("name").(string),
		Status:                    d.Get("status").(string),
		MinLength:                 d.Get("min_length").(int),
		RequireUppercase:          d.Get("require_uppercase").(bool),
		RequireLowercase:          d.Get("require_lowercase").(bool),
		RequireNumber:             d.Get("require_number").(bool),
		RequireSymbol:             d.Get("require_symbol").(bool),
		MinimumAge:                d.Get("minimum_age").(int),
		ExpirationTime:            d.Get("expiration_time").(int),
		ExpirationWarningTime:     d.Get("expiration_warning_time").(int),
		DisallowPreviousPasswords: d.Get("disallow_previous_passwords").(int),
		DisallowCommonPasswords:   d.Get("disallow_common_passwords").(bool),
		DisallowUsername:          d.Get("disallow_username").(bool),
		DisallowNameAndEmail:      d.Get("disallow_name_and_email").(bool),
		Scope:                     d.Get("scope").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		policy.Description = v.(string)
	}

	if v, ok := d.GetOk("max_length"); ok {
		policy.MaxLength = v.(int)
	}

	if v, ok := d.GetOk("org_id"); ok {
		policy.OrgID = v.(string)
	}

	// Process lists
	if v, ok := d.GetOk("disallow_passwords_from_list"); ok {
		rawList := v.([]interface{})
		disallowList := make([]string, len(rawList))
		for i, item := range rawList {
			disallowList[i] = item.(string)
		}
		policy.DisallowPasswordsFromList = disallowList
	}

	if v, ok := d.GetOk("target_resources"); ok {
		rawList := v.([]interface{})
		targetResources := make([]string, len(rawList))
		for i, item := range rawList {
			targetResources[i] = item.(string)
		}
		policy.TargetResources = targetResources
	}

	return policy
}

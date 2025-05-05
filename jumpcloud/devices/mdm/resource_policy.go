package mdm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// MDMPolicy represents an MDM policy in JumpCloud
type MDMPolicy struct {
	ID          string            `json:"_id,omitempty"`
	OrgID       string            `json:"orgId,omitempty"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Platform    string            `json:"platform"`
	Settings    json.RawMessage   `json:"settings"`
	ScopeType   string            `json:"scopeType,omitempty"`
	ScopeIDs    []string          `json:"scopeIds,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Created     string            `json:"created,omitempty"`
	Updated     string            `json:"updated,omitempty"`
}

// ResourcePolicy returns the schema resource for MDM policy
func ResourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMDMPolicyCreate,
		ReadContext:   resourceMDMPolicyRead,
		UpdateContext: resourceMDMPolicyUpdate,
		DeleteContext: resourceMDMPolicyDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the MDM policy",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the MDM policy",
			},
			"platform": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Device platform (ios, android, windows, macos)",
				ValidateFunc: validation.StringInSlice([]string{"ios", "android", "windows", "macos"}, false),
			},
			"settings": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "JSON string of policy settings",
				DiffSuppressFunc: suppressEquivalentJSONDiffs,
			},
			"scope_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "all",
				Description:  "Scope type for the policy (all, group, device)",
				ValidateFunc: validation.StringInSlice([]string{"all", "group", "device"}, false),
			},
			"scope_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "IDs of groups or devices in the scope",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"tags": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Tags to associate with the policy",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the MDM policy",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of the last update to the MDM policy",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceMDMPolicyCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Build MDM policy
	policy := &MDMPolicy{
		Name:      d.Get("name").(string),
		Platform:  d.Get("platform").(string),
		ScopeType: d.Get("scope_type").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("org_id"); ok {
		policy.OrgID = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		policy.Description = v.(string)
	}

	// Process settings (JSON string)
	if v, ok := d.GetOk("settings"); ok {
		settingsJSON := []byte(v.(string))
		// Validate JSON
		var settingsMap map[string]interface{}
		if err := json.Unmarshal(settingsJSON, &settingsMap); err != nil {
			return diag.FromErr(fmt.Errorf("invalid JSON in settings: %v", err))
		}
		policy.Settings = settingsJSON
	}

	// Process scope IDs
	if v, ok := d.GetOk("scope_ids"); ok {
		for _, sid := range v.([]interface{}) {
			policy.ScopeIDs = append(policy.ScopeIDs, sid.(string))
		}
	}

	// Process tags
	if v, ok := d.GetOk("tags"); ok {
		tags := make(map[string]string)
		for key, value := range v.(map[string]interface{}) {
			tags[key] = value.(string)
		}
		policy.Tags = tags
	}

	// Serialize to JSON
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing MDM policy: %v", err))
	}

	// Create policy via API
	tflog.Debug(ctx, "Creating MDM policy")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/mdm/policies", policyJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating MDM policy: %v", err))
	}

	// Deserialize response
	var createdPolicy MDMPolicy
	if err := json.Unmarshal(resp, &createdPolicy); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	if createdPolicy.ID == "" {
		return diag.FromErr(fmt.Errorf("MDM policy created without ID"))
	}

	d.SetId(createdPolicy.ID)
	return resourceMDMPolicyRead(ctx, d, meta)
}

func resourceMDMPolicyRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("MDM policy ID not provided"))
	}

	// Fetch policy via API
	tflog.Debug(ctx, fmt.Sprintf("Reading MDM policy with ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mdm/policies/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("MDM policy %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading MDM policy: %v", err))
	}

	// Deserialize response
	var policy MDMPolicy
	if err := json.Unmarshal(resp, &policy); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set values in state
	d.Set("name", policy.Name)
	d.Set("description", policy.Description)
	d.Set("platform", policy.Platform)

	// Format settings as JSON string
	if policy.Settings != nil {
		settingsStr, err := normalizeJSONString(string(policy.Settings))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error normalizing settings JSON: %v", err))
		}
		d.Set("settings", settingsStr)
	}

	d.Set("scope_type", policy.ScopeType)
	d.Set("created", policy.Created)
	d.Set("updated", policy.Updated)

	if policy.OrgID != "" {
		d.Set("org_id", policy.OrgID)
	}

	if len(policy.ScopeIDs) > 0 {
		d.Set("scope_ids", policy.ScopeIDs)
	}

	if len(policy.Tags) > 0 {
		d.Set("tags", policy.Tags)
	}

	return diags
}

func resourceMDMPolicyUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("MDM policy ID not provided"))
	}

	// Build MDM policy with current values
	policy := &MDMPolicy{
		ID:        id,
		Name:      d.Get("name").(string),
		Platform:  d.Get("platform").(string),
		ScopeType: d.Get("scope_type").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("org_id"); ok {
		policy.OrgID = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		policy.Description = v.(string)
	}

	// Process settings (JSON string)
	if v, ok := d.GetOk("settings"); ok {
		settingsJSON := []byte(v.(string))
		// Validate JSON
		var settingsMap map[string]interface{}
		if err := json.Unmarshal(settingsJSON, &settingsMap); err != nil {
			return diag.FromErr(fmt.Errorf("invalid JSON in settings: %v", err))
		}
		policy.Settings = settingsJSON
	}

	// Process scope IDs
	if v, ok := d.GetOk("scope_ids"); ok {
		for _, sid := range v.([]interface{}) {
			policy.ScopeIDs = append(policy.ScopeIDs, sid.(string))
		}
	}

	// Process tags
	if v, ok := d.GetOk("tags"); ok {
		tags := make(map[string]string)
		for key, value := range v.(map[string]interface{}) {
			tags[key] = value.(string)
		}
		policy.Tags = tags
	}

	// Serialize to JSON
	policyJSON, err := json.Marshal(policy)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing MDM policy: %v", err))
	}

	// Update policy via API
	tflog.Debug(ctx, fmt.Sprintf("Updating MDM policy: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/mdm/policies/%s", id), policyJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating MDM policy: %v", err))
	}

	// Deserialize response
	var updatedPolicy MDMPolicy
	if err := json.Unmarshal(resp, &updatedPolicy); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	return resourceMDMPolicyRead(ctx, d, meta)
}

func resourceMDMPolicyDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("MDM policy ID not provided"))
	}

	// Delete policy via API
	tflog.Debug(ctx, fmt.Sprintf("Deleting MDM policy: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/mdm/policies/%s", id), nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("MDM policy %s not found, considering deleted", id))
		} else {
			return diag.FromErr(fmt.Errorf("error deleting MDM policy: %v", err))
		}
	}

	d.SetId("")
	return diags
}

// Helper function to suppress JSON differences that are semantically equivalent
func suppressEquivalentJSONDiffs(k, old, new string, d *schema.ResourceData) bool {
	oldNormalized, err := normalizeJSONString(old)
	if err != nil {
		return false
	}
	newNormalized, err := normalizeJSONString(new)
	if err != nil {
		return false
	}
	return oldNormalized == newNormalized
}

// Helper function to normalize a JSON string
func normalizeJSONString(jsonStr string) (string, error) {
	var data interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", err
	}
	normalizedBytes, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return string(normalizedBytes), nil
}

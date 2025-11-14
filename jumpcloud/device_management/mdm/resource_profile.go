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
)

// MDMProfile represents an MDM profile in JumpCloud
type MDMProfile struct {
	ID          string            `json:"_id,omitempty"`
	OrgID       string            `json:"orgId,omitempty"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Platform    string            `json:"platform"`
	PayloadType string            `json:"payloadType"`
	Payload     json.RawMessage   `json:"payload"`
	ScopeType   string            `json:"scopeType,omitempty"`
	ScopeIDs    []string          `json:"scopeIds,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
	Created     string            `json:"created,omitempty"`
	Updated     string            `json:"updated,omitempty"`
}

// ResourceProfile returns the schema resource for MDM profile
func ResourceProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMDMProfileCreate,
		ReadContext:   resourceMDMProfileRead,
		UpdateContext: resourceMDMProfileUpdate,
		DeleteContext: resourceMDMProfileDelete,
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
				Description: "Name of the MDM profile",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the MDM profile",
			},
			"platform": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Device platform (ios, android, windows, macos)",
				ValidateFunc: validation.StringInSlice([]string{"ios", "android", "windows", "macos"}, false),
			},
			"payload_type": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Type of profile payload (configuration, certificate, etc.)",
				ValidateFunc: validation.StringInSlice([]string{"configuration", "certificate", "vpn", "wifi", "mail", "app", "custom"}, false),
			},
			"payload": {
				Type:             schema.TypeString,
				Required:         true,
				Description:      "JSON string of profile payload",
				DiffSuppressFunc: suppressEquivalentJSONDiffs,
			},
			"scope_type": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "all",
				Description:  "Scope type for the profile (all, group, device)",
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
				Description: "Tags to associate with the profile",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the MDM profile",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of the last update to the MDM profile",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceMDMProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	// Build MDM profile
	profile := &MDMProfile{
		Name:        d.Get("name").(string),
		Platform:    d.Get("platform").(string),
		PayloadType: d.Get("payload_type").(string),
		ScopeType:   d.Get("scope_type").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("org_id"); ok {
		profile.OrgID = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		profile.Description = v.(string)
	}

	// Process payload (JSON string)
	if v, ok := d.GetOk("payload"); ok {
		payloadJSON := []byte(v.(string))
		// Validate JSON
		var payloadMap map[string]interface{}
		if err := json.Unmarshal(payloadJSON, &payloadMap); err != nil {
			return diag.FromErr(fmt.Errorf("invalid JSON in payload: %v", err))
		}
		profile.Payload = payloadJSON
	}

	// Process scope IDs
	if v, ok := d.GetOk("scope_ids"); ok {
		for _, sid := range v.([]interface{}) {
			profile.ScopeIDs = append(profile.ScopeIDs, sid.(string))
		}
	}

	// Process tags
	if v, ok := d.GetOk("tags"); ok {
		tags := make(map[string]string)
		for key, value := range v.(map[string]interface{}) {
			tags[key] = value.(string)
		}
		profile.Tags = tags
	}

	// Serialize to JSON
	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing MDM profile: %v", err))
	}

	// Create profile via API
	tflog.Debug(ctx, "Creating MDM profile")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/mdm/profiles", profileJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating MDM profile: %v", err))
	}

	// Deserialize response
	var createdProfile MDMProfile
	if err := json.Unmarshal(resp, &createdProfile); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	if createdProfile.ID == "" {
		return diag.FromErr(fmt.Errorf("MDM profile created without ID"))
	}

	d.SetId(createdProfile.ID)
	return resourceMDMProfileRead(ctx, d, meta)
}

func resourceMDMProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("MDM profile ID not provided"))
	}

	// Fetch profile via API
	tflog.Debug(ctx, fmt.Sprintf("Reading MDM profile with ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mdm/profiles/%s", id), nil)
	if err != nil {
		if err.Error() == "404 Not Found" {
			tflog.Warn(ctx, fmt.Sprintf("MDM profile %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading MDM profile: %v", err))
	}

	// Deserialize response
	var profile MDMProfile
	if err := json.Unmarshal(resp, &profile); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set values in state
	if err := d.Set("name", profile.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %v", err))
	}

	if err := d.Set("description", profile.Description); err != nil {
		return diag.FromErr(fmt.Errorf("error setting description: %v", err))
	}

	if err := d.Set("platform", profile.Platform); err != nil {
		return diag.FromErr(fmt.Errorf("error setting platform: %v", err))
	}

	if err := d.Set("payload_type", profile.PayloadType); err != nil {
		return diag.FromErr(fmt.Errorf("error setting payload_type: %v", err))
	}

	// Format payload as JSON string
	if profile.Payload != nil {
		payloadStr, err := normalizeJSONString(string(profile.Payload))
		if err != nil {
			return diag.FromErr(fmt.Errorf("error normalizing payload JSON: %v", err))
		}
		if err := d.Set("payload", payloadStr); err != nil {
			return diag.FromErr(fmt.Errorf("error setting payload: %v", err))
		}
	}

	if err := d.Set("scope_type", profile.ScopeType); err != nil {
		return diag.FromErr(fmt.Errorf("error setting scope_type: %v", err))
	}

	if err := d.Set("created", profile.Created); err != nil {
		return diag.FromErr(fmt.Errorf("error setting created: %v", err))
	}

	if err := d.Set("updated", profile.Updated); err != nil {
		return diag.FromErr(fmt.Errorf("error setting updated: %v", err))
	}

	if profile.OrgID != "" {
		if err := d.Set("org_id", profile.OrgID); err != nil {
			return diag.FromErr(fmt.Errorf("error setting org_id: %v", err))
		}
	}

	if len(profile.ScopeIDs) > 0 {
		if err := d.Set("scope_ids", profile.ScopeIDs); err != nil {
			return diag.FromErr(fmt.Errorf("error setting scope_ids: %v", err))
		}
	}

	if len(profile.Tags) > 0 {
		if err := d.Set("tags", profile.Tags); err != nil {
			return diag.FromErr(fmt.Errorf("error setting tags: %v", err))
		}
	}

	return diags
}

func resourceMDMProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("MDM profile ID not provided"))
	}

	// Build MDM profile with current values
	profile := &MDMProfile{
		ID:          id,
		Name:        d.Get("name").(string),
		Platform:    d.Get("platform").(string),
		PayloadType: d.Get("payload_type").(string),
		ScopeType:   d.Get("scope_type").(string),
	}

	// Optional fields
	if v, ok := d.GetOk("org_id"); ok {
		profile.OrgID = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		profile.Description = v.(string)
	}

	// Process payload (JSON string)
	if v, ok := d.GetOk("payload"); ok {
		payloadJSON := []byte(v.(string))
		// Validate JSON
		var payloadMap map[string]interface{}
		if err := json.Unmarshal(payloadJSON, &payloadMap); err != nil {
			return diag.FromErr(fmt.Errorf("invalid JSON in payload: %v", err))
		}
		profile.Payload = payloadJSON
	}

	// Process scope IDs
	if v, ok := d.GetOk("scope_ids"); ok {
		for _, sid := range v.([]interface{}) {
			profile.ScopeIDs = append(profile.ScopeIDs, sid.(string))
		}
	}

	// Process tags
	if v, ok := d.GetOk("tags"); ok {
		tags := make(map[string]string)
		for key, value := range v.(map[string]interface{}) {
			tags[key] = value.(string)
		}
		profile.Tags = tags
	}

	// Serialize to JSON
	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing MDM profile: %v", err))
	}

	// Update profile via API
	tflog.Debug(ctx, fmt.Sprintf("Updating MDM profile: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/mdm/profiles/%s", id), profileJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating MDM profile: %v", err))
	}

	// Deserialize response
	var updatedProfile MDMProfile
	if err := json.Unmarshal(resp, &updatedProfile); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	return resourceMDMProfileRead(ctx, d, meta)
}

func resourceMDMProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("MDM profile ID not provided"))
	}

	// Delete profile via API
	tflog.Debug(ctx, fmt.Sprintf("Deleting MDM profile: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/mdm/profiles/%s", id), nil)
	if err != nil {
		if err.Error() == "404 Not Found" {
			tflog.Warn(ctx, fmt.Sprintf("MDM profile %s not found, considering deleted", id))
		} else {
			return diag.FromErr(fmt.Errorf("error deleting MDM profile: %v", err))
		}
	}

	d.SetId("")
	return diags
}

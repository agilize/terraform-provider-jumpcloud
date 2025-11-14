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

// MDMEnrollmentProfile represents an enrollment profile for MDM in JumpCloud
type MDMEnrollmentProfile struct {
	ID                 string   `json:"_id,omitempty"`
	OrgID              string   `json:"orgId,omitempty"`
	Name               string   `json:"name"`
	Description        string   `json:"description,omitempty"`
	Platform           string   `json:"platform"`
	EnrollmentMethod   string   `json:"enrollmentMethod"`
	GroupID            string   `json:"groupId,omitempty"`
	GroupIDs           []string `json:"groupIds,omitempty"`
	AllowByod          bool     `json:"allowByod"`
	RequirePasscode    bool     `json:"requirePasscode"`
	UserAuthentication bool     `json:"userAuthentication"`
	Created            string   `json:"created,omitempty"`
	Updated            string   `json:"updated,omitempty"`
}

// ResourceEnrollmentProfile returns the schema resource for MDM enrollment profile
func ResourceEnrollmentProfile() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMDMEnrollmentProfileCreate,
		ReadContext:   resourceMDMEnrollmentProfileRead,
		UpdateContext: resourceMDMEnrollmentProfileUpdate,
		DeleteContext: resourceMDMEnrollmentProfileDelete,
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
				Description: "Name of the enrollment profile",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the enrollment profile",
			},
			"platform": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Device platform (ios, android, windows, macos)",
				ValidateFunc: validation.StringInSlice([]string{"ios", "android", "windows", "macos"}, false),
			},
			"enrollment_method": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Enrollment method (user_initiated, admin_initiated, dep)",
				ValidateFunc: validation.StringInSlice([]string{"user_initiated", "admin_initiated", "dep"}, false),
			},
			"group_id": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "User group ID for enrollment (single group)",
				ConflictsWith: []string{"group_ids"},
			},
			"group_ids": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "User group IDs for enrollment (multiple groups)",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				ConflictsWith: []string{"group_id"},
			},
			"allow_byod": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to allow personal devices (Bring Your Own Device)",
			},
			"require_passcode": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to require a passcode on enrolled devices",
			},
			"user_authentication": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether to require user authentication during enrollment",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the enrollment profile",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of the last update to the enrollment profile",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceMDMEnrollmentProfileCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	// Build enrollment profile
	profile := &MDMEnrollmentProfile{
		Name:               d.Get("name").(string),
		Platform:           d.Get("platform").(string),
		EnrollmentMethod:   d.Get("enrollment_method").(string),
		AllowByod:          d.Get("allow_byod").(bool),
		RequirePasscode:    d.Get("require_passcode").(bool),
		UserAuthentication: d.Get("user_authentication").(bool),
	}

	// Optional fields
	if v, ok := d.GetOk("org_id"); ok {
		profile.OrgID = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		profile.Description = v.(string)
	}

	if v, ok := d.GetOk("group_id"); ok {
		profile.GroupID = v.(string)
	}

	if v, ok := d.GetOk("group_ids"); ok {
		for _, gid := range v.([]interface{}) {
			profile.GroupIDs = append(profile.GroupIDs, gid.(string))
		}
	}

	// Serialize to JSON
	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing enrollment profile: %v", err))
	}

	// Create enrollment profile via API
	tflog.Debug(ctx, "Creating MDM enrollment profile")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/mdm/enrollmentprofiles", profileJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating enrollment profile: %v", err))
	}

	// Deserialize response
	var createdProfile MDMEnrollmentProfile
	if err := json.Unmarshal(resp, &createdProfile); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	if createdProfile.ID == "" {
		return diag.FromErr(fmt.Errorf("enrollment profile created without ID"))
	}

	d.SetId(createdProfile.ID)
	return resourceMDMEnrollmentProfileRead(ctx, d, meta)
}

func resourceMDMEnrollmentProfileRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("enrollment profile ID not provided"))
	}

	// Fetch enrollment profile via API
	tflog.Debug(ctx, fmt.Sprintf("Reading MDM enrollment profile with ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mdm/enrollmentprofiles/%s", id), nil)
	if err != nil {
		if err.Error() == "404 Not Found" {
			tflog.Warn(ctx, fmt.Sprintf("MDM enrollment profile %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading enrollment profile: %v", err))
	}

	// Deserialize response
	var profile MDMEnrollmentProfile
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

	if err := d.Set("enrollment_method", profile.EnrollmentMethod); err != nil {
		return diag.FromErr(fmt.Errorf("error setting enrollment_method: %v", err))
	}

	if err := d.Set("allow_byod", profile.AllowByod); err != nil {
		return diag.FromErr(fmt.Errorf("error setting allow_byod: %v", err))
	}

	if err := d.Set("require_passcode", profile.RequirePasscode); err != nil {
		return diag.FromErr(fmt.Errorf("error setting require_passcode: %v", err))
	}

	if err := d.Set("user_authentication", profile.UserAuthentication); err != nil {
		return diag.FromErr(fmt.Errorf("error setting user_authentication: %v", err))
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

	if profile.GroupID != "" {
		if err := d.Set("group_id", profile.GroupID); err != nil {
			return diag.FromErr(fmt.Errorf("error setting group_id: %v", err))
		}
	}

	if len(profile.GroupIDs) > 0 {
		if err := d.Set("group_ids", profile.GroupIDs); err != nil {
			return diag.FromErr(fmt.Errorf("error setting group_ids: %v", err))
		}
	}

	return diags
}

func resourceMDMEnrollmentProfileUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("enrollment profile ID not provided"))
	}

	// Build enrollment profile with current values
	profile := &MDMEnrollmentProfile{
		ID:                 id,
		Name:               d.Get("name").(string),
		Platform:           d.Get("platform").(string),
		EnrollmentMethod:   d.Get("enrollment_method").(string),
		AllowByod:          d.Get("allow_byod").(bool),
		RequirePasscode:    d.Get("require_passcode").(bool),
		UserAuthentication: d.Get("user_authentication").(bool),
	}

	// Optional fields
	if v, ok := d.GetOk("org_id"); ok {
		profile.OrgID = v.(string)
	}

	if v, ok := d.GetOk("description"); ok {
		profile.Description = v.(string)
	}

	if v, ok := d.GetOk("group_id"); ok {
		profile.GroupID = v.(string)
	}

	if v, ok := d.GetOk("group_ids"); ok {
		for _, gid := range v.([]interface{}) {
			profile.GroupIDs = append(profile.GroupIDs, gid.(string))
		}
	}

	// Serialize to JSON
	profileJSON, err := json.Marshal(profile)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing enrollment profile: %v", err))
	}

	// Update enrollment profile via API
	tflog.Debug(ctx, fmt.Sprintf("Updating MDM enrollment profile: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/mdm/enrollmentprofiles/%s", id), profileJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating enrollment profile: %v", err))
	}

	// Deserialize response
	var updatedProfile MDMEnrollmentProfile
	if err := json.Unmarshal(resp, &updatedProfile); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	return resourceMDMEnrollmentProfileRead(ctx, d, meta)
}

func resourceMDMEnrollmentProfileDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("enrollment profile ID not provided"))
	}

	// Delete enrollment profile via API
	tflog.Debug(ctx, fmt.Sprintf("Deleting MDM enrollment profile: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/mdm/enrollmentprofiles/%s", id), nil)
	if err != nil {
		if err.Error() == "404 Not Found" {
			tflog.Warn(ctx, fmt.Sprintf("MDM enrollment profile %s not found, considering deleted", id))
		} else {
			return diag.FromErr(fmt.Errorf("error deleting enrollment profile: %v", err))
		}
	}

	d.SetId("")
	return diags
}

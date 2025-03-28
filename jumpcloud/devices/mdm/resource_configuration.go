package mdm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// MDMConfiguration represents an MDM configuration in JumpCloud
type MDMConfiguration struct {
	ID                       string `json:"_id,omitempty"`
	OrgID                    string `json:"orgId,omitempty"`
	Enabled                  bool   `json:"enabled"`
	AppleEnabled             bool   `json:"appleEnabled"`
	AndroidEnabled           bool   `json:"androidEnabled"`
	WindowsEnabled           bool   `json:"windowsEnabled"`
	AppleMDMServerURL        string `json:"appleMdmServerUrl,omitempty"`
	AppleMDMPushCertificate  string `json:"appleMdmPushCertificate,omitempty"`
	AppleMDMTokenExpiresAt   string `json:"appleMdmTokenExpiresAt,omitempty"`
	AndroidEnterpriseEnabled bool   `json:"androidEnterpriseEnabled"`
	AndroidPlayStoreID       string `json:"androidPlayStoreId,omitempty"`
	DefaultAppCatalogEnabled bool   `json:"defaultAppCatalogEnabled"`
	AutoEnrollmentEnabled    bool   `json:"autoEnrollmentEnabled"`
	Created                  string `json:"created,omitempty"`
	Updated                  string `json:"updated,omitempty"`
}

// ResourceConfiguration returns the schema resource for MDM configuration
func ResourceConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMDMConfigurationCreate,
		ReadContext:   resourceMDMConfigurationRead,
		UpdateContext: resourceMDMConfigurationUpdate,
		DeleteContext: resourceMDMConfigurationDelete,
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
			"enabled": {
				Type:        schema.TypeBool,
				Required:    true,
				Description: "Whether MDM is globally enabled",
			},
			"apple_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether MDM for Apple devices is enabled",
			},
			"android_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether MDM for Android devices is enabled",
			},
			"windows_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether MDM for Windows devices is enabled",
			},
			"apple_mdm_server_url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "MDM server URL for Apple devices",
			},
			"apple_mdm_push_certificate": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Push certificate for Apple MDM",
			},
			"apple_mdm_token_expires_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Expiration date of the Apple MDM token",
			},
			"android_enterprise_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether Android Enterprise is enabled",
			},
			"android_play_store_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Play Store ID for Android MDM",
			},
			"default_app_catalog_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the default app catalog is enabled",
			},
			"auto_enrollment_enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether auto-enrollment is enabled",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the MDM configuration",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of the last update to the MDM configuration",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceMDMConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	// Build MDM configuration
	config := &MDMConfiguration{
		Enabled:                  d.Get("enabled").(bool),
		AppleEnabled:             d.Get("apple_enabled").(bool),
		AndroidEnabled:           d.Get("android_enabled").(bool),
		WindowsEnabled:           d.Get("windows_enabled").(bool),
		AndroidEnterpriseEnabled: d.Get("android_enterprise_enabled").(bool),
		DefaultAppCatalogEnabled: d.Get("default_app_catalog_enabled").(bool),
		AutoEnrollmentEnabled:    d.Get("auto_enrollment_enabled").(bool),
	}

	// Optional fields
	if v, ok := d.GetOk("org_id"); ok {
		config.OrgID = v.(string)
	}

	if v, ok := d.GetOk("apple_mdm_push_certificate"); ok {
		config.AppleMDMPushCertificate = v.(string)
	}

	if v, ok := d.GetOk("android_play_store_id"); ok {
		config.AndroidPlayStoreID = v.(string)
	}

	// Serialize to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing MDM configuration: %v", err))
	}

	// Create configuration via API
	tflog.Debug(ctx, "Creating MDM configuration")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/mdm/config", configJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating MDM configuration: %v", err))
	}

	// Deserialize response
	var createdConfig MDMConfiguration
	if err := json.Unmarshal(resp, &createdConfig); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	if createdConfig.ID == "" {
		return diag.FromErr(fmt.Errorf("MDM configuration created without ID"))
	}

	d.SetId(createdConfig.ID)
	return resourceMDMConfigurationRead(ctx, d, meta)
}

func resourceMDMConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("MDM configuration ID not provided"))
	}

	// Fetch configuration via API
	tflog.Debug(ctx, fmt.Sprintf("Reading MDM configuration with ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/mdm/config/%s", id), nil)
	if err != nil {
		if err.Error() == "404 Not Found" {
			tflog.Warn(ctx, fmt.Sprintf("MDM configuration %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading MDM configuration: %v", err))
	}

	// Deserialize response
	var config MDMConfiguration
	if err := json.Unmarshal(resp, &config); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set values in state
	d.Set("enabled", config.Enabled)
	d.Set("apple_enabled", config.AppleEnabled)
	d.Set("android_enabled", config.AndroidEnabled)
	d.Set("windows_enabled", config.WindowsEnabled)
	d.Set("apple_mdm_server_url", config.AppleMDMServerURL)
	d.Set("apple_mdm_token_expires_at", config.AppleMDMTokenExpiresAt)
	d.Set("android_enterprise_enabled", config.AndroidEnterpriseEnabled)
	d.Set("android_play_store_id", config.AndroidPlayStoreID)
	d.Set("default_app_catalog_enabled", config.DefaultAppCatalogEnabled)
	d.Set("auto_enrollment_enabled", config.AutoEnrollmentEnabled)
	d.Set("created", config.Created)
	d.Set("updated", config.Updated)

	// We don't set the Apple push certificate in the state as it's sensitive
	// and isn't returned in full by the API

	if config.OrgID != "" {
		d.Set("org_id", config.OrgID)
	}

	return diags
}

func resourceMDMConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("MDM configuration ID not provided"))
	}

	// Build MDM configuration with current values
	config := &MDMConfiguration{
		ID:                       id,
		Enabled:                  d.Get("enabled").(bool),
		AppleEnabled:             d.Get("apple_enabled").(bool),
		AndroidEnabled:           d.Get("android_enabled").(bool),
		WindowsEnabled:           d.Get("windows_enabled").(bool),
		AndroidEnterpriseEnabled: d.Get("android_enterprise_enabled").(bool),
		DefaultAppCatalogEnabled: d.Get("default_app_catalog_enabled").(bool),
		AutoEnrollmentEnabled:    d.Get("auto_enrollment_enabled").(bool),
	}

	// Optional fields
	if v, ok := d.GetOk("org_id"); ok {
		config.OrgID = v.(string)
	}

	// Only include the Apple MDM push certificate if it's changed
	if d.HasChange("apple_mdm_push_certificate") {
		if v, ok := d.GetOk("apple_mdm_push_certificate"); ok {
			config.AppleMDMPushCertificate = v.(string)
		}
	}

	if v, ok := d.GetOk("android_play_store_id"); ok {
		config.AndroidPlayStoreID = v.(string)
	}

	// Serialize to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing MDM configuration: %v", err))
	}

	// Update configuration via API
	tflog.Debug(ctx, fmt.Sprintf("Updating MDM configuration: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/mdm/config/%s", id), configJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating MDM configuration: %v", err))
	}

	// Deserialize response
	var updatedConfig MDMConfiguration
	if err := json.Unmarshal(resp, &updatedConfig); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	return resourceMDMConfigurationRead(ctx, d, meta)
}

func resourceMDMConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("MDM configuration ID not provided"))
	}

	// Delete configuration via API
	tflog.Debug(ctx, fmt.Sprintf("Deleting MDM configuration: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/mdm/config/%s", id), nil)
	if err != nil {
		if err.Error() == "404 Not Found" {
			tflog.Warn(ctx, fmt.Sprintf("MDM configuration %s not found, considering deleted", id))
		} else {
			return diag.FromErr(fmt.Errorf("error deleting MDM configuration: %v", err))
		}
	}

	d.SetId("")
	return diags
}

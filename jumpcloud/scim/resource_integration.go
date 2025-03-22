package scim

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
)

// ScimIntegration represents a SCIM integration in JumpCloud
type ScimIntegration struct {
	ID             string                 `json:"_id,omitempty"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	Type           string                 `json:"type"` // saas, identity_provider, custom
	ServerID       string                 `json:"serverId"`
	Status         string                 `json:"status,omitempty"` // active, pending, error
	Enabled        bool                   `json:"enabled"`
	Settings       map[string]interface{} `json:"settings,omitempty"`
	SyncSchedule   string                 `json:"syncSchedule,omitempty"` // manual, daily, hourly, etc
	SyncInterval   int                    `json:"syncInterval,omitempty"` // in minutes
	LastSyncTime   string                 `json:"lastSyncTime,omitempty"`
	LastSyncStatus string                 `json:"lastSyncStatus,omitempty"`
	MappingIDs     []string               `json:"mappingIds,omitempty"`
	OrgID          string                 `json:"orgId,omitempty"`
	Created        string                 `json:"created,omitempty"`
	Updated        string                 `json:"updated,omitempty"`
}

// ResourceIntegration returns a schema resource for managing SCIM integrations in JumpCloud
func ResourceIntegration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceIntegrationCreate,
		ReadContext:   resourceIntegrationRead,
		UpdateContext: resourceIntegrationUpdate,
		DeleteContext: resourceIntegrationDelete,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(1 * time.Minute),
			Read:   schema.DefaultTimeout(1 * time.Minute),
			Update: schema.DefaultTimeout(1 * time.Minute),
			Delete: schema.DefaultTimeout(1 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringLenBetween(1, 255),
				Description:  "Name of the SCIM integration",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the SCIM integration",
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				ValidateFunc: validation.StringInSlice([]string{
					"saas", "identity_provider", "custom",
				}, false),
				Description: "Type of SCIM integration (saas, identity_provider, custom)",
			},
			"server_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "ID of the SCIM server associated with the integration",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Indicates if the SCIM integration is enabled",
			},
			"settings": {
				Type:             schema.TypeString,
				Optional:         true,
				ValidateFunc:     validation.StringIsJSON,
				DiffSuppressFunc: common.SuppressEquivalentJSONDiffs,
				Description:      "Integration-specific settings in JSON format",
			},
			"sync_schedule": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "manual",
				ValidateFunc: validation.StringInSlice([]string{
					"manual", "hourly", "daily", "weekly", "custom",
				}, false),
				Description: "Synchronization schedule (manual, hourly, daily, weekly, custom)",
			},
			"sync_interval": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      1440, // 24 hours in minutes
				ValidateFunc: validation.IntAtLeast(15),
				Description:  "Synchronization interval in minutes (when sync_schedule is 'custom')",
			},
			"mapping_ids": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "IDs of associated attribute mappings",
			},
			"status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Current status of the integration (active, pending, error)",
			},
			"last_sync_time": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date and time of the last synchronization",
			},
			"last_sync_status": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Status of the last synchronization",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the integration",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Last update date of the integration",
			},
		},
	}
}

func resourceIntegrationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetJumpCloudClient(meta)
	if diagErr != nil {
		return diagErr
	}

	// Build ScimIntegration object from terraform data
	integration := &ScimIntegration{
		Name:         d.Get("name").(string),
		Type:         d.Get("type").(string),
		ServerID:     d.Get("server_id").(string),
		Enabled:      d.Get("enabled").(bool),
		SyncSchedule: d.Get("sync_schedule").(string),
		SyncInterval: d.Get("sync_interval").(int),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		integration.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		integration.OrgID = v.(string)
	}

	// Process settings (JSON)
	if v, ok := d.GetOk("settings"); ok {
		settingsJSON := v.(string)
		var settings map[string]interface{}
		if err := json.Unmarshal([]byte(settingsJSON), &settings); err != nil {
			return diag.FromErr(fmt.Errorf("error deserializing settings: %v", err))
		}
		integration.Settings = settings
	}

	// Process mapping IDs
	if v, ok := d.GetOk("mapping_ids"); ok {
		mappingsList := v.([]interface{})
		mappingIDs := make([]string, len(mappingsList))
		for i, id := range mappingsList {
			mappingIDs[i] = id.(string)
		}
		integration.MappingIDs = mappingIDs
	}

	// Serialize to JSON
	reqBody, err := json.Marshal(integration)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing SCIM integration: %v", err))
	}

	// Build URL for request
	url := fmt.Sprintf("/api/v2/scim/servers/%s/integrations", integration.ServerID)
	if integration.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, integration.OrgID)
	}

	// Make request to create integration
	tflog.Debug(ctx, fmt.Sprintf("Creating SCIM integration for server: %s", integration.ServerID))
	resp, err := c.DoRequest(http.MethodPost, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating SCIM integration: %v", err))
	}

	// Deserialize response
	var createdIntegration ScimIntegration
	if err := json.Unmarshal(resp, &createdIntegration); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	if createdIntegration.ID == "" {
		return diag.FromErr(fmt.Errorf("created SCIM integration returned without an ID"))
	}

	// Set ID in state
	d.SetId(createdIntegration.ID)

	// Read the resource to update state with all computed fields
	return resourceIntegrationRead(ctx, d, meta)
}

func resourceIntegrationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetJumpCloudClient(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get integration ID
	integrationID := d.Id()
	if integrationID == "" {
		return diag.FromErr(fmt.Errorf("SCIM integration ID is required"))
	}

	// Get server ID
	var serverID string
	if v, ok := d.GetOk("server_id"); ok {
		serverID = v.(string)
	} else {
		// If we don't have the server_id in state (possibly during import),
		// we need to fetch the integration by ID to discover the server_id
		url := fmt.Sprintf("/api/v2/scim/integrations/%s", integrationID)
		if v, ok := d.GetOk("org_id"); ok {
			url = fmt.Sprintf("%s?orgId=%s", url, v.(string))
		}

		resp, err := c.DoRequest(http.MethodGet, url, nil)
		if err != nil {
			if common.IsNotFoundError(err) {
				tflog.Warn(ctx, fmt.Sprintf("SCIM integration %s not found, removing from state", integrationID))
				d.SetId("")
				return diags
			}
			return diag.FromErr(fmt.Errorf("error fetching SCIM integration: %v", err))
		}

		var integration ScimIntegration
		if err := json.Unmarshal(resp, &integration); err != nil {
			return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
		}

		serverID = integration.ServerID
		d.Set("server_id", serverID)
	}

	// Get orgId parameter if available
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Build URL for request
	url := fmt.Sprintf("/api/v2/scim/servers/%s/integrations/%s%s", serverID, integrationID, orgIDParam)

	// Make request to get integration details
	tflog.Debug(ctx, fmt.Sprintf("Reading SCIM integration: %s", integrationID))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("SCIM integration %s not found, removing from state", integrationID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading SCIM integration: %v", err))
	}

	// Deserialize response
	var integration ScimIntegration
	if err := json.Unmarshal(resp, &integration); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing SCIM integration: %v", err))
	}

	// Set values in state
	d.Set("name", integration.Name)
	d.Set("description", integration.Description)
	d.Set("type", integration.Type)
	d.Set("server_id", integration.ServerID)
	d.Set("enabled", integration.Enabled)
	d.Set("sync_schedule", integration.SyncSchedule)
	d.Set("sync_interval", integration.SyncInterval)
	d.Set("status", integration.Status)
	d.Set("last_sync_time", integration.LastSyncTime)
	d.Set("last_sync_status", integration.LastSyncStatus)
	d.Set("created", integration.Created)
	d.Set("updated", integration.Updated)

	// Set mapping IDs
	if integration.MappingIDs != nil {
		if err := d.Set("mapping_ids", integration.MappingIDs); err != nil {
			return diag.FromErr(fmt.Errorf("error setting mapping_ids: %v", err))
		}
	}

	// Set settings - Convert settings map to JSON string
	if integration.Settings != nil {
		settingsJSON, err := json.Marshal(integration.Settings)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error serializing settings: %v", err))
		}
		d.Set("settings", string(settingsJSON))
	}

	// Set OrgID if present
	if integration.OrgID != "" {
		d.Set("org_id", integration.OrgID)
	}

	return diags
}

func resourceIntegrationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetJumpCloudClient(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get integration ID
	integrationID := d.Id()
	if integrationID == "" {
		return diag.FromErr(fmt.Errorf("SCIM integration ID is required"))
	}

	// Build ScimIntegration object from terraform data
	integration := &ScimIntegration{
		ID:           integrationID,
		Name:         d.Get("name").(string),
		Type:         d.Get("type").(string),
		ServerID:     d.Get("server_id").(string),
		Enabled:      d.Get("enabled").(bool),
		SyncSchedule: d.Get("sync_schedule").(string),
		SyncInterval: d.Get("sync_interval").(int),
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		integration.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		integration.OrgID = v.(string)
	}

	// Process settings (JSON)
	if v, ok := d.GetOk("settings"); ok {
		settingsJSON := v.(string)
		var settings map[string]interface{}
		if err := json.Unmarshal([]byte(settingsJSON), &settings); err != nil {
			return diag.FromErr(fmt.Errorf("error deserializing settings: %v", err))
		}
		integration.Settings = settings
	}

	// Process mapping IDs
	if v, ok := d.GetOk("mapping_ids"); ok {
		mappingsList := v.([]interface{})
		mappingIDs := make([]string, len(mappingsList))
		for i, id := range mappingsList {
			mappingIDs[i] = id.(string)
		}
		integration.MappingIDs = mappingIDs
	}

	// Serialize to JSON
	reqBody, err := json.Marshal(integration)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing SCIM integration: %v", err))
	}

	// Build URL for request
	url := fmt.Sprintf("/api/v2/scim/servers/%s/integrations/%s", integration.ServerID, integrationID)
	if integration.OrgID != "" {
		url = fmt.Sprintf("%s?orgId=%s", url, integration.OrgID)
	}

	// Make request to update integration
	tflog.Debug(ctx, fmt.Sprintf("Updating SCIM integration: %s", integrationID))
	_, err = c.DoRequest(http.MethodPut, url, reqBody)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating SCIM integration: %v", err))
	}

	// Read the resource to update state with all computed fields
	return resourceIntegrationRead(ctx, d, meta)
}

func resourceIntegrationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetJumpCloudClient(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get integration ID
	integrationID := d.Id()
	if integrationID == "" {
		return diag.FromErr(fmt.Errorf("SCIM integration ID is required"))
	}

	// Get server ID
	serverID := d.Get("server_id").(string)
	if serverID == "" {
		return diag.FromErr(fmt.Errorf("SCIM server ID is required"))
	}

	// Get orgId parameter if available
	var orgIDParam string
	if v, ok := d.GetOk("org_id"); ok {
		orgIDParam = fmt.Sprintf("?orgId=%s", v.(string))
	}

	// Build URL for request
	url := fmt.Sprintf("/api/v2/scim/servers/%s/integrations/%s%s", serverID, integrationID, orgIDParam)

	// Make request to delete integration
	tflog.Debug(ctx, fmt.Sprintf("Deleting SCIM integration: %s", integrationID))
	_, err := c.DoRequest(http.MethodDelete, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting SCIM integration: %v", err))
	}

	// Clear ID from state
	d.SetId("")

	return diags
}

package directory_insights

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

// ResourceConfiguration returns a schema resource for managing JumpCloud Directory Insights configuration
func ResourceConfiguration() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceConfigurationCreate,
		ReadContext:   resourceConfigurationRead,
		UpdateContext: resourceConfigurationUpdate,
		DeleteContext: resourceConfigurationDelete,

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
			"retention_days": {
				Type:         schema.TypeInt,
				Required:     true,
				ValidateFunc: validation.IntBetween(1, 90),
				Description:  "Number of days to retain events (1-90)",
			},
			"enabled_event_types": {
				Type:     schema.TypeList,
				Required: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of event types enabled for collection",
			},
			"export_to_cloudwatch": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether events should be exported to AWS CloudWatch",
			},
			"export_to_datadog": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     false,
				Description: "Whether events should be exported to Datadog",
			},
			"datadog_region": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Datadog region for event export",
			},
			"datadog_api_key": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				Description: "Datadog API key for event export",
			},
			"enabled_alerting_events": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of event types for which alerts will be sent",
			},
			"notification_emails": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Description: "List of email addresses to receive alert notifications",
			},
		},

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(30 * time.Second),
			Read:   schema.DefaultTimeout(30 * time.Second),
			Update: schema.DefaultTimeout(30 * time.Second),
			Delete: schema.DefaultTimeout(30 * time.Second),
		},
	}
}

// resourceConfigurationCreate creates a new Directory Insights configuration
func resourceConfigurationCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	// Build configuration
	config := &Config{
		RetentionDays:      d.Get("retention_days").(int),
		ExportToCloudWatch: d.Get("export_to_cloudwatch").(bool),
		ExportToDatadog:    d.Get("export_to_datadog").(bool),
	}

	// Optional fields
	if v, ok := d.GetOk("org_id"); ok {
		config.OrgID = v.(string)
	}

	if v, ok := d.GetOk("datadog_region"); ok {
		config.DatadogRegion = v.(string)
	}

	if v, ok := d.GetOk("datadog_api_key"); ok {
		config.DatadogAPIKey = v.(string)
	}

	// Process lists
	if v, ok := d.GetOk("enabled_event_types"); ok {
		eventTypes := v.([]interface{})
		config.EnabledEventTypes = make([]string, len(eventTypes))
		for i, eventType := range eventTypes {
			config.EnabledEventTypes[i] = eventType.(string)
		}
	}

	if v, ok := d.GetOk("enabled_alerting_events"); ok {
		alertingEvents := v.([]interface{})
		config.EnabledAlertingEvents = make([]string, len(alertingEvents))
		for i, event := range alertingEvents {
			config.EnabledAlertingEvents[i] = event.(string)
		}
	}

	if v, ok := d.GetOk("notification_emails"); ok {
		emails := v.([]interface{})
		config.NotificationEmails = make([]string, len(emails))
		for i, email := range emails {
			config.NotificationEmails[i] = email.(string)
		}
	}

	// Serialize to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing Directory Insights configuration: %v", err))
	}

	// Create configuration via API
	tflog.Debug(ctx, "Creating Directory Insights configuration")
	resp, err := client.DoRequest(http.MethodPost, "/insights/directory/v1/config", configJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating Directory Insights configuration: %v", err))
	}

	// Deserialize response
	var createdConfig Config
	if err := json.Unmarshal(resp, &createdConfig); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	if createdConfig.ID == "" {
		return diag.FromErr(fmt.Errorf("Directory Insights configuration created without an ID"))
	}

	d.SetId(createdConfig.ID)
	return resourceConfigurationRead(ctx, d, meta)
}

// resourceConfigurationRead reads the details of a Directory Insights configuration
func resourceConfigurationRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("Directory Insights configuration ID not provided"))
	}

	// Get configuration via API
	tflog.Debug(ctx, fmt.Sprintf("Reading Directory Insights configuration with ID: %s", id))
	resp, err := client.DoRequest(http.MethodGet, "/insights/directory/v1/config", nil)
	if err != nil {
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Directory Insights configuration %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading Directory Insights configuration: %v", err))
	}

	// Deserialize response
	var config Config
	if err := json.Unmarshal(resp, &config); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set values in state
	d.Set("retention_days", config.RetentionDays)
	d.Set("export_to_cloudwatch", config.ExportToCloudWatch)
	d.Set("export_to_datadog", config.ExportToDatadog)
	d.Set("datadog_region", config.DatadogRegion)

	// Don't set datadog_api_key to avoid exposing sensitive credentials

	if config.OrgID != "" {
		d.Set("org_id", config.OrgID)
	}

	if config.EnabledEventTypes != nil {
		d.Set("enabled_event_types", config.EnabledEventTypes)
	}

	if config.EnabledAlertingEvents != nil {
		d.Set("enabled_alerting_events", config.EnabledAlertingEvents)
	}

	if config.NotificationEmails != nil {
		d.Set("notification_emails", config.NotificationEmails)
	}

	return diags
}

// resourceConfigurationUpdate updates an existing Directory Insights configuration
func resourceConfigurationUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("Directory Insights configuration ID not provided"))
	}

	// Check if anything changed
	if !d.HasChanges("retention_days", "enabled_event_types", "export_to_cloudwatch",
		"export_to_datadog", "datadog_region", "datadog_api_key",
		"enabled_alerting_events", "notification_emails", "org_id") {
		return resourceConfigurationRead(ctx, d, meta)
	}

	// Build updated configuration
	config := &Config{
		ID:                 id,
		RetentionDays:      d.Get("retention_days").(int),
		ExportToCloudWatch: d.Get("export_to_cloudwatch").(bool),
		ExportToDatadog:    d.Get("export_to_datadog").(bool),
	}

	// Optional fields
	if v, ok := d.GetOk("org_id"); ok {
		config.OrgID = v.(string)
	}

	if v, ok := d.GetOk("datadog_region"); ok {
		config.DatadogRegion = v.(string)
	}

	// Only include API key if it's changed
	if d.HasChange("datadog_api_key") {
		if v, ok := d.GetOk("datadog_api_key"); ok {
			config.DatadogAPIKey = v.(string)
		}
	}

	// Process lists
	if v, ok := d.GetOk("enabled_event_types"); ok {
		eventTypes := v.([]interface{})
		config.EnabledEventTypes = make([]string, len(eventTypes))
		for i, eventType := range eventTypes {
			config.EnabledEventTypes[i] = eventType.(string)
		}
	}

	if v, ok := d.GetOk("enabled_alerting_events"); ok {
		alertingEvents := v.([]interface{})
		config.EnabledAlertingEvents = make([]string, len(alertingEvents))
		for i, event := range alertingEvents {
			config.EnabledAlertingEvents[i] = event.(string)
		}
	}

	if v, ok := d.GetOk("notification_emails"); ok {
		emails := v.([]interface{})
		config.NotificationEmails = make([]string, len(emails))
		for i, email := range emails {
			config.NotificationEmails[i] = email.(string)
		}
	}

	// Serialize to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing Directory Insights configuration: %v", err))
	}

	// Update configuration via API
	tflog.Debug(ctx, fmt.Sprintf("Updating Directory Insights configuration with ID: %s", id))
	_, err = client.DoRequest(http.MethodPut, fmt.Sprintf("/insights/directory/v1/config/%s", id), configJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating Directory Insights configuration: %v", err))
	}

	return resourceConfigurationRead(ctx, d, meta)
}

// resourceConfigurationDelete deletes a Directory Insights configuration
func resourceConfigurationDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	client, err := common.ConvertToClientInterface(meta)
	if err != nil {
		return diag.FromErr(err)
	}

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("Directory Insights configuration ID not provided"))
	}

	// Delete configuration via API
	tflog.Debug(ctx, fmt.Sprintf("Deleting Directory Insights configuration with ID: %s", id))
	_, err = client.DoRequest(http.MethodDelete, fmt.Sprintf("/insights/directory/v1/config/%s", id), nil)
	if err != nil {
		if !common.IsNotFoundError(err) {
			return diag.FromErr(fmt.Errorf("error deleting Directory Insights configuration: %v", err))
		}
		// If it's already gone, that's fine
		tflog.Warn(ctx, fmt.Sprintf("Directory Insights configuration %s was already deleted", id))
	}

	d.SetId("")
	return diags
}

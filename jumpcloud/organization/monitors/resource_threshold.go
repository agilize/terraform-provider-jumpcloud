package monitoring

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

// MonitoringThreshold represents a monitoring threshold in JumpCloud
type MonitoringThreshold struct {
	ID           string                 `json:"_id,omitempty"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	MetricType   string                 `json:"metricType"` // cpu, memory, disk, network, etc.
	ResourceType string                 `json:"resourceType"`
	Operator     string                 `json:"operator"`  // gt, lt, eq, ne, etc.
	Threshold    float64                `json:"threshold"` // threshold numeric value
	Duration     int                    `json:"duration"`  // duration in seconds
	Severity     string                 `json:"severity,omitempty"`
	Tags         []string               `json:"tags,omitempty"`
	Actions      map[string]interface{} `json:"actions,omitempty"`
	OrgID        string                 `json:"orgId,omitempty"`
	Created      string                 `json:"created,omitempty"`
	Updated      string                 `json:"updated,omitempty"`
}

// ResourceThreshold returns a schema resource for JumpCloud monitoring thresholds
func ResourceThreshold() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceMonitoringThresholdCreate,
		ReadContext:   resourceMonitoringThresholdRead,
		UpdateContext: resourceMonitoringThresholdUpdate,
		DeleteContext: resourceMonitoringThresholdDelete,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the monitoring threshold",
			},
			"description": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Description of the monitoring threshold",
			},
			"metric_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"cpu", "memory", "disk", "network", "application", "process",
					"login", "security", "system_uptime", "agent", "services",
				}, false),
				Description: "Type of monitored metric (cpu, memory, disk, etc.)",
			},
			"resource_type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"system", "user", "group", "application", "directory", "policy",
					"organization", "device", "service",
				}, false),
				Description: "Type of monitored resource",
			},
			"operator": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"gt", "lt", "eq", "ne", "ge", "le",
				}, false),
				Description: "Comparison operator (gt=greater than, lt=less than, eq=equal, ne=not equal, etc.)",
			},
			"threshold": {
				Type:        schema.TypeFloat,
				Required:    true,
				Description: "Numeric value of the threshold",
			},
			"duration": {
				Type:        schema.TypeInt,
				Required:    true,
				Description: "Duration in seconds to consider the threshold reached",
			},
			"severity": {
				Type:         schema.TypeString,
				Optional:     true,
				Default:      "medium",
				ValidateFunc: validation.StringInSlice([]string{"critical", "high", "medium", "low", "info"}, false),
				Description:  "Severity of the alert when the threshold is reached",
			},
			"tags": {
				Type:        schema.TypeSet,
				Optional:    true,
				Description: "Tags associated with the threshold",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
			"actions": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Actions to be executed when the threshold is reached, in JSON format",
				ValidateFunc: func(val interface{}, key string) (warns []string, errs []error) {
					jsonStr := val.(string)
					var js map[string]interface{}
					if err := json.Unmarshal([]byte(jsonStr), &js); err != nil {
						errs = append(errs, fmt.Errorf("%q: Invalid JSON: %s", key, err))
					}
					return
				},
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Creation date of the threshold",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Date of the last update to the threshold",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceMonitoringThresholdCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	// Process actions (JSON string to map), if provided
	var actions map[string]interface{}
	if v, ok := d.GetOk("actions"); ok {
		if err := json.Unmarshal([]byte(v.(string)), &actions); err != nil {
			return diag.FromErr(fmt.Errorf("error deserializing actions: %v", err))
		}
	}

	// Build monitoring threshold
	threshold := &MonitoringThreshold{
		Name:         d.Get("name").(string),
		MetricType:   d.Get("metric_type").(string),
		ResourceType: d.Get("resource_type").(string),
		Operator:     d.Get("operator").(string),
		Threshold:    d.Get("threshold").(float64),
		Duration:     d.Get("duration").(int),
		Severity:     d.Get("severity").(string),
		Actions:      actions,
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		threshold.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		threshold.OrgID = v.(string)
	}

	// Process tags
	if v, ok := d.GetOk("tags"); ok {
		tagsSet := v.(*schema.Set).List()
		tags := make([]string, len(tagsSet))
		for i, t := range tagsSet {
			tags[i] = t.(string)
		}
		threshold.Tags = tags
	}

	// Serialize to JSON
	thresholdJSON, err := json.Marshal(threshold)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing monitoring threshold: %v", err))
	}

	// Create monitoring threshold via API
	tflog.Debug(ctx, "Creating monitoring threshold")
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/monitoring-thresholds", thresholdJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating monitoring threshold: %v", err))
	}

	// Deserialize response
	var createdThreshold MonitoringThreshold
	if err := json.Unmarshal(resp, &createdThreshold); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	if createdThreshold.ID == "" {
		return diag.FromErr(fmt.Errorf("monitoring threshold created without ID"))
	}

	d.SetId(createdThreshold.ID)
	return resourceMonitoringThresholdRead(ctx, d, meta)
}

func resourceMonitoringThresholdRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("monitoring threshold ID not provided"))
	}

	// Fetch monitoring threshold via API
	tflog.Debug(ctx, fmt.Sprintf("Reading monitoring threshold with ID: %s", id))
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/monitoring-thresholds/%s", id), nil)
	if err != nil {
		if err.Error() == "404 Not Found" {
			tflog.Warn(ctx, fmt.Sprintf("Monitoring threshold %s not found, removing from state", id))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading monitoring threshold: %v", err))
	}

	// Deserialize response
	var threshold MonitoringThreshold
	if err := json.Unmarshal(resp, &threshold); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Set values in state
	d.Set("name", threshold.Name)
	d.Set("description", threshold.Description)
	d.Set("metric_type", threshold.MetricType)
	d.Set("resource_type", threshold.ResourceType)
	d.Set("operator", threshold.Operator)
	d.Set("threshold", threshold.Threshold)
	d.Set("duration", threshold.Duration)
	d.Set("severity", threshold.Severity)
	d.Set("created", threshold.Created)
	d.Set("updated", threshold.Updated)

	if threshold.OrgID != "" {
		d.Set("org_id", threshold.OrgID)
	}

	if len(threshold.Tags) > 0 {
		d.Set("tags", threshold.Tags)
	}

	// Convert actions map to JSON string
	if threshold.Actions != nil && len(threshold.Actions) > 0 {
		actionsJSON, err := json.Marshal(threshold.Actions)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error serializing actions to JSON: %v", err))
		}
		d.Set("actions", string(actionsJSON))
	}

	return diags
}

func resourceMonitoringThresholdUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("monitoring threshold ID not provided"))
	}

	// Process actions (JSON string to map), if provided
	var actions map[string]interface{}
	if v, ok := d.GetOk("actions"); ok {
		if err := json.Unmarshal([]byte(v.(string)), &actions); err != nil {
			return diag.FromErr(fmt.Errorf("error deserializing actions: %v", err))
		}
	}

	// Build monitoring threshold with current values
	threshold := &MonitoringThreshold{
		ID:           id,
		Name:         d.Get("name").(string),
		MetricType:   d.Get("metric_type").(string),
		ResourceType: d.Get("resource_type").(string),
		Operator:     d.Get("operator").(string),
		Threshold:    d.Get("threshold").(float64),
		Duration:     d.Get("duration").(int),
		Severity:     d.Get("severity").(string),
		Actions:      actions,
	}

	// Optional fields
	if v, ok := d.GetOk("description"); ok {
		threshold.Description = v.(string)
	}

	if v, ok := d.GetOk("org_id"); ok {
		threshold.OrgID = v.(string)
	}

	// Process tags
	if v, ok := d.GetOk("tags"); ok {
		tagsSet := v.(*schema.Set).List()
		tags := make([]string, len(tagsSet))
		for i, t := range tagsSet {
			tags[i] = t.(string)
		}
		threshold.Tags = tags
	}

	// Serialize to JSON
	thresholdJSON, err := json.Marshal(threshold)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing monitoring threshold: %v", err))
	}

	// Update monitoring threshold via API
	tflog.Debug(ctx, fmt.Sprintf("Updating monitoring threshold: %s", id))
	resp, err := c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/monitoring-thresholds/%s", id), thresholdJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating monitoring threshold: %v", err))
	}

	// Deserialize response
	var updatedThreshold MonitoringThreshold
	if err := json.Unmarshal(resp, &updatedThreshold); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	return resourceMonitoringThresholdRead(ctx, d, meta)
}

func resourceMonitoringThresholdDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	id := d.Id()
	if id == "" {
		return diag.FromErr(fmt.Errorf("monitoring threshold ID not provided"))
	}

	// Delete monitoring threshold via API
	tflog.Debug(ctx, fmt.Sprintf("Deleting monitoring threshold: %s", id))
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/monitoring-thresholds/%s", id), nil)
	if err != nil {
		if err.Error() == "404 Not Found" {
			tflog.Warn(ctx, fmt.Sprintf("Monitoring threshold %s not found, considering deleted", id))
		} else {
			return diag.FromErr(fmt.Errorf("error deleting monitoring threshold: %v", err))
		}
	}

	d.SetId("")
	return diags
}

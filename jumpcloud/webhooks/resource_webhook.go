package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/customdiff"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// Webhook represents a webhook structure in JumpCloud
type Webhook struct {
	ID          string    `json:"_id,omitempty"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Secret      string    `json:"secret,omitempty"`
	Enabled     bool      `json:"enabled"`
	EventTypes  []string  `json:"eventTypes,omitempty"`
	Description string    `json:"description,omitempty"`
	Created     time.Time `json:"created,omitempty"`
	Updated     time.Time `json:"updated,omitempty"`
}

// ResourceWebhook returns the resource to manage webhooks in JumpCloud
func ResourceWebhook() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWebhookCreate,
		ReadContext:   resourceWebhookRead,
		UpdateContext: resourceWebhookUpdate,
		DeleteContext: resourceWebhookDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "Name of the webhook",
				ValidateFunc: validation.StringLenBetween(1, 128),
			},
			"url": {
				Type:         schema.TypeString,
				Required:     true,
				Description:  "URL where webhook events will be sent",
				ValidateFunc: validation.IsURLWithHTTPorHTTPS,
			},
			"secret": {
				Type:         schema.TypeString,
				Optional:     true,
				Sensitive:    true,
				Description:  "Secret used to sign webhook payloads",
				ValidateFunc: validation.StringLenBetween(1, 64),
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the webhook is enabled",
			},
			"event_types": {
				Type:        schema.TypeList,
				Optional:    true,
				Description: "Event types that trigger this webhook",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Description of the webhook",
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the webhook was created",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the webhook was last updated",
			},
		},
		CustomizeDiff: customdiff.Sequence(
			customdiff.ValidateChange("event_types", func(ctx context.Context, old, new, meta interface{}) error {
				if new == nil {
					return nil
				}
				eventTypes := []string{}
				for _, v := range new.([]interface{}) {
					eventTypes = append(eventTypes, v.(string))
				}
				return ValidateEventTypes(eventTypes)
			}),
		),
	}
}

// ValidateEventTypes checks if the event types are valid
func ValidateEventTypes(eventTypes []string) error {
	validEventTypes := map[string]bool{
		"user.created":                    true,
		"user.updated":                    true,
		"user.deleted":                    true,
		"user_group.created":              true,
		"user_group.updated":              true,
		"user_group.deleted":              true,
		"system.created":                  true,
		"system.updated":                  true,
		"system.deleted":                  true,
		"system_group.created":            true,
		"system_group.updated":            true,
		"system_group.deleted":            true,
		"application.created":             true,
		"application.updated":             true,
		"application.deleted":             true,
		"policy.created":                  true,
		"policy.updated":                  true,
		"policy.deleted":                  true,
		"user_authentication.succeeded":   true,
		"user_authentication.failed":      true,
		"system_authentication.succeeded": true,
		"system_authentication.failed":    true,
		"directory.created":               true,
		"directory.updated":               true,
		"directory.deleted":               true,
		"radius_server.created":           true,
		"radius_server.updated":           true,
		"radius_server.deleted":           true,
	}

	for _, eventType := range eventTypes {
		if !validEventTypes[eventType] {
			return fmt.Errorf("invalid event type: %s", eventType)
		}
	}

	return nil
}

func resourceWebhookCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Build webhook from resource data
	webhook := &Webhook{
		Name:        d.Get("name").(string),
		URL:         d.Get("url").(string),
		Enabled:     d.Get("enabled").(bool),
		Description: d.Get("description").(string),
	}

	// Handle secret if provided
	if v, ok := d.GetOk("secret"); ok {
		webhook.Secret = v.(string)
	}

	// Handle event_types if provided
	if v, ok := d.GetOk("event_types"); ok {
		eventTypes := []string{}
		for _, item := range v.([]interface{}) {
			eventTypes = append(eventTypes, item.(string))
		}
		webhook.EventTypes = eventTypes
	}

	// Convert to JSON
	webhookJSON, err := json.Marshal(webhook)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing webhook: %v", err))
	}

	// Create webhook via API
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/webhooks", webhookJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating webhook: %v", err))
	}

	// Parse response
	var newWebhook Webhook
	if err := json.Unmarshal(resp, &newWebhook); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing webhook response: %v", err))
	}

	// Set ID in state
	d.SetId(newWebhook.ID)

	// Read the webhook to set all computed fields
	return resourceWebhookRead(ctx, d, meta)
}

func resourceWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	webhookID := d.Id()

	// Get webhook via API
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/webhooks/%s", webhookID), nil)
	if err != nil {
		// Handle 404 specifically
		if err.Error() == "status code 404" {
			tflog.Warn(ctx, fmt.Sprintf("Webhook %s not found, removing from state", webhookID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading webhook %s: %v", webhookID, err))
	}

	// Decode response
	var webhook Webhook
	if err := json.Unmarshal(resp, &webhook); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing webhook response: %v", err))
	}

	// Set resource data
	d.Set("name", webhook.Name)
	d.Set("url", webhook.URL)
	d.Set("enabled", webhook.Enabled)
	d.Set("description", webhook.Description)

	// Set event_types
	if webhook.EventTypes != nil {
		d.Set("event_types", webhook.EventTypes)
	}

	// Set timestamps
	if !webhook.Created.IsZero() {
		d.Set("created", webhook.Created.Format(time.RFC3339))
	}

	if !webhook.Updated.IsZero() {
		d.Set("updated", webhook.Updated.Format(time.RFC3339))
	}

	return diags
}

func resourceWebhookUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	webhookID := d.Id()

	// Build webhook from resource data
	webhook := &Webhook{
		Name:        d.Get("name").(string),
		URL:         d.Get("url").(string),
		Enabled:     d.Get("enabled").(bool),
		Description: d.Get("description").(string),
	}

	// Handle secret if provided or changed
	if d.HasChange("secret") {
		if v, ok := d.GetOk("secret"); ok {
			webhook.Secret = v.(string)
		}
	}

	// Handle event_types if provided
	if v, ok := d.GetOk("event_types"); ok {
		eventTypes := []string{}
		for _, item := range v.([]interface{}) {
			eventTypes = append(eventTypes, item.(string))
		}
		webhook.EventTypes = eventTypes
	}

	// Convert to JSON
	webhookJSON, err := json.Marshal(webhook)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing webhook: %v", err))
	}

	// Update webhook via API
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/webhooks/%s", webhookID), webhookJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating webhook %s: %v", webhookID, err))
	}

	return resourceWebhookRead(ctx, d, meta)
}

func resourceWebhookDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	webhookID := d.Id()

	// Delete webhook via API
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/webhooks/%s", webhookID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting webhook %s: %v", webhookID, err))
	}

	// Set ID to empty to signify resource has been removed
	d.SetId("")

	return nil
}

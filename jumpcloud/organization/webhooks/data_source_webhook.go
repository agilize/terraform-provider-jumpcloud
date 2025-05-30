package webhooks

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// Webhook represents a JumpCloud webhook
type Webhook struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	URL         string    `json:"url"`
	Enabled     bool      `json:"enabled"`
	Description string    `json:"description,omitempty"`
	EventTypes  []string  `json:"eventTypes,omitempty"`
	Created     time.Time `json:"created,omitempty"`
	Updated     time.Time `json:"updated,omitempty"`
}

// DataSourceWebhook returns the data source to get information about an existing webhook
func DataSourceWebhook() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceWebhookRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ConflictsWith: []string{"name"},
				Description:   "ID of the webhook in JumpCloud",
			},
			"name": {
				Type:          schema.TypeString,
				Optional:      true,
				ConflictsWith: []string{"id"},
				Description:   "Name of the webhook",
			},
			"url": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "URL where webhook events are sent",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "Whether the webhook is enabled",
			},
			"event_types": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "Event types that trigger this webhook",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"description": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Description of the webhook",
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
	}
}

func dataSourceWebhookRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Get by ID if specified
	if id, ok := d.GetOk("id"); ok {
		webhookID := id.(string)
		resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/webhooks/%s", webhookID), nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error getting webhook by ID %s: %v", webhookID, err))
		}

		// Set ID and populate the rest of the data
		d.SetId(webhookID)
		return populateWebhookData(d, resp)
	}

	// Otherwise, get by name
	if name, ok := d.GetOk("name"); ok {
		// List all webhooks
		webhookName := name.(string)
		resp, err := c.DoRequest(http.MethodGet, "/api/v2/webhooks", nil)
		if err != nil {
			return diag.FromErr(fmt.Errorf("error listing webhooks: %v", err))
		}

		// Parse the response
		var webhooks []Webhook
		if err := json.Unmarshal(resp, &webhooks); err != nil {
			return diag.FromErr(fmt.Errorf("error parsing webhooks response: %v", err))
		}

		// Find the webhook with the matching name
		var matchingWebhook *Webhook
		for _, webhook := range webhooks {
			if webhook.Name == webhookName {
				matchingWebhook = &webhook
				break
			}
		}

		if matchingWebhook == nil {
			return diag.FromErr(fmt.Errorf("no webhook found with name: %s", webhookName))
		}

		// Set ID
		d.SetId(matchingWebhook.ID)

		// Set other attributes
		if err := d.Set("name", matchingWebhook.Name); err != nil {
			return diag.FromErr(fmt.Errorf("error setting name: %v", err))
		}

		if err := d.Set("url", matchingWebhook.URL); err != nil {
			return diag.FromErr(fmt.Errorf("error setting url: %v", err))
		}

		if err := d.Set("enabled", matchingWebhook.Enabled); err != nil {
			return diag.FromErr(fmt.Errorf("error setting enabled: %v", err))
		}

		if err := d.Set("description", matchingWebhook.Description); err != nil {
			return diag.FromErr(fmt.Errorf("error setting description: %v", err))
		}

		if matchingWebhook.EventTypes != nil {
			if err := d.Set("event_types", matchingWebhook.EventTypes); err != nil {
				return diag.FromErr(fmt.Errorf("error setting event_types: %v", err))
			}
		}

		// Set timestamps
		if !matchingWebhook.Created.IsZero() {
			if err := d.Set("created", matchingWebhook.Created.Format(time.RFC3339)); err != nil {
				return diag.FromErr(fmt.Errorf("error setting created: %v", err))
			}
		}

		if !matchingWebhook.Updated.IsZero() {
			if err := d.Set("updated", matchingWebhook.Updated.Format(time.RFC3339)); err != nil {
				return diag.FromErr(fmt.Errorf("error setting updated: %v", err))
			}
		}

		return diags
	}

	return diag.FromErr(fmt.Errorf("either id or name must be specified"))
}

func populateWebhookData(d *schema.ResourceData, response []byte) diag.Diagnostics {
	var diags diag.Diagnostics
	var webhook Webhook

	// Parse the webhook data
	if err := json.Unmarshal(response, &webhook); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing webhook response: %v", err))
	}

	// Set attributes
	if err := d.Set("name", webhook.Name); err != nil {
		return diag.FromErr(fmt.Errorf("error setting name: %v", err))
	}

	if err := d.Set("url", webhook.URL); err != nil {
		return diag.FromErr(fmt.Errorf("error setting url: %v", err))
	}

	if err := d.Set("enabled", webhook.Enabled); err != nil {
		return diag.FromErr(fmt.Errorf("error setting enabled: %v", err))
	}

	if err := d.Set("description", webhook.Description); err != nil {
		return diag.FromErr(fmt.Errorf("error setting description: %v", err))
	}

	if webhook.EventTypes != nil {
		if err := d.Set("event_types", webhook.EventTypes); err != nil {
			return diag.FromErr(fmt.Errorf("error setting event_types: %v", err))
		}
	}

	// Set timestamps
	if !webhook.Created.IsZero() {
		if err := d.Set("created", webhook.Created.Format(time.RFC3339)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting created: %v", err))
		}
	}

	if !webhook.Updated.IsZero() {
		if err := d.Set("updated", webhook.Updated.Format(time.RFC3339)); err != nil {
			return diag.FromErr(fmt.Errorf("error setting updated: %v", err))
		}
	}

	return diags
}

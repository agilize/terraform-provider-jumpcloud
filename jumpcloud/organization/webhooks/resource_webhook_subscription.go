package webhooks

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

// ResourceWebhookSubscription returns the resource to manage webhook subscriptions
func ResourceWebhookSubscription() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceWebhookSubscriptionCreate,
		ReadContext:   resourceWebhookSubscriptionRead,
		UpdateContext: resourceWebhookSubscriptionUpdate,
		DeleteContext: resourceWebhookSubscriptionDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"webhook_id": {
				Type:         schema.TypeString,
				Required:     true,
				ForceNew:     true,
				Description:  "ID of the webhook this subscription belongs to",
				ValidateFunc: validation.StringIsNotEmpty,
			},
			"event_type": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Event type that triggers the webhook",
				ValidateFunc: validation.StringInSlice([]string{
					"user.created", "user.updated", "user.deleted",
					"user_group.created", "user_group.updated", "user_group.deleted",
					"system.created", "system.updated", "system.deleted",
					"system_group.created", "system_group.updated", "system_group.deleted",
					"application.created", "application.updated", "application.deleted",
					"policy.created", "policy.updated", "policy.deleted",
					"user_authentication.succeeded", "user_authentication.failed",
					"system_authentication.succeeded", "system_authentication.failed",
					"directory.created", "directory.updated", "directory.deleted",
					"radius_server.created", "radius_server.updated", "radius_server.deleted",
				}, false),
			},
			"description": {
				Type:         schema.TypeString,
				Optional:     true,
				Description:  "Description of the webhook subscription",
				ValidateFunc: validation.StringLenBetween(0, 1024),
			},
			"created": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the webhook subscription was created",
			},
			"updated": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Timestamp when the webhook subscription was last updated",
			},
		},
	}
}

func resourceWebhookSubscriptionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Build webhook subscription from resource data
	subscription := &common.WebhookSubscription{
		WebhookID:   d.Get("webhook_id").(string),
		EventType:   d.Get("event_type").(string),
		Description: d.Get("description").(string),
	}

	// Convert to JSON
	subscriptionJSON, err := json.Marshal(subscription)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing webhook subscription: %v", err))
	}

	// Create webhook subscription via API
	resp, err := c.DoRequest(http.MethodPost, "/api/v2/webhooksubscriptions", subscriptionJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error creating webhook subscription: %v", err))
	}

	// Parse response
	var newSubscription common.WebhookSubscription
	if err := json.Unmarshal(resp, &newSubscription); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing webhook subscription response: %v", err))
	}

	// Set ID in state
	d.SetId(newSubscription.ID)

	// Read the subscription to set all computed fields
	return resourceWebhookSubscriptionRead(ctx, d, meta)
}

func resourceWebhookSubscriptionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	subscriptionID := d.Id()

	// Get webhook subscription via API
	resp, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/webhooksubscriptions/%s", subscriptionID), nil)
	if err != nil {
		// Handle 404 specifically
		if common.IsNotFoundError(err) {
			tflog.Warn(ctx, fmt.Sprintf("Webhook subscription %s not found, removing from state", subscriptionID))
			d.SetId("")
			return diags
		}
		return diag.FromErr(fmt.Errorf("error reading webhook subscription %s: %v", subscriptionID, err))
	}

	// Decode response
	var subscription common.WebhookSubscription
	if err := json.Unmarshal(resp, &subscription); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing webhook subscription response: %v", err))
	}

	// Set resource data
	d.Set("webhook_id", subscription.WebhookID)
	d.Set("event_type", subscription.EventType)
	d.Set("description", subscription.Description)
	d.Set("created", subscription.Created)
	d.Set("updated", subscription.Updated)

	return diags
}

func resourceWebhookSubscriptionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	subscriptionID := d.Id()

	// Build webhook subscription from resource data
	subscription := &common.WebhookSubscription{
		WebhookID:   d.Get("webhook_id").(string),
		EventType:   d.Get("event_type").(string),
		Description: d.Get("description").(string),
	}

	// Convert to JSON
	subscriptionJSON, err := json.Marshal(subscription)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error serializing webhook subscription: %v", err))
	}

	// Update webhook subscription via API
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/webhooksubscriptions/%s", subscriptionID), subscriptionJSON)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error updating webhook subscription %s: %v", subscriptionID, err))
	}

	return resourceWebhookSubscriptionRead(ctx, d, meta)
}

func resourceWebhookSubscriptionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	subscriptionID := d.Id()

	// Delete webhook subscription via API
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/webhooksubscriptions/%s", subscriptionID), nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error deleting webhook subscription %s: %v", subscriptionID, err))
	}

	// Set ID to empty to signify resource has been removed
	d.SetId("")

	return nil
}

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

// Policy representa uma política do JumpCloud
type Policy struct {
	ID             string                 `json:"_id,omitempty"`
	Name           string                 `json:"name"`
	Description    string                 `json:"description,omitempty"`
	Type           string                 `json:"type"`
	Template       string                 `json:"template"`
	Configurations map[string]interface{} `json:"configField,omitempty"`
	Active         bool                   `json:"active"`
	OrganizationID string                 `json:"organizationId,omitempty"`
}

// resourcePolicy retorna um recurso para gerenciar políticas do JumpCloud
func resourcePolicy() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourcePolicyCreate,
		ReadContext:   resourcePolicyRead,
		UpdateContext: resourcePolicyUpdate,
		DeleteContext: resourcePolicyDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ValidateFunc: validation.StringInSlice([]string{
					"password_complexity",
					"samba_ad_password_sync",
					"password_expiration",
					"custom",
					"password_reused",
					"password_failed_attempts",
					"account_lockout_timeout",
					"mfa",
					"system_updates",
				}, false),
				Description: "Type of policy. Supported values are: password_complexity, samba_ad_password_sync, password_expiration, custom, password_reused, password_failed_attempts, account_lockout_timeout, mfa, system_updates",
			},
			"template": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Template name for the policy. Required for some policy types.",
			},
			"active": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				Description: "Whether the policy is active. Defaults to true.",
			},
			"created": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"configurations": {
				Type:        schema.TypeMap,
				Optional:    true,
				Description: "Configuration options for the policy. Specific to each policy type.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

// resourcePolicyCreate cria uma nova política no JumpCloud
func resourcePolicyCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	// Collect policy data from schema
	policyData := map[string]interface{}{
		"name":   d.Get("name").(string),
		"type":   d.Get("type").(string),
		"active": d.Get("active").(bool),
	}

	if description, ok := d.GetOk("description"); ok {
		policyData["description"] = description.(string)
	}

	if template, ok := d.GetOk("template"); ok {
		policyData["template"] = template.(string)
	}

	// Handle configurations as configField
	if configs, ok := d.GetOk("configurations"); ok {
		configMap := configs.(map[string]interface{})
		policyData["configField"] = configMap
	}

	// Convert to JSON
	policyJSON, err := json.Marshal(policyData)
	if err != nil {
		return diag.FromErr(err)
	}

	// Create policy request
	res, err := c.DoRequest(http.MethodPost, "/api/v2/policies", policyJSON)
	if err != nil {
		return diag.FromErr(err)
	}

	// Parse response to get policy ID
	var respData map[string]interface{}
	if err := json.Unmarshal(res, &respData); err != nil {
		return diag.FromErr(err)
	}

	// Save the policy ID
	policyID, ok := respData["_id"].(string)
	if !ok {
		return diag.FromErr(fmt.Errorf("error parsing policy ID from response"))
	}
	d.SetId(policyID)

	// Read the state to ensure it matches what we expect
	return resourcePolicyRead(ctx, d, m)
}

// resourcePolicyRead lê os detalhes de uma política do JumpCloud
func resourcePolicyRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	policyID := d.Id()

	// Get policy details
	res, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/policies/%s", policyID), nil)
	if err != nil {
		if isNotFoundError(err) {
			d.SetId("")
			return nil
		}
		return diag.FromErr(err)
	}

	// Parse response
	var policyData map[string]interface{}
	if err := json.Unmarshal(res, &policyData); err != nil {
		return diag.FromErr(err)
	}

	// Set standard fields
	if err := d.Set("name", policyData["name"]); err != nil {
		return diag.FromErr(err)
	}
	if desc, ok := policyData["description"]; ok {
		if err := d.Set("description", desc); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("type", policyData["type"]); err != nil {
		return diag.FromErr(err)
	}
	if template, ok := policyData["template"]; ok {
		if err := d.Set("template", template); err != nil {
			return diag.FromErr(err)
		}
	}
	if err := d.Set("active", policyData["active"]); err != nil {
		return diag.FromErr(err)
	}

	// Handle configurations
	if configField, ok := policyData["configField"].(map[string]interface{}); ok {
		// Convert all values to strings for Terraform schema compatibility
		stringConfigs := make(map[string]interface{})
		for k, v := range configField {
			switch val := v.(type) {
			case string:
				stringConfigs[k] = val
			case bool:
				stringConfigs[k] = fmt.Sprintf("%t", val)
			case float64:
				stringConfigs[k] = fmt.Sprintf("%g", val)
			default:
				stringConfigs[k] = fmt.Sprintf("%v", val)
			}
		}
		if err := d.Set("configurations", stringConfigs); err != nil {
			return diag.FromErr(err)
		}
	}

	// Get policy metadata to retrieve creation time
	metaRes, err := c.DoRequest(http.MethodGet, fmt.Sprintf("/api/v2/policies/%s/metadata", policyID), nil)
	if err != nil {
		log.Printf("[WARN] Failed to get policy metadata: %s", err)
	} else {
		var metaData map[string]interface{}
		if err := json.Unmarshal(metaRes, &metaData); err != nil {
			log.Printf("[WARN] Failed to parse policy metadata: %s", err)
		} else {
			if created, ok := metaData["created"]; ok {
				if err := d.Set("created", created); err != nil {
					return diag.FromErr(err)
				}
			}
		}
	}

	return nil
}

// resourcePolicyUpdate atualiza uma política existente no JumpCloud
func resourcePolicyUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	policyID := d.Id()

	// Collect updated policy data
	policyData := map[string]interface{}{
		"name":   d.Get("name").(string),
		"type":   d.Get("type").(string),
		"active": d.Get("active").(bool),
	}

	if description, ok := d.GetOk("description"); ok {
		policyData["description"] = description.(string)
	}

	if template, ok := d.GetOk("template"); ok {
		policyData["template"] = template.(string)
	}

	// Handle configurations
	if configs, ok := d.GetOk("configurations"); ok {
		configMap := configs.(map[string]interface{})
		policyData["configField"] = configMap
	}

	// Convert to JSON
	policyJSON, err := json.Marshal(policyData)
	if err != nil {
		return diag.FromErr(err)
	}

	// Update policy
	_, err = c.DoRequest(http.MethodPut, fmt.Sprintf("/api/v2/policies/%s", policyID), policyJSON)
	if err != nil {
		return diag.FromErr(err)
	}

	return resourcePolicyRead(ctx, d, m)
}

// resourcePolicyDelete exclui uma política do JumpCloud
func resourcePolicyDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	c, diags := ConvertToClientInterface(m)
	if diags != nil {
		return diags
	}

	policyID := d.Id()

	// Delete policy
	_, err := c.DoRequest(http.MethodDelete, fmt.Sprintf("/api/v2/policies/%s", policyID), nil)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")
	return nil
}

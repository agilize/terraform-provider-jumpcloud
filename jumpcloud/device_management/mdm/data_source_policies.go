package mdm

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
)

// DataSourcePolicies returns the schema for the MDM policies data source
func DataSourcePolicies() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceMDMPoliciesRead,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environment",
			},
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Field to filter on (name, platform, scopeType)",
							ValidateFunc: validation.StringInSlice([]string{
								"name", "platform", "scopeType",
							}, false),
						},
						"operator": {
							Type:         schema.TypeString,
							Required:     true,
							Description:  "Operator to use for comparison (eq, ne, contains, startswith, endswith)",
							ValidateFunc: validation.StringInSlice([]string{"eq", "ne", "contains", "startswith", "endswith"}, false),
						},
						"value": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Value to compare against",
						},
					},
				},
			},
			"sort": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Field to sort by (name, platform, created, updated)",
							ValidateFunc: validation.StringInSlice([]string{
								"name", "platform", "created", "updated",
							}, false),
						},
						"direction": {
							Type:         schema.TypeString,
							Optional:     true,
							Default:      "asc",
							Description:  "Sort direction (asc, desc)",
							ValidateFunc: validation.StringInSlice([]string{"asc", "desc"}, false),
						},
					},
				},
			},
			"policies": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Policy ID",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Policy name",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Policy description",
						},
						"platform": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Device platform (ios, android, windows, macos)",
						},
						"settings": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "JSON string of policy settings",
						},
						"scope_type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Scope type for the policy (all, group, device)",
						},
						"scope_ids": {
							Type:        schema.TypeList,
							Computed:    true,
							Description: "IDs of groups or devices in the scope",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation date of the policy",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Date of the last update to the policy",
						},
						"tags": {
							Type:        schema.TypeMap,
							Computed:    true,
							Description: "Policy tags",
							Elem: &schema.Schema{
								Type: schema.TypeString,
							},
						},
					},
				},
			},
		},
	}
}

func dataSourceMDMPoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c := meta.(interface {
		DoRequest(method, path string, body []byte) ([]byte, error)
	})

	// Build query parameters for the request
	queryParams := ""
	if orgID, ok := d.GetOk("org_id"); ok {
		queryParams += fmt.Sprintf("orgId=%s&", orgID.(string))
	}

	// Handle filters
	if filters, ok := d.GetOk("filter"); ok {
		for _, f := range filters.(*schema.Set).List() {
			filter := f.(map[string]interface{})
			field := filter["field"].(string)
			operator := filter["operator"].(string)
			value := filter["value"].(string)

			queryParams += fmt.Sprintf("filter[%s][%s]=%s&", field, operator, value)
		}
	}

	// Handle sort
	if sorts, ok := d.GetOk("sort"); ok && sorts.(*schema.Set).Len() > 0 {
		sort := sorts.(*schema.Set).List()[0].(map[string]interface{})
		field := sort["field"].(string)
		direction := sort["direction"].(string)

		queryParams += fmt.Sprintf("sort=%s:%s&", field, direction)
	}

	// Remove trailing '&' if present
	if len(queryParams) > 0 {
		queryParams = queryParams[:len(queryParams)-1]
	}

	// Make the request to the API
	url := "/api/v2/mdm/policies"
	if queryParams != "" {
		url += "?" + queryParams
	}

	tflog.Debug(ctx, fmt.Sprintf("Querying MDM policies: %s", url))
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error querying MDM policies: %v", err))
	}

	// Deserialize response
	var policies []MDMPolicy
	if err := json.Unmarshal(resp, &policies); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Format policies for output
	formattedPolicies := make([]map[string]interface{}, len(policies))
	for i, policy := range policies {
		policyMap := map[string]interface{}{
			"id":         policy.ID,
			"name":       policy.Name,
			"platform":   policy.Platform,
			"scope_type": policy.ScopeType,
		}

		// Format settings as JSON string
		if policy.Settings != nil {
			settingsStr, err := normalizeJSONString(string(policy.Settings))
			if err != nil {
				return diag.FromErr(fmt.Errorf("error normalizing settings JSON for policy %s: %v", policy.ID, err))
			}
			policyMap["settings"] = settingsStr
		}

		// Add optional fields if present
		if policy.Description != "" {
			policyMap["description"] = policy.Description
		}
		if len(policy.ScopeIDs) > 0 {
			policyMap["scope_ids"] = policy.ScopeIDs
		}
		if len(policy.Tags) > 0 {
			policyMap["tags"] = policy.Tags
		}
		if policy.Created != "" {
			policyMap["created"] = policy.Created
		}
		if policy.Updated != "" {
			policyMap["updated"] = policy.Updated
		}

		formattedPolicies[i] = policyMap
	}

	if err := d.Set("policies", formattedPolicies); err != nil {
		return diag.FromErr(fmt.Errorf("error setting policies in state: %v", err))
	}

	// Set a unique ID for this data source
	d.SetId(time.Now().Format(time.RFC3339))

	return diags
}

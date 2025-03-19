package provider

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

// SoftwareUpdatePolicyListItem represents a software update policy in the list response
type SoftwareUpdatePolicyListItem struct {
	ID          string                 `json:"_id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description,omitempty"`
	OSFamily    string                 `json:"osFamily"`
	Enabled     bool                   `json:"enabled"`
	AllPackages bool                   `json:"allPackages"`
	AutoApprove bool                   `json:"autoApprove"`
	Status      string                 `json:"status,omitempty"`
	Schedule    map[string]interface{} `json:"schedule"`
	TargetCount int                    `json:"targetCount"`
	OrgID       string                 `json:"orgId,omitempty"`
	Created     string                 `json:"created"`
	Updated     string                 `json:"updated"`
}

// SoftwareUpdatePoliciesResponse represents the API response for listing software update policies
type SoftwareUpdatePoliciesResponse struct {
	Results    []SoftwareUpdatePolicyListItem `json:"results"`
	TotalCount int                            `json:"totalCount"`
}

func dataSourceSoftwareUpdatePolicies() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSoftwareUpdatePoliciesRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Filter by policy name (partial match)",
			},
			"os_family": {
				Type:     schema.TypeString,
				Optional: true,
				ValidateFunc: validation.StringInSlice([]string{
					"windows", "macos", "linux",
				}, false),
				Description: "Filter by operating system family",
			},
			"enabled": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Filter by enabled status",
			},
			"auto_approve": {
				Type:        schema.TypeBool,
				Optional:    true,
				Description: "Filter by auto-approve setting",
			},
			"search": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Search term for policy name or description",
			},
			"sort": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "name",
				ValidateFunc: validation.StringInSlice([]string{
					"name", "osFamily", "enabled", "created", "updated",
				}, false),
				Description: "Field to sort results by",
			},
			"sort_dir": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "asc",
				ValidateFunc: validation.StringInSlice([]string{
					"asc", "desc",
				}, false),
				Description: "Sort direction (asc, desc)",
			},
			"limit": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      50,
				ValidateFunc: validation.IntBetween(1, 1000),
				Description:  "Maximum number of results to return (1-1000)",
			},
			"skip": {
				Type:         schema.TypeInt,
				Optional:     true,
				Default:      0,
				ValidateFunc: validation.IntAtLeast(0),
				Description:  "Number of results to skip (for pagination)",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID for multi-tenant environments",
			},
			"policies": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of software update policies",
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
						"os_family": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Operating system family (windows, macos, linux)",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether policy is enabled",
						},
						"all_packages": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether policy applies to all packages",
						},
						"auto_approve": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether updates are auto-approved",
						},
						"status": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Current status of the policy",
						},
						"schedule": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Schedule configuration in JSON format",
						},
						"target_count": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Number of targets for this policy",
						},
						"org_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Organization ID",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Creation date",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Last update date",
						},
					},
				},
			},
			"total": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of policies matching the filter criteria",
			},
		},
	}
}

func dataSourceSoftwareUpdatePoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := ConvertToClientInterface(meta)
	if diagErr != nil {
		return diagErr
	}

	// Build query parameters
	queryParams := constructSoftwareUpdatePoliciesQueryParams(d)

	// Build URL for request
	url := fmt.Sprintf("/api/v2/software/update-policies%s", queryParams)

	// Make API request to list policies
	tflog.Debug(ctx, "Listing software update policies")
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing software update policies: %v", err))
	}

	// Deserialize response
	var policiesResp SoftwareUpdatePoliciesResponse
	if err := json.Unmarshal(resp, &policiesResp); err != nil {
		return diag.FromErr(fmt.Errorf("error deserializing response: %v", err))
	}

	// Generate ID for data source
	d.SetId(fmt.Sprintf("software-update-policies-%d", time.Now().Unix()))

	// Map values to schema
	d.Set("total", policiesResp.TotalCount)

	// Process policies
	policies := flattenSoftwareUpdatePolicies(policiesResp.Results)
	if err := d.Set("policies", policies); err != nil {
		return diag.FromErr(fmt.Errorf("error setting policies: %v", err))
	}

	return diags
}

// Builds the query parameters string for the API request
func constructSoftwareUpdatePoliciesQueryParams(d *schema.ResourceData) string {
	params := "?"

	// Add filters
	if v, ok := d.GetOk("name"); ok {
		params += fmt.Sprintf("name=%s&", v.(string))
	}

	if v, ok := d.GetOk("os_family"); ok {
		params += fmt.Sprintf("osFamily=%s&", v.(string))
	}

	if v, ok := d.GetOk("enabled"); ok {
		params += fmt.Sprintf("enabled=%t&", v.(bool))
	}

	if v, ok := d.GetOk("auto_approve"); ok {
		params += fmt.Sprintf("autoApprove=%t&", v.(bool))
	}

	if v, ok := d.GetOk("search"); ok {
		params += fmt.Sprintf("search=%s&", v.(string))
	}

	// Add pagination and sorting
	params += fmt.Sprintf("limit=%d&", d.Get("limit").(int))
	params += fmt.Sprintf("skip=%d&", d.Get("skip").(int))
	params += fmt.Sprintf("sort=%s&", d.Get("sort").(string))
	params += fmt.Sprintf("sortDir=%s&", d.Get("sort_dir").(string))

	// Add orgId if provided
	if v, ok := d.GetOk("org_id"); ok {
		params += fmt.Sprintf("orgId=%s&", v.(string))
	}

	// Remove the last '&' if it exists
	if len(params) > 1 && params[len(params)-1] == '&' {
		params = params[:len(params)-1]
	}

	return params
}

// Converts API response to format expected by Terraform
func flattenSoftwareUpdatePolicies(policies []SoftwareUpdatePolicyListItem) []interface{} {
	if policies == nil {
		return []interface{}{}
	}

	result := make([]interface{}, len(policies))
	for i, policy := range policies {
		// Convert schedule to JSON string
		var scheduleJSON string
		if policy.Schedule != nil {
			scheduleData, err := json.Marshal(policy.Schedule)
			if err == nil {
				scheduleJSON = string(scheduleData)
			}
		}

		policyMap := map[string]interface{}{
			"id":           policy.ID,
			"name":         policy.Name,
			"description":  policy.Description,
			"os_family":    policy.OSFamily,
			"enabled":      policy.Enabled,
			"all_packages": policy.AllPackages,
			"auto_approve": policy.AutoApprove,
			"status":       policy.Status,
			"schedule":     scheduleJSON,
			"target_count": policy.TargetCount,
			"created":      policy.Created,
			"updated":      policy.Updated,
		}

		if policy.OrgID != "" {
			policyMap["org_id"] = policy.OrgID
		}

		result[i] = policyMap
	}

	return result
}

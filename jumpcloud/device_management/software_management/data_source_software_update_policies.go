package software_management

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"registry.terraform.io/agilize/jumpcloud/jumpcloud/common"
)

// SoftwareUpdatePolicyListItem represents a software update policy in the list response
type SoftwareUpdatePolicyListItem struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	OSFamily    string    `json:"osFamily"`
	Enabled     bool      `json:"enabled"`
	AllPackages bool      `json:"allPackages,omitempty"`
	AutoApprove bool      `json:"autoApprove,omitempty"`
	Created     time.Time `json:"created,omitempty"`
	Updated     time.Time `json:"updated,omitempty"`
}

// SoftwareUpdatePoliciesResponse represents the API response for listing software update policies
type SoftwareUpdatePoliciesResponse struct {
	Results    []SoftwareUpdatePolicyListItem `json:"results"`
	TotalCount int                            `json:"totalCount"`
}

// DataSourceSoftwareUpdatePolicies returns a data source for software update policies
func DataSourceSoftwareUpdatePolicies() *schema.Resource {
	return &schema.Resource{
		ReadContext: dataSourceSoftwareUpdatePoliciesRead,
		Schema: map[string]*schema.Schema{
			"filter": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter policies by name (partial match)",
						},
						"os_family": {
							Type:        schema.TypeString,
							Optional:    true,
							Description: "Filter policies by OS family (windows, mac, linux)",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Optional:    true,
							Description: "Filter policies by enabled status",
						},
					},
				},
			},
			"sort": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"field": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "Field to sort by (name, created, updated)",
						},
						"direction": {
							Type:        schema.TypeString,
							Optional:    true,
							Default:     "asc",
							Description: "Sort direction (asc, desc)",
						},
					},
				},
			},
			"limit": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     100,
				Description: "Maximum number of policies to return",
			},
			"skip": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     0,
				Description: "Number of policies to skip for pagination",
			},
			"org_id": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Organization ID to use for API requests",
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
							Description: "ID of the software update policy",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Name of the software update policy",
						},
						"description": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Description of the software update policy",
						},
						"os_family": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "OS family this policy applies to (windows, mac, linux)",
						},
						"enabled": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether the policy is enabled",
						},
						"all_packages": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether all packages are included in the policy",
						},
						"auto_approve": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether updates are automatically approved",
						},
						"created": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp when the policy was created",
						},
						"updated": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Timestamp when the policy was last updated",
						},
					},
				},
			},
		},
	}
}

func dataSourceSoftwareUpdatePoliciesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	c, diagErr := common.GetClientFromMeta(meta)
	if diagErr != nil {
		return diagErr
	}

	// Construct query parameters
	queryParams := constructSoftwareUpdatePoliciesQueryParams(d)

	// Make the API request
	url := fmt.Sprintf("/api/v2/software/update-policies%s", queryParams)

	tflog.Debug(ctx, "Listing software update policies")
	resp, err := c.DoRequest(http.MethodGet, url, nil)
	if err != nil {
		return diag.FromErr(fmt.Errorf("error listing software update policies: %v", err))
	}

	// Parse the response
	var policiesResp SoftwareUpdatePoliciesResponse
	if err := json.Unmarshal(resp, &policiesResp); err != nil {
		return diag.FromErr(fmt.Errorf("error parsing software update policies response: %v", err))
	}

	// Set the ID to a timestamp
	d.SetId(fmt.Sprintf("software-update-policies-%d", time.Now().Unix()))

	// Set the policies in the state
	policies := flattenSoftwareUpdatePolicies(policiesResp.Results)
	if err := d.Set("policies", policies); err != nil {
		return diag.FromErr(fmt.Errorf("error setting policies: %v", err))
	}

	return diags
}

func constructSoftwareUpdatePoliciesQueryParams(d *schema.ResourceData) string {
	query := url.Values{}

	if v, ok := d.GetOk("limit"); ok {
		query.Add("limit", strconv.Itoa(v.(int)))
	}

	if v, ok := d.GetOk("skip"); ok {
		query.Add("skip", strconv.Itoa(v.(int)))
	}

	if v, ok := d.GetOk("filter"); ok && len(v.([]interface{})) > 0 {
		filter := v.([]interface{})[0].(map[string]interface{})

		if name, ok := filter["name"]; ok && name.(string) != "" {
			query.Add("name", name.(string))
		}

		if osFamily, ok := filter["os_family"]; ok && osFamily.(string) != "" {
			query.Add("osFamily", osFamily.(string))
		}

		if enabled, ok := filter["enabled"]; ok {
			query.Add("enabled", strconv.FormatBool(enabled.(bool)))
		}
	}

	if v, ok := d.GetOk("sort"); ok && len(v.([]interface{})) > 0 {
		sort := v.([]interface{})[0].(map[string]interface{})

		field := sort["field"].(string)
		direction := sort["direction"].(string)

		// API expects camelCase for field names
		switch field {
		case "os_family":
			field = "osFamily"
		case "all_packages":
			field = "allPackages"
		case "auto_approve":
			field = "autoApprove"
		}

		query.Add("sort", fmt.Sprintf("%s:%s", field, direction))
	}

	if len(query) == 0 {
		return ""
	}

	return "?" + query.Encode()
}

func flattenSoftwareUpdatePolicies(policies []SoftwareUpdatePolicyListItem) []interface{} {
	var result []interface{}

	for _, policy := range policies {
		p := map[string]interface{}{
			"id":           policy.ID,
			"name":         policy.Name,
			"description":  policy.Description,
			"os_family":    policy.OSFamily,
			"enabled":      policy.Enabled,
			"all_packages": policy.AllPackages,
			"auto_approve": policy.AutoApprove,
		}

		if !policy.Created.IsZero() {
			p["created"] = policy.Created.Format(time.RFC3339)
		}

		if !policy.Updated.IsZero() {
			p["updated"] = policy.Updated.Format(time.RFC3339)
		}

		result = append(result, p)
	}

	return result
}
